# Fort-Scoped Routing Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the single-fort active-state model with fort-scoped URL paths, per-fort auth isolation, and lazy initialization supporting 24+ forts.

**Architecture:** `FortRouter` is the new top-level HTTP handler. It lazily creates `FortInstance`s (tracker + token converter + handler) on first request per fort, using `singleflight` for concurrency safety. Cookie paths are scoped to `/forts/{fort}/` via `Set-Cookie` rewriting on the auth proxy. The SolidJS frontend routes through `/:fort/:service/*rest` with a fort picker at `/`.

**Tech Stack:** Go 1.25 (net/http, singleflight, sync/atomic), SolidJS, @solidjs/router, UnoCSS

**Spec:** `docs/2026-03-14-fort-scoped-routing-design.md`

---

## File Map

### Go — New files

| File | Responsibility |
|------|---------------|
| `internal/infra/httpapi/fort_router.go` | `FortRouter`, `FortInstance`, lazy init, idle cleanup, `/api/forts` endpoint, fort name validation |
| `internal/infra/httpapi/fort_router_test.go` | Multi-fort routing, lazy init, cookie scoping, idle cleanup tests |

### Go — Modified files

| File | Change |
|------|--------|
| `internal/domain/web.go` | Add `Fort(name) (Fort, bool)` to `FortRegistry`. Remove `Active()`, `SetActive()`. Add `ValidFortName()`. |
| `internal/infra/fortconfig/registry.go` | Implement `Fort(name)`. Remove `Active()`/`SetActive()`. |
| `internal/infra/httpapi/handler.go` | Accept `fortName` param. Pass fort name to `bffMiddleware` and `writeAuthError` for cookie path scoping. Use `NewAuthProxy` for auth routes. SPA fallback is disabled by passing `nil` for `spaFS` (guard already exists). |
| `internal/infra/httpapi/handler_test.go` | Update `NewHandler` calls with fort name. |
| `internal/infra/httpapi/proxy.go` | Add `NewAuthProxy` with `ModifyResponse` for `Set-Cookie` path rewriting. |
| `cmd/web/web.go` | Create `FortRouter` instead of single fort. Remove single tracker/TC. Remove `topMux`. |

### Frontend — New files

| File | Responsibility |
|------|---------------|
| `web/shell/src/components/fort-picker.tsx` | Fort selection landing page at `/` |

### Frontend — Modified files

| File | Change |
|------|--------|
| `web/shell/src/lib/api.ts` | Add `fetchForts()`, `FortInfo` type. All functions take `fort` param. |
| `web/shell/src/stores/services.ts` | Fort-scoped polling. Start/stop tied to URL param. |
| `web/shell/src/app.tsx` | Router: `/` → FortPicker, `/forts/:fort/:service/*rest` → ServicePage. |
| `web/shell/src/components/nav-bar.tsx` | Show fort name. Routes include fort prefix. |

---

## Chunk 1: Domain & Registry

### Task 1: Update FortRegistry interface

**Files:**
- Modify: `internal/domain/web.go`

- [ ] **Step 1: Update the FortRegistry interface and add fort name validation**

Replace the current interface with fort-scoped lookups, and add `ValidFortName` to the domain package (so both `fortconfig` and `httpapi` can use it without import coupling):

```go
// FortRegistry provides access to configured forts.
type FortRegistry interface {
	Forts() []Fort
	Fort(name string) (Fort, bool)
}

var fortNameRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// ValidFortName returns true if name is a valid fort identifier.
// Lowercase alphanumeric + hyphens, no leading/trailing hyphens.
func ValidFortName(name string) bool {
	return fortNameRe.MatchString(name)
}
```

Add `"regexp"` to the import block. Remove `Active()` and `SetActive()` — no server-side active fort.

- [ ] **Step 2: Verify build**

Run: `cd /home/kazw/Work/WorkFort/scope/lead && go build ./internal/domain/...`
Expected: PASS (domain has no dependents that use Active/SetActive yet — we fix those in later tasks)

- [ ] **Step 3: Commit**

```bash
git add internal/domain/web.go
git commit -m "refactor(domain): replace Active/SetActive with Fort(name) lookup"
```

### Task 2: Update fortconfig registry

**Files:**
- Modify: `internal/infra/fortconfig/registry.go`

- [ ] **Step 1: Write the failing test**

Create test for `Fort(name)` lookup and `domain.ValidFortName` validation:

