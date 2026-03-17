package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/Work-Fort/Scope/internal/domain"
)

var errFortNotFound = errors.New("fort not found")

type FortInstance struct {
	fort    domain.Fort
	tracker *ServiceTracker
	tc      *TokenConverter
	handler http.Handler
	lastReq atomic.Int64

	mu     sync.Mutex
	cancel context.CancelFunc
}

func (fi *FortInstance) isIdle() bool {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	return fi.cancel == nil
}

func (fi *FortInstance) stopPolling() {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	if fi.cancel != nil {
		fi.cancel()
		fi.cancel = nil
	}
}

type FortRouter struct {
	registry   domain.FortRegistry
	spaHandler http.Handler
	instances  sync.Map
	initGroup  singleflight.Group
	mux        *http.ServeMux
}

func NewFortRouter(registry domain.FortRegistry, spaHandler http.Handler) *FortRouter {
	fr := &FortRouter{
		registry:   registry,
		spaHandler: spaHandler,
	}
	fr.mux = http.NewServeMux()
	fr.mux.HandleFunc("GET /api/forts", fr.listFortsHandler)
	fr.mux.HandleFunc("/forts/{fort}/{rest...}", fr.fortDispatch)
	if spaHandler != nil {
		fr.mux.Handle("/", spaHandler)
	}
	return fr
}

func (fr *FortRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fr.mux.ServeHTTP(w, r)
}

func (fr *FortRouter) listFortsHandler(w http.ResponseWriter, r *http.Request) {
	forts := fr.registry.Forts()
	type fortInfo struct {
		Name    string `json:"name"`
		Local   bool   `json:"local"`
		Gateway string `json:"gateway,omitempty"`
	}
	out := make([]fortInfo, len(forts))
	for i, f := range forts {
		out[i] = fortInfo{Name: f.Name, Local: f.Local, Gateway: f.Gateway}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (fr *FortRouter) fortDispatch(w http.ResponseWriter, r *http.Request) {
	fortName := r.PathValue("fort")
	if !domain.ValidFortName(fortName) {
		http.NotFound(w, r)
		return
	}

	inst, err := fr.getInstance(r.Context(), fortName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	inst.lastReq.Store(time.Now().Unix())

	prefix := "/forts/" + fortName
	r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	// Non-API paths are client-side routes — serve the SPA so the
	// browser's router can handle them (e.g., /chat, /nexus).
	if fr.spaHandler != nil && !strings.HasPrefix(r.URL.Path, "/api/") {
		fr.spaHandler.ServeHTTP(w, r)
		return
	}

	inst.handler.ServeHTTP(w, r)
}

// getInstance returns an active FortInstance, reinitializing it if idle.
// The idle check and singleflight.Do are not atomic — a concurrent cleanup
// could set cancel=nil between Load and isIdle, causing a redundant (but
// harmless) reinit that the singleflight deduplicates.
func (fr *FortRouter) getInstance(ctx context.Context, name string) (*FortInstance, error) {
	if v, ok := fr.instances.Load(name); ok {
		inst := v.(*FortInstance)
		if inst.isIdle() {
			return fr.initInstance(ctx, name)
		}
		return inst, nil
	}
	return fr.initInstance(ctx, name)
}

func (fr *FortRouter) initInstance(ctx context.Context, name string) (*FortInstance, error) {
	v, err, _ := fr.initGroup.Do(name, func() (any, error) {
		fort, ok := fr.registry.Fort(name)
		if !ok {
			return nil, errFortNotFound
		}

		urls := make([]string, len(fort.Services))
		for i, s := range fort.Services {
			urls[i] = s.URL
		}

		tracker := NewServiceTracker(urls)
		tracker.InitialProbe(ctx)

		var tc *TokenConverter
		if authSvc, ok := tracker.ServiceByName("auth"); ok {
			tc = NewTokenConverter(authSvc.URL)
		}

		handler := NewHandler(fort, tracker, tc, nil)

		pollCtx, cancel := context.WithCancel(context.Background())
		tracker.StartPolling(pollCtx, 10*time.Second)

		inst := &FortInstance{
			fort:    fort,
			tracker: tracker,
			tc:      tc,
			handler: handler,
			cancel:  cancel,
		}
		inst.lastReq.Store(time.Now().Unix())
		fr.instances.Store(name, inst)
		return inst, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*FortInstance), nil
}

func (fr *FortRouter) StartIdleCleanup(ctx context.Context, maxIdle time.Duration) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				// Stop all active instances on shutdown.
				fr.instances.Range(func(_, value any) bool {
					value.(*FortInstance).stopPolling()
					return true
				})
				return
			case <-ticker.C:
				now := time.Now().Unix()
				fr.instances.Range(func(key, value any) bool {
					inst := value.(*FortInstance)
					if !inst.isIdle() && now-inst.lastReq.Load() > int64(maxIdle.Seconds()) {
						inst.stopPolling()
					}
					return true
				})
			}
		}
	}()
}
