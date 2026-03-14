package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/Work-Fort/Scope/pkg/frontend"
)

// TrackedService is the live state of a discovered service.
type TrackedService struct {
	URL       string   `json:"-"`
	Name      string   `json:"name"`
	Label     string   `json:"label"`
	Route     string   `json:"route"`
	Enabled   bool     `json:"enabled"`
	UI        bool     `json:"ui"`
	Connected bool     `json:"connected"`
	WSPaths   []string `json:"-"`

	wsRefCount int32
	hasWS      bool
}

// Conflict records a service that was excluded due to a collision.
type Conflict struct {
	URL    string `json:"url"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// ServiceTracker maintains live state for all services via health probing
// and WebSocket connection tracking.
type ServiceTracker struct {
	urls   []string
	client *http.Client

	mu        sync.RWMutex
	services  []TrackedService
	conflicts []Conflict
	byName    map[string]int // name -> index into services
	byRoute   map[string]int // route -> index into services

	// OnServiceDiscovered is called when a new service is discovered after
	// the initial probe. Called WITHOUT holding the mutex.
	OnServiceDiscovered func(svc TrackedService)
}

// NewServiceTracker creates a tracker for the given service URLs.
func NewServiceTracker(urls []string) *ServiceTracker {
	return &ServiceTracker{
		urls: urls,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		byName:  make(map[string]int),
		byRoute: make(map[string]int),
	}
}

// InitialProbe runs a synchronous probe of all services in config order.
// Sequential so conflict resolution is deterministic (first in config wins).
func (t *ServiceTracker) InitialProbe(ctx context.Context) {
	for _, u := range t.urls {
		t.probeOne(ctx, u, false)
	}
}

// StartPolling begins background health probing on the given interval.
func (t *ServiceTracker) StartPolling(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Rebuild conflicts on each cycle.
				t.mu.Lock()
				t.conflicts = nil
				t.mu.Unlock()

				for _, url := range t.urls {
					t.probeOne(ctx, url, true)
				}
			}
		}
	}()
}

func (t *ServiceTracker) probeOne(ctx context.Context, serviceURL string, notify bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serviceURL+"/ui/health", nil)
	if err != nil {
		return
	}

	resp, err := t.client.Do(req)
	if err != nil {
		// Service unreachable — mark disconnected if already known (non-WS only).
		t.mu.Lock()
		for i := range t.services {
			if t.services[i].URL == serviceURL && !t.services[i].hasWS {
				t.services[i].Connected = false
			}
		}
		t.mu.Unlock()
		return
	}
	defer resp.Body.Close()

	// Both 200 and 503 include the manifest.
	var health struct {
		Status string `json:"status"`
		frontend.Manifest
	}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil || health.Name == "" {
		return
	}

	hasUI := resp.StatusCode == http.StatusOK
	hasWS := len(health.WSPaths) > 0

	t.mu.Lock()

	// Check for conflicts.
	if idx, exists := t.byName[health.Name]; exists {
		if t.services[idx].URL != serviceURL {
			t.conflicts = append(t.conflicts, Conflict{
				URL:    serviceURL,
				Name:   health.Name,
				Reason: "duplicate name (already registered from " + t.services[idx].URL + ")",
			})
			t.mu.Unlock()
			return
		}
	}
	if idx, exists := t.byRoute[health.Route]; exists {
		if t.services[idx].URL != serviceURL {
			t.conflicts = append(t.conflicts, Conflict{
				URL:    serviceURL,
				Name:   health.Name,
				Reason: "duplicate route " + health.Route + " (already registered from " + t.services[idx].URL + ")",
			})
			t.mu.Unlock()
			return
		}
	}

	// Update existing service.
	if idx, exists := t.byName[health.Name]; exists {
		t.services[idx].Label = health.Label
		t.services[idx].Route = health.Route
		t.services[idx].UI = hasUI
		t.services[idx].hasWS = hasWS
		t.services[idx].WSPaths = health.WSPaths
		if !hasWS {
			t.services[idx].Connected = true
		}
		t.mu.Unlock()
		return
	}

	// New service.
	svc := TrackedService{
		URL:       serviceURL,
		Name:      health.Name,
		Label:     health.Label,
		Route:     health.Route,
		Enabled:   true,
		UI:        hasUI,
		Connected: !hasWS,
		WSPaths:   health.WSPaths,
		hasWS:     hasWS,
	}

	idx := len(t.services)
	t.services = append(t.services, svc)
	t.byName[health.Name] = idx
	t.byRoute[health.Route] = idx

	// Release lock BEFORE calling the callback to avoid deadlock.
	t.mu.Unlock()

	if notify && t.OnServiceDiscovered != nil {
		t.OnServiceDiscovered(svc)
	}
}

// OnConnect increments the WS connection ref count for a service.
func (t *ServiceTracker) OnConnect(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if idx, ok := t.byName[name]; ok {
		t.services[idx].wsRefCount++
		t.services[idx].Connected = t.services[idx].wsRefCount > 0
	}
}

// OnDisconnect decrements the WS connection ref count for a service.
func (t *ServiceTracker) OnDisconnect(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if idx, ok := t.byName[name]; ok {
		t.services[idx].wsRefCount--
		if t.services[idx].wsRefCount < 0 {
			t.services[idx].wsRefCount = 0
		}
		t.services[idx].Connected = t.services[idx].wsRefCount > 0
	}
}

// Services returns a snapshot of all discovered services.
func (t *ServiceTracker) Services() []TrackedService {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]TrackedService, len(t.services))
	copy(out, t.services)
	return out
}

// Conflicts returns a snapshot of all detected conflicts.
func (t *ServiceTracker) Conflicts() []Conflict {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]Conflict, len(t.conflicts))
	copy(out, t.conflicts)
	return out
}

// ServiceByName returns a discovered service by name, or false if not found.
func (t *ServiceTracker) ServiceByName(name string) (TrackedService, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if idx, ok := t.byName[name]; ok {
		return t.services[idx], true
	}
	return TrackedService{}, false
}