```go
// internal/infra/fortconfig/registry_test.go
package fortconfig_test

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/Work-Fort/Scope/internal/domain"
	"github.com/Work-Fort/Scope/internal/infra/fortconfig"
)

func TestRegistry_Fort(t *testing.T) {
	viper.Reset()
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services", []map[string]any{
		{"url": "http://127.0.0.1:3000"},
		{"url": "http://127.0.0.1:9600"},
	})

	reg := fortconfig.New()
	fort, ok := reg.Fort("local")
	if !ok {
		t.Fatal("expected fort 'local' to exist")
	}
	if fort.Name != "local" {
		t.Errorf("got name %q, want %q", fort.Name, "local")
	}
	if len(fort.Services) != 2 {
		t.Errorf("got %d services, want 2", len(fort.Services))
	}

	_, ok = reg.Fort("nonexistent")
	if ok {
		t.Error("expected nonexistent fort to return false")
	}
}

func TestValidFortName(t *testing.T) {
	valid := []string{"local", "acme-corp", "a", "a1", "test-fort-123"}
	invalid := []string{"-leading", "trailing-", "UPPER", "has space", "has/slash", "has..dots", ""}

	for _, name := range valid {
		if !domain.ValidFortName(name) {
			t.Errorf("expected %q to be valid", name)
		}
	}
	for _, name := range invalid {
		if domain.ValidFortName(name) {
			t.Errorf("expected %q to be invalid", name)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/kazw/Work/WorkFort/scope/lead && go test ./internal/infra/fortconfig/... -v -run TestRegistry_Fort`
Expected: FAIL (Fort method doesn't exist, ValidFortName doesn't exist)

- [ ] **Step 3: Implement Fort(name), remove Active/SetActive**

Add the `Fort(name)` method and remove `Active()`/`SetActive()`. Keep the existing `readFort` implementation (uses `viper.UnmarshalKey` with struct tags). Use `domain.ValidFortName` for validation (defined in Task 1).

```go
// internal/infra/fortconfig/registry.go
package fortconfig

import (
	"sort"

	"github.com/spf13/viper"

	"github.com/Work-Fort/Scope/internal/domain"
)

var _ domain.FortRegistry = (*Registry)(nil)

type Registry struct{}

func New() *Registry {
	return &Registry{}
}

func (r *Registry) Forts() []domain.Fort {
	fortsMap := viper.GetStringMap("forts")
	forts := make([]domain.Fort, 0, len(fortsMap))
	for name := range fortsMap {
		if domain.ValidFortName(name) {
			forts = append(forts, r.readFort(name))
		}
	}
	sort.Slice(forts, func(i, j int) bool {
		return forts[i].Name < forts[j].Name
	})
	return forts
}

func (r *Registry) Fort(name string) (domain.Fort, bool) {
	if !domain.ValidFortName(name) {
		return domain.Fort{}, false
	}
	fortsMap := viper.GetStringMap("forts")
	if _, exists := fortsMap[name]; !exists {
		return domain.Fort{}, false
	}
	return r.readFort(name), true
}

func (r *Registry) readFort(name string) domain.Fort {
	prefix := "forts." + name

	fort := domain.Fort{
		Name:    name,
		Local:   viper.GetBool(prefix + ".local"),
		Gateway: viper.GetString(prefix + ".gateway"),
	}

	var svcs []struct {
		URL string `mapstructure:"url"`
	}
	if err := viper.UnmarshalKey(prefix+".services", &svcs); err == nil {
		for _, s := range svcs {
			fort.Services = append(fort.Services, domain.ConfigService{URL: s.URL})
		}
	}

	return fort
}
```

- [ ] **Step 4: Run tests**

Run: `cd /home/kazw/Work/WorkFort/scope/lead && go test ./internal/infra/fortconfig/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/infra/fortconfig/
git commit -m "refactor(fortconfig): implement Fort(name), add name validation, remove Active/SetActive"
```

---

## Chunk 2: Cookie Scoping & BFF Changes

### Task 3: Add Set-Cookie rewriting to auth proxy

**Files:**
- Modify: `internal/infra/httpapi/proxy.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/infra/httpapi/proxy_test.go
package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Work-Fort/Scope/internal/infra/httpapi"
)

func TestAuthProxy_RewritesCookiePath(t *testing.T) {
	// Auth backend sets a cookie with Path=/
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "better-auth.session_token",
			Value: "abc123",
			Path:  "/",
		})
		w.WriteHeader(200)
	}))
	defer backend.Close()

	proxy := httpapi.NewAuthProxy("auth", backend.URL, true, "", "local")

	req := httptest.NewRequest("GET", "/api/auth/session", nil)
	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected Set-Cookie header")
	}
	if cookies[0].Path != "/forts/local/" {
		t.Errorf("got cookie path %q, want %q", cookies[0].Path, "/forts/local/")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/kazw/Work/WorkFort/scope/lead && go test ./internal/infra/httpapi/... -v -run TestAuthProxy`
Expected: FAIL (NewAuthProxy doesn't exist)

- [ ] **Step 3: Implement NewAuthProxy**

Add `NewAuthProxy` and `rewriteCookiePaths` to `internal/infra/httpapi/proxy.go`. Keep the existing `NewServiceProxy` unchanged — no refactor needed. `NewAuthProxy` creates its own reverse proxy with `ModifyResponse` for cookie path rewriting:

```go
// NewAuthProxy creates a reverse proxy for auth that rewrites Set-Cookie paths
// to scope cookies to the fort's URL prefix.
func NewAuthProxy(serviceName, targetURL string, local bool, gatewayURL string, fortName string) http.Handler {
	prefix := "/api/" + serviceName
	cookiePath := "/forts/" + fortName + "/"

	if local {
		target, _ := url.Parse(targetURL)
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ModifyResponse = func(resp *http.Response) error {
			rewriteCookiePaths(resp, cookiePath)
			return nil
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
			proxy.ServeHTTP(w, r)
		})
	}

	target, _ := url.Parse(gatewayURL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ModifyResponse = func(resp *http.Response) error {
		rewriteCookiePaths(resp, cookiePath)
		return nil
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
}

func rewriteCookiePaths(resp *http.Response, path string) {
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		return
	}
	resp.Header.Del("Set-Cookie")
	for _, c := range cookies {
		c.Path = path
		if v := c.String(); v != "" {
			resp.Header.Add("Set-Cookie", v)
		}
	}
}
```

- [ ] **Step 4: Run tests**

Run: `cd /home/kazw/Work/WorkFort/scope/lead && go test ./internal/infra/httpapi/... -v -run TestAuthProxy`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/infra/httpapi/proxy.go internal/infra/httpapi/proxy_test.go
git commit -m "feat(proxy): add NewAuthProxy with Set-Cookie path rewriting"
```

### Task 4: Fort-scoped cookie path in BFF middleware

**Files:**
- Modify: `internal/infra/httpapi/handler.go` (contains `bffMiddleware`, `writeAuthError`, and `registerOneServiceRoute`)

Note: `bff.go` contains `TokenConverter` which does NOT need changes — it already takes `authServiceURL` at construction and the cookie name is unchanged.

- [ ] **Step 1: Update bffMiddleware and writeAuthError signatures**

In `handler.go`, add `fortName string` as first parameter to both functions:

```go
func bffMiddleware(fortName string, tc *TokenConverter, proxy http.Handler, wsHandler http.Handler) http.Handler {
```

```go
func writeAuthError(w http.ResponseWriter, err error, fortName string) {
```

In `writeAuthError`, update the cookie clearing to use fort-scoped path. The existing code uses `switch err` with `case` sentinel values — keep that pattern:

```go
func writeAuthError(w http.ResponseWriter, err error, fortName string) {
	w.Header().Set("Content-Type", "application/json")
	switch err {
	case errNoSession, errSessionExpired:
		w.WriteHeader(http.StatusUnauthorized)
		if err == errSessionExpired {
			http.SetCookie(w, &http.Cookie{
				Name:   sessionCookieName,
				Value:  "",
				Path:   "/forts/" + fortName + "/",
				MaxAge: -1,
			})
		}
	case errAuthDown:
		w.WriteHeader(http.StatusBadGateway)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
```

- [ ] **Step 2: Update all call sites in handler.go**

In `bffMiddleware`, pass `fortName` to `writeAuthError`:

```go
writeAuthError(w, err, fortName)
```

In `registerOneServiceRoute`, pass `fort.Name` to `bffMiddleware`:

```go
handler := bffMiddleware(fort.Name, tc, proxy, wsHandler)
```

- [ ] **Step 3: Update auth route to use NewAuthProxy**

In `registerOneServiceRoute`, replace the existing auth service proxy with `NewAuthProxy`:

```go
if svc.Name == "auth" {
	proxy := NewAuthProxy(svc.Name, svc.URL, fort.Local, fort.Gateway, fort.Name)
	mux.Handle("/api/auth/", proxy)
	return
}
```

- [ ] **Step 4: Run all tests**

Run: `cd /home/kazw/Work/WorkFort/scope/lead && go test ./internal/infra/httpapi/... -v`
Expected: PASS (existing tests still work with updated signatures)

- [ ] **Step 5: Commit**

```bash
git add internal/infra/httpapi/handler.go
git commit -m "feat(bff): fort-scoped cookie path in middleware and auth error clearing"
```

---

## Chunk 3: FortRouter & Lazy Initialization

### Task 5: FortRouter with lazy FortInstance creation

**Files:**
- Create: `internal/infra/httpapi/fort_router.go`
- Create: `internal/infra/httpapi/fort_router_test.go`

- [ ] **Step 1: Write the failing test — fort list endpoint**

```go
// internal/infra/httpapi/fort_router_test.go
package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Work-Fort/Scope/internal/domain"
	"github.com/Work-Fort/Scope/internal/infra/httpapi"
)

type mockRegistry struct {
	forts []domain.Fort
}

func (m *mockRegistry) Forts() []domain.Fort { return m.forts }
func (m *mockRegistry) Fort(name string) (domain.Fort, bool) {
	for _, f := range m.forts {
		if f.Name == name {
			return f, true
		}
	}
	return domain.Fort{}, false
}

func TestFortRouter_ListForts(t *testing.T) {
	reg := &mockRegistry{forts: []domain.Fort{
		{Name: "local", Local: true},
		{Name: "acme-corp", Local: false, Gateway: "https://fort.acme.com"},
	}}
	router := httpapi.NewFortRouter(reg, nil)

	req := httptest.NewRequest("GET", "/api/forts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("got status %d, want 200", w.Code)
	}

	var resp []map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Fatalf("got %d forts, want 2", len(resp))
	}
	if resp[0]["name"] != "local" {
		t.Errorf("first fort = %v, want local", resp[0]["name"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/kazw/Work/WorkFort/scope/lead && go test ./internal/infra/httpapi/... -v -run TestFortRouter_ListForts`
Expected: FAIL (NewFortRouter doesn't exist)

- [ ] **Step 3: Implement FortRouter**

Note: `golang.org/x/sync` is an indirect dependency. After creating this file, run `go mod tidy` to promote it to direct.

```go
// internal/infra/httpapi/fort_router.go
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

// FortInstance holds the per-fort isolation unit: tracker, token converter, handler.
type FortInstance struct {
	fort    domain.Fort
	tracker *ServiceTracker
	tc      *TokenConverter
	handler http.Handler
	lastReq atomic.Int64

	mu     sync.Mutex // protects cancel
	cancel context.CancelFunc
}

// isIdle reports whether the instance has been stopped by idle cleanup.
func (fi *FortInstance) isIdle() bool {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	return fi.cancel == nil
}

// stopPolling cancels the polling context and marks the instance as idle.
func (fi *FortInstance) stopPolling() {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	if fi.cancel != nil {
		fi.cancel()
		fi.cancel = nil
	}
}

// FortRouter is the top-level HTTP handler that dispatches to per-fort handlers.
type FortRouter struct {
	registry   domain.FortRegistry
	spaHandler http.Handler
	instances  sync.Map // map[string]*FortInstance
	initGroup  singleflight.Group
	mux        *http.ServeMux
}

// NewFortRouter creates a FortRouter that lazily initializes per-fort handlers.
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

	// Strip /forts/{fort} prefix before dispatching to the per-fort handler.
	prefix := "/forts/" + fortName
	r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}
	inst.handler.ServeHTTP(w, r)
}

func (fr *FortRouter) getInstance(ctx context.Context, name string) (*FortInstance, error) {
	if v, ok := fr.instances.Load(name); ok {
		inst := v.(*FortInstance)
		// Re-initialize if idle (cancel was called).
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

// StartIdleCleanup starts a background goroutine that stops polling for idle forts.
func (fr *FortRouter) StartIdleCleanup(ctx context.Context, maxIdle time.Duration) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
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
```

- [ ] **Step 4: Run tests**

Run: `cd /home/kazw/Work/WorkFort/scope/lead && go test ./internal/infra/httpapi/... -v -run TestFortRouter`
Expected: PASS

- [ ] **Step 5: Write test — fort dispatch routes to per-fort handler**

```go
func TestFortRouter_DispatchToFort(t *testing.T) {
	// Start test services that respond to /ui/health
	tracker, cleanup := newTestTracker(t)
	defer cleanup()

	fort := newTestFort(tracker)
	reg := &mockRegistry{forts: []domain.Fort{fort}}
	router := httpapi.NewFortRouter(reg, nil)

	req := httptest.NewRequest("GET", "/forts/local/api/services", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("got status %d, want 200", w.Code)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["fort"] != "local" {
		t.Errorf("got fort %v, want local", resp["fort"])
	}
}

func TestFortRouter_InvalidFortName_404(t *testing.T) {
	reg := &mockRegistry{forts: []domain.Fort{}}
	router := httpapi.NewFortRouter(reg, nil)

	for _, name := range []string{"../evil", "UPPER", "has space"} {
		req := httptest.NewRequest("GET", "/forts/"+name+"/api/services", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 404 {
			t.Errorf("fort %q: got status %d, want 404", name, w.Code)
		}
	}
}

func TestFortRouter_UnknownFort_404(t *testing.T) {
	reg := &mockRegistry{forts: []domain.Fort{}}
	router := httpapi.NewFortRouter(reg, nil)

	req := httptest.NewRequest("GET", "/forts/nonexistent/api/services", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("got status %d, want 404", w.Code)
	}
}
```

- [ ] **Step 6: Write test — concurrent requests use singleflight**

Add `"sync"` to the test file imports.

```go
func TestFortRouter_ConcurrentInit(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()

	fort := newTestFort(tracker)
	reg := &mockRegistry{forts: []domain.Fort{fort}}
	router := httpapi.NewFortRouter(reg, nil)

	// Fire 10 concurrent requests to the same fort.
	var wg sync.WaitGroup
	results := make([]int, 10)
	for i := range results {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/forts/local/api/services", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			results[idx] = w.Code
		}(i)
	}
	wg.Wait()

	for i, code := range results {
		if code != 200 {
			t.Errorf("request %d: got status %d, want 200", i, code)
		}
	}
}
```

- [ ] **Step 7: Run tests**

Run: `cd /home/kazw/Work/WorkFort/scope/lead && go test ./internal/infra/httpapi/... -v -run TestFortRouter`
Expected: PASS

- [ ] **Step 8: Tidy module and commit**

```bash
go mod tidy
git add go.mod go.sum internal/infra/httpapi/fort_router.go internal/infra/httpapi/fort_router_test.go
git commit -m "feat(httpapi): add FortRouter with lazy init, singleflight, idle cleanup"
```

---

## Chunk 4: Wire FortRouter into cmd/web

### Task 6: Update cmd/web/web.go

**Files:**
- Modify: `cmd/web/web.go`

- [ ] **Step 1: Replace single-fort startup with FortRouter**

Rewrite the `run` function. Key changes:
- Remove single-fort `registry.Active()`, tracker, token converter, and `topMux`
- Create `FortRouter` with full registry — lazy init handles per-fort setup
- Use existing `webFS` embed variable (defined in `cmd/web/embed_spa.go` with `//go:embed all:dist`), not `spaFS`
- Use structured logging consistent with the rest of the codebase (`charm.land/log/v2`)

```go
func run(cmd *cobra.Command, args []string) error {
	registry := fortconfig.New()
	forts := registry.Forts()
	if len(forts) == 0 {
		return fmt.Errorf("no forts configured")
	}

	var spaHandler http.Handler
	if dev {
		spaHandler = httpapi.NewSPADevProxy(devURL)
	} else {
		distFS, err := fs.Sub(webFS, "dist")
		if err != nil {
			return fmt.Errorf("embedded SPA: %w", err)
		}
		spaHandler = httpapi.NewSPAHandler(distFS)
	}

	router := httpapi.NewFortRouter(registry, spaHandler)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	router.StartIdleCleanup(ctx, 30*time.Minute)

	addr := fmt.Sprintf("%s:%d", bind, port)
	srv := &http.Server{Addr: addr, Handler: router}

	url := fmt.Sprintf("http://%s", addr)
	log.Info("web server listening", "url", url, "forts", len(forts))

	if openBrowserFlag {
		go openBrowser(url)
	}

	go func() {
		<-ctx.Done()
		log.Info("shutting down web server")
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("web server: %w", err)
	}
	return nil
}
```

Remove all references to single-fort tracker, token converter, and `topMux`. Remove unused `"github.com/Work-Fort/Scope/internal/domain"` import if present.

- [ ] **Step 2: Update imports**

Remove unused imports (`"github.com/Work-Fort/Scope/internal/domain"`), add `"context"` if not present. Keep `fortconfig` and `httpapi`.

- [ ] **Step 3: Build**

Run: `cd /home/kazw/Work/WorkFort/scope/lead && go build ./cmd/web/...`
Expected: PASS

- [ ] **Step 4: Run all Go tests**

Run: `cd /home/kazw/Work/WorkFort/scope/lead && go test ./... 2>&1 | tail -20`
Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/web/web.go
git commit -m "refactor(web): wire FortRouter, remove single-fort startup"
```

---

## Chunk 5: Frontend — API & Stores

### Task 7: Update API client with fort parameter

**Files:**
- Modify: `web/shell/src/lib/api.ts`

- [ ] **Step 1: Add FortInfo type and fetchForts, update existing functions**

```typescript
export interface FortInfo {
  name: string;
  local: boolean;
  gateway?: string;
}

export interface FortsResponse extends Array<FortInfo> {}

export interface ServiceInfo {
  name: string;
  label: string;
  route: string;
  enabled: boolean;
  ui: boolean;
  connected: boolean;
}

export interface Conflict {
  url: string;
  name: string;
  reason: string;
}

export interface ServicesResponse {
  fort: string;
  services: ServiceInfo[];
  conflicts: Conflict[];
}

export interface ConfigResponse {
  fort: string;
}

export async function fetchForts(): Promise<FortInfo[]> {
  const res = await fetch('/api/forts');
  if (!res.ok) throw new Error(`/api/forts: ${res.status}`);
  return res.json();
}

export async function fetchServices(fort: string): Promise<ServicesResponse> {
  const res = await fetch(`/forts/${fort}/api/services`);
  if (!res.ok) throw new Error(`/forts/${fort}/api/services: ${res.status}`);
  return res.json();
}

export async function fetchConfig(fort: string): Promise<ConfigResponse> {
  const res = await fetch(`/forts/${fort}/api/config`);
  if (!res.ok) throw new Error(`/forts/${fort}/api/config: ${res.status}`);
  return res.json();
}
```

- [ ] **Step 2: Commit**

```bash
git add web/shell/src/lib/api.ts
git commit -m "feat(api): add fetchForts, scope all endpoints to fort path"
```

### Task 8: Fort-scoped services store

**Files:**
- Modify: `web/shell/src/stores/services.ts`

- [ ] **Step 1: Rewrite store to accept fort name parameter**

```typescript
import { createSignal } from 'solid-js';
import { fetchServices, type ServiceInfo, type Conflict, type ServicesResponse } from '../lib/api';
import { registerNewRemotes } from '../lib/remotes';
import { addBanner, removeBanner, banners } from './banners';
import { addToast } from './toasts';

const POLL_INTERVAL = 30_000;

const [serviceList, setServiceList] = createSignal<ServiceInfo[]>([]);
const [conflictList, setConflictList] = createSignal<Conflict[]>([]);
const [currentFort, setCurrentFort] = createSignal('');

let prevConnected = new Map<string, boolean>();

function handlePollResult(res: ServicesResponse): void {
  setCurrentFort(res.fort);
  setConflictList(res.conflicts ?? []);

  const nextConnected = new Map<string, boolean>();
  for (const svc of res.services) {
    nextConnected.set(svc.name, svc.connected);
    const was = prevConnected.get(svc.name);
    if (was !== undefined && was !== svc.connected) {
      if (svc.connected) {
        addToast('success', `${svc.label} reconnected`);
        removeBanner(`disconnected:${svc.name}`);
      } else {
        addToast('error', `${svc.label} disconnected`, { sticky: true });
        addBanner(
          `disconnected:${svc.name}`,
          'warning',
          `${svc.label} is not responding`,
          `Service "${svc.name}" is unreachable. This page will update when it recovers.`,
          'system',
        );
      }
    }
  }
  prevConnected = nextConnected;

  registerNewRemotes(res.services);

  const activeConflictKeys = new Set((res.conflicts ?? []).map((c) => `conflict:${c.name}`));
  for (const conflict of res.conflicts ?? []) {
    addBanner(
      `conflict:${conflict.name}`,
      'error',
      `Service conflict: "${conflict.name}"`,
      `${conflict.reason}\nURL: ${conflict.url}`,
      'system',
    );
  }
  for (const b of banners()) {
    if (b.key.startsWith('conflict:') && !activeConflictKeys.has(b.key)) {
      removeBanner(b.key);
    }
  }

  for (const svc of res.services) {
    if (svc.connected) {
      removeBanner(`disconnected:${svc.name}`);
    }
  }

  setServiceList(res.services);
}

let intervalId: ReturnType<typeof setInterval> | null = null;
let activeFort: string | null = null;

export function startPolling(fort: string): void {
  // If fort changed, reset state.
  if (activeFort !== fort) {
    stopPolling();
    prevConnected = new Map();
    setServiceList([]);
    setConflictList([]);
  }
  activeFort = fort;
  fetchServices(fort).then(handlePollResult).catch(console.error);
  intervalId = setInterval(() => {
    fetchServices(fort).then(handlePollResult).catch(console.error);
  }, POLL_INTERVAL);
}

export function stopPolling(): void {
  if (intervalId) {
    clearInterval(intervalId);
    intervalId = null;
  }
  activeFort = null;
}

export const services = serviceList;
export const conflicts = conflictList;
export const fortName = currentFort;
```

- [ ] **Step 2: Commit**

```bash
git add web/shell/src/stores/services.ts
git commit -m "feat(stores): fort-scoped service polling"
```

---

## Chunk 6: Frontend — Components & Router

### Task 9: Fort picker component

**Files:**
- Create: `web/shell/src/components/fort-picker.tsx`

- [ ] **Step 1: Create fort picker**

```typescript
import { createResource, Show, For, type Component } from 'solid-js';
import { Navigate, useNavigate } from '@solidjs/router';
import { fetchForts } from '../lib/api';

const FortPicker: Component = () => {
  const [forts] = createResource(fetchForts);
  const navigate = useNavigate();

  return (
    <Show when={!forts.loading} fallback={<wf-skeleton width="100%" height="200px" />}>
      <Show
        when={forts() && forts()!.length !== 1}
        fallback={
          forts() && forts()!.length === 1
            ? <Navigate href={`/forts/${forts()![0].name}`} />
            : <div class="shell-unavailable">No forts configured.</div>
        }
      >
        <div class="fort-picker">
          <h2 class="fort-picker__title">Select a Fort</h2>
          <wf-list>
            <For each={forts()}>
              {(fort) => (
                <wf-list-item on:wf-select={() => navigate(`/forts/${fort.name}`)}>
                  {fort.name}
                </wf-list-item>
              )}
            </For>
          </wf-list>
        </div>
      </Show>
    </Show>
  );
};

export default FortPicker;
```

- [ ] **Step 2: Commit**

```bash
git add web/shell/src/components/fort-picker.tsx
git commit -m "feat(shell): add fort picker component"
```

### Task 10: Update app router for fort-scoped routes

**Files:**
- Modify: `web/shell/src/app.tsx`
- Modify: `web/shell/src/components/nav-bar.tsx`

- [ ] **Step 1: Rewrite app.tsx with fort-scoped routing**

All imports consolidated at the top — no mid-file imports:

```typescript
import {
  type Component,
  Show,
  createContext,
  createEffect,
  createMemo,
  createSignal,
  onCleanup,
  useContext,
} from 'solid-js';
import { Navigate, Route, Router, useParams } from '@solidjs/router';
import { services, startPolling, stopPolling } from './stores/services';
import './stores/theme';
import ShellLayout from './components/shell-layout';
import ServiceMount from './components/service-mount';
import Unavailable from './components/unavailable';
import FortPicker from './components/fort-picker';
import type { ServiceModule } from './lib/remotes';

// Context to pass sidebar setter from FortShell to ServicePage.
const FortShellContext = createContext<{
  setSidebarComponent: (v: (() => any) | undefined) => void;
}>({ setSidebarComponent: () => {} });

const App: Component = () => {
  return (
    <Router>
      <Route path="/" component={FortPicker} />
      <Route path="/forts/:fort" component={FortShell}>
        <Route path="/:service/*rest" component={ServicePage} />
        <Route path="/" component={FortIndex} />
      </Route>
    </Router>
  );
};

const FortShell: Component = (props: { children?: any }) => {
  const params = useParams<{ fort: string }>();
  const [sidebarComponent, setSidebarComponent] = createSignal<(() => any) | undefined>();

  createEffect(() => {
    const fort = params.fort;
    startPolling(fort);
  });
  onCleanup(() => stopPolling());

  return (
    <FortShellContext.Provider value={{ setSidebarComponent }}>
      <ShellLayout sidebar={sidebarComponent()}>{props.children}</ShellLayout>
    </FortShellContext.Provider>
  );
};

const ServicePage: Component = () => {
  const params = useParams<{ fort: string; service: string }>();
  const ctx = useContext(FortShellContext);

  const handleModule = (mod: ServiceModule | null) => {
    ctx.setSidebarComponent(
      mod?.SidebarContent ? () => mod.SidebarContent! : undefined,
    );
  };

  const svc = createMemo(() =>
    services().find((s) => s.enabled && s.route === `/${params.service}`),
  );

  return (
    <>
      {svc() ? (
        svc()!.ui ? (
          <ServiceMount name={svc()!.name} label={svc()!.label} connected={svc()!.connected} onModule={handleModule} />
        ) : (
          <Unavailable label={svc()!.label} />
        )
      ) : (
        <Navigate href={`/forts/${params.fort}`} />
      )}
    </>
  );
};

const FortIndex: Component = () => {
  const params = useParams<{ fort: string }>();
  const firstRoute = createMemo(() => {
    const enabled = services().find((s) => s.enabled);
    return enabled ? `/forts/${params.fort}${enabled.route}` : null;
  });

  return (
    <Show when={firstRoute()} fallback={<div class="shell-unavailable">No services available.</div>}>
      <Navigate href={firstRoute()!} />
    </Show>
  );
};

export default App;
```

- [ ] **Step 2: Update nav-bar.tsx with fort-scoped navigation**

```typescript
import { For, Show, type Component } from 'solid-js';
import { useNavigate, useLocation, useParams } from '@solidjs/router';
import { services, fortName } from '../stores/services';
import { toggleTheme } from '../stores/theme';
import { useTheme } from '@workfort/ui-solid';

const NavBar: Component = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const params = useParams<{ fort: string }>();
  const theme = useTheme();

  return (
    <nav class="shell-nav">
      <span class="shell-nav__brand">{fortName() || 'WorkFort'}</span>
      <wf-list class="shell-nav__tabs">
        <For each={services().filter((s) => s.enabled)}>
          {(svc) => (
            <wf-list-item
              active={location.pathname.includes(svc.route)}
              class={!svc.ui ? 'shell-nav__tab--disabled' : ''}
              on:wf-select={() => navigate(`/forts/${params.fort}${svc.route}`)}
            >
              <Show when={svc.ui}>
                <wf-status-dot status={svc.connected ? 'online' : 'offline'} />
              </Show>
              {svc.label}
            </wf-list-item>
          )}
        </For>
      </wf-list>
      <div class="shell-nav__spacer" />
      <wf-button variant="text" on:wf-click={() => toggleTheme()}>
        {theme() === 'dark' ? 'Light' : 'Dark'}
      </wf-button>
    </nav>
  );
};

export default NavBar;
```

- [ ] **Step 3: Commit**

```bash
git add web/shell/src/app.tsx web/shell/src/components/nav-bar.tsx
git commit -m "feat(shell): fort-scoped routing with picker, FortShell layout"
```

---

## Verification

After all tasks complete:

- [ ] **Start dev servers:** `mise run dev:go` + `mise run dev:web`
- [ ] **Visit root:** `http://127.0.0.1:16100/` — should show fort picker (or auto-redirect if single fort)
- [ ] **Navigate to fort:** `/forts/local/` — should redirect to first enabled service
- [ ] **Check API:** `curl http://127.0.0.1:16100/api/forts` — should list all forts
- [ ] **Check fort API:** `curl http://127.0.0.1:16100/forts/local/api/services` — should return services
- [ ] **Check invalid fort:** `curl http://127.0.0.1:16100/forts/INVALID/api/services` — should 404
- [ ] **Check unknown fort:** `curl http://127.0.0.1:16100/forts/nonexistent/api/services` — should 404
- [ ] **Run all Go tests:** `mise run test`
- [ ] **Verify cookie scoping:** Auth login in one fort should not send cookies to another fort's endpoints
