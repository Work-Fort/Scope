# Service Discovery & Frontend Status — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement dynamic service discovery with self-describing services, health probing, WS connection tracking, conflict detection, and a complete frontend status UX with banners and toasts.

**Architecture:** Services self-describe via `pkg/frontend.Manifest` returned from `/ui/health`. A Go `ServiceTracker` probes services and tracks WS connection state. The shell SPA polls `/api/services` and surfaces all status/errors in the chrome via banners, toasts, and nav status dots.

**Tech Stack:** Go 1.23+, Lit web components (`@workfort/ui`), SolidJS (shell SPA), Module Federation runtime, gorilla/websocket.

**Spec:** `docs/2026-03-13-service-discovery-design.md`

---

## File Structure

### New files
| File | Responsibility |
|------|---------------|
| `internal/infra/httpapi/tracker.go` | ServiceTracker — background health probing, connection ref-counting, conflict detection |
| `internal/infra/httpapi/tracker_test.go` | Tests for ServiceTracker |
| `web/packages/ui/src/components/banner.ts` | `wf-banner` web component (imperative DOM, no Shadow DOM slots) |
| `web/packages/ui/src/components/toast.ts` | `wf-toast` web component |
| `web/packages/ui/src/components/toast-container.ts` | `wf-toast-container` positioned wrapper |
| `web/packages/ui/src/styles/banner.css` | Banner component styles |
| `web/packages/ui/src/styles/toast.css` | Toast component styles |
| `web/shell/src/stores/banners.ts` | Banner store — add/remove, shell + remote access |
| `web/shell/src/stores/toasts.ts` | Toast store — add/dismiss with auto-dismiss timer |

### Modified files
| File | Changes |
|------|---------|
| `pkg/frontend/frontend.go` | `Handler()` takes `Manifest`, `/ui/health` ALWAYS returns manifest (200 or 503) |
| `pkg/frontend/frontend_test.go` | Update tests for new `Handler(fsys, manifest)` signature |
| `internal/domain/web.go` | Replace `Service` with `ConfigService` (URL-only), update `Fort` |
| `internal/infra/fortconfig/registry.go` | Parse simplified config (list of URL structs) |
| `internal/infra/fortconfig/registry_test.go` | Update test config format and assertions |
| `internal/infra/httpapi/ws.go` | `NewWSProxy` accepts `ConnectionCallbacks` |
| `internal/infra/httpapi/ws_test.go` | Test callback invocation on connect/disconnect |
| `internal/infra/httpapi/proxy.go` | `NewServiceProxy` takes `(name, url string, local bool, gateway string)` |
| `internal/infra/httpapi/proxy_test.go` | Update for new signature |
| `internal/infra/httpapi/handler.go` | Delete `serviceMetadata`, wire ServiceTracker, dynamic route registration |
| `internal/infra/httpapi/handler_test.go` | Update for new `NewHandler` signature with tracker |
| `cmd/web/web.go` | Create tracker, find auth URL from tracker, pass tracker to handler |
| `web/packages/ui/src/styles/tokens.css` | Add `--wf-error`, `--wf-warning`, `--wf-success` tokens |
| `web/packages/ui/src/styles/components.css` | Import banner and toast styles |
| `web/packages/ui/src/index.ts` | Register `wf-banner`, `wf-toast`, `wf-toast-container` |
| `web/shell/src/lib/api.ts` | Add `connected`, `conflicts` to types |
| `web/shell/src/lib/remotes.ts` | `ServiceModule.default` takes `{ connected }` prop, remove `initialized` guard |
| `web/shell/src/stores/services.ts` | Polling loop, expose `conflicts()`, incremental remote registration |
| `web/shell/src/components/service-mount.tsx` | Pass `connected` prop to remote, show banner when never-loaded + disconnected |
| `web/shell/src/components/nav-bar.tsx` | Status dots on nav tabs |
| `web/shell/src/components/shell-layout.tsx` | Banner bar above nav, toast container |
| `web/shell/src/global.css` | Grid update: `banners` row above nav |

---

## Chunk 1: `pkg/frontend` Manifest

### Task 1: Add Manifest type and update Handler signature

**Files:**
- Modify: `pkg/frontend/frontend.go`
- Test: `pkg/frontend/frontend_test.go`

**Design note:** The spec says 503 responses return `{"status":"unavailable"}` but this creates a problem — the tracker can't discover a service's identity if it returns 503 (no UI built). Fix: the manifest is ALWAYS returned in the `/ui/health` response regardless of status code. The service always knows its own identity. This allows the tracker to register no-UI services correctly (grayed out in nav, `ui: false`).

- [ ] **Step 1: Write the failing test for manifest in health response**

In `pkg/frontend/frontend_test.go`, add `"encoding/json"` to imports, then add:

```go
func TestHealthProbe_WithManifest(t *testing.T) {
	fsys := fstest.MapFS{
		"remoteEntry.js": &fstest.MapFile{Data: []byte("// entry")},
	}

	handler := frontend.Handler(fsys, frontend.Manifest{
		Name:    "sharkfin",
		Label:   "Chat",
		Route:   "/chat",
		WSPaths: []string{"/ws", "/presence"},
	})

	req := httptest.NewRequest(http.MethodGet, "/ui/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Status  string   `json:"status"`
		Name    string   `json:"name"`
		Label   string   `json:"label"`
		Route   string   `json:"route"`
		WSPaths []string `json:"ws_paths"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("expected status ok, got %q", resp.Status)
	}
	if resp.Name != "sharkfin" {
		t.Fatalf("expected name sharkfin, got %q", resp.Name)
	}
	if resp.Label != "Chat" {
		t.Fatalf("expected label Chat, got %q", resp.Label)
	}
	if resp.Route != "/chat" {
		t.Fatalf("expected route /chat, got %q", resp.Route)
	}
	if len(resp.WSPaths) != 2 || resp.WSPaths[0] != "/ws" {
		t.Fatalf("unexpected ws_paths: %v", resp.WSPaths)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/frontend/ -run TestHealthProbe_WithManifest -v`
Expected: FAIL — `Handler` doesn't accept a `Manifest` argument.

- [ ] **Step 3: Implement Manifest type and update Handler**

In `pkg/frontend/frontend.go`, add the `Manifest` struct before `Handler`:

```go
// Manifest describes a service's frontend identity.
// Returned in the /ui/health response for service discovery.
type Manifest struct {
	Name    string   `json:"name"`
	Label   string   `json:"label"`
	Route   string   `json:"route"`
	WSPaths []string `json:"ws_paths,omitempty"`
}
```

Update `Handler` signature and health endpoint:

```go
func Handler(fsys fs.FS, m Manifest) http.Handler {
	hasRemoteEntry := fileExists(fsys, "remoteEntry.js")
	fileServer := http.StripPrefix("/ui/", http.FileServer(http.FS(fsys)))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ui/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		type healthResponse struct {
			Status string `json:"status"`
			Manifest
		}

		if hasRemoteEntry {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok", Manifest: m})
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(healthResponse{Status: "unavailable", Manifest: m})
		}
	})

	mux.HandleFunc("/ui/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/ui/")
		if strings.HasPrefix(path, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		fileServer.ServeHTTP(w, r)
	})

	return mux
}
```

Key change: the 503 response now also includes the manifest fields. This allows the tracker to learn a service's identity even when its UI isn't built yet.

- [ ] **Step 4: Update existing tests for new signature**

All existing tests call `frontend.Handler(fsys)` — update them to `frontend.Handler(fsys, frontend.Manifest{})`.

Update `TestHealthProbe_Unavailable` to verify the 503 response still includes manifest fields when provided:

```go
func TestHealthProbe_Unavailable_IncludesManifest(t *testing.T) {
	fsys := fstest.MapFS{
		"other.js": &fstest.MapFile{Data: []byte("// other")},
	}

	handler := frontend.Handler(fsys, frontend.Manifest{
		Name:  "auth",
		Label: "Auth",
		Route: "/auth",
	})

	req := httptest.NewRequest(http.MethodGet, "/ui/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var resp struct {
		Status string `json:"status"`
		Name   string `json:"name"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Status != "unavailable" {
		t.Fatalf("expected status unavailable, got %q", resp.Status)
	}
	if resp.Name != "auth" {
		t.Fatalf("expected name auth in 503 response, got %q", resp.Name)
	}
}
```

- [ ] **Step 5: Run all frontend tests**

Run: `go test ./pkg/frontend/ -v`
Expected: All PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/frontend/frontend.go pkg/frontend/frontend_test.go
git commit -m "feat(frontend): add Manifest type, return it in /ui/health (200 and 503)"
```

---

## Chunk 2: Go Backend — Domain, Config, Proxy, WS, Tracker, Handler

This chunk changes the domain types, config parsing, proxy signature, WS callbacks, tracker, and handler in a single buildable unit. All files are committed together so the build stays green at every commit boundary.

### Task 2: Simplify domain types, config, and proxy signature

**Files:**
- Modify: `internal/domain/web.go`
- Modify: `internal/infra/fortconfig/registry.go`
- Modify: `internal/infra/fortconfig/registry_test.go`
- Modify: `internal/infra/httpapi/proxy.go`
- Modify: `internal/infra/httpapi/proxy_test.go`

- [ ] **Step 1: Replace Service with ConfigService in domain**

In `internal/domain/web.go`, replace the `Service` struct with:

```go
// ConfigService is what comes from the fort config file — just a URL.
// All configured services are considered enabled. To disable a service,
// remove it from the config.
type ConfigService struct {
	URL string
}
```

Update `Fort`:

```go
type Fort struct {
	Name     string
	Local    bool
	Gateway  string
	Services []ConfigService
}
```

Delete the old `Service` struct entirely.

- [ ] **Step 2: Update fortconfig to parse URL list**

In `internal/infra/fortconfig/registry.go`, replace `readFort`'s service-parsing:

```go
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

Remove `sort` import (services are in config order now).

- [ ] **Step 3: Update fortconfig tests**

In `internal/infra/fortconfig/registry_test.go`, replace `setupViper`:

```go
func setupViper(t *testing.T) {
	t.Helper()
	viper.Reset()

	viper.Set("active-fort", "local")
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services", []map[string]string{
		{"url": "http://127.0.0.1:16000"},
		{"url": "http://127.0.0.1:9600"},
		{"url": "http://127.0.0.1:3000"},
	})
}
```

Update `TestActive` to check for 3 `ConfigService` entries with URLs:

```go
func TestActive(t *testing.T) {
	setupViper(t)
	reg := fortconfig.New()
	fort := reg.Active()

	if fort.Name != "local" {
		t.Fatalf("expected fort name 'local', got %q", fort.Name)
	}
	if !fort.Local {
		t.Fatal("expected fort to be local")
	}
	if len(fort.Services) != 3 {
		t.Fatalf("expected 3 services, got %d", len(fort.Services))
	}

	urls := make(map[string]bool)
	for _, svc := range fort.Services {
		urls[svc.URL] = true
	}
	if !urls["http://127.0.0.1:16000"] {
		t.Fatal("missing sharkfin URL")
	}
	if !urls["http://127.0.0.1:9600"] {
		t.Fatal("missing nexus URL")
	}
	if !urls["http://127.0.0.1:3000"] {
		t.Fatal("missing auth URL")
	}
}
```

Update `TestForts` to use the new config format for both `local` and `remote` forts. Update `TestSetActive_Valid` and `TestSetActive_Invalid` similarly.

- [ ] **Step 4: Update proxy.go signature**

In `internal/infra/httpapi/proxy.go`, change `NewServiceProxy` to take simple arguments instead of `domain.Service`:

```go
// NewServiceProxy creates an http.Handler that proxies requests to a service.
func NewServiceProxy(serviceName, targetURL string, local bool, gatewayURL string) http.Handler {
	prefix := "/api/" + serviceName

	if local {
		target, _ := url.Parse(targetURL)
		proxy := httputil.NewSingleHostReverseProxy(target)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
			proxy.ServeHTTP(w, r)
		})
	}

	// Gateway mode — preserve the /api/{service} prefix.
	target, _ := url.Parse(gatewayURL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
}
```

Remove `domain` import from `proxy.go`. The `Enabled` check is gone — all tracked services are enabled by definition (disabled = removed from config).

- [ ] **Step 5: Update proxy tests**

In `proxy_test.go`, update tests to use the new signature:

```go
func TestProxy_PathStripping(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer backend.Close()

	proxy := httpapi.NewServiceProxy("nexus", backend.URL, true, "")

	req := httptest.NewRequest(http.MethodGet, "/api/nexus/v1/vms", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	body, _ := io.ReadAll(rec.Body)
	if string(body) != "/v1/vms" {
		t.Fatalf("expected path /v1/vms, got %q", string(body))
	}
}
```

Remove `TestProxy_DisabledService` (disabled services no longer exist — removed from config instead).

Update `TestProxy_GatewayFort`:

```go
func TestProxy_GatewayFort(t *testing.T) {
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer gateway.Close()

	proxy := httpapi.NewServiceProxy("nexus", "", false, gateway.URL)
	// ...same assertions
}
```

Remove `domain` import from `proxy_test.go`.

- [ ] **Step 6: Run fortconfig and proxy tests**

Run: `go test ./internal/infra/fortconfig/ ./internal/infra/httpapi/ -run "TestActive|TestForts|TestSetActive|TestProxy" -v`
Expected: All PASS (handler tests will still fail until Task 5).

- [ ] **Step 7: Commit (fortconfig + proxy only)**

```bash
git add internal/domain/web.go internal/infra/fortconfig/ internal/infra/httpapi/proxy.go internal/infra/httpapi/proxy_test.go
git commit -m "feat(domain): simplify Service to ConfigService (URL-only), update fortconfig and proxy"
```

### Task 3: Add ConnectionCallbacks to NewWSProxy

**Files:**
- Modify: `internal/infra/httpapi/ws.go`
- Modify: `internal/infra/httpapi/ws_test.go`

- [ ] **Step 1: Write the failing test for connection callbacks**

In `ws_test.go`, add `"sync/atomic"` and `"time"` to imports, then add:

```go
func TestWSProxy_ConnectionCallbacks(t *testing.T) {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_, _, _ = conn.ReadMessage()
	}))
	defer backend.Close()

	var connected, disconnected int32
	cb := &httpapi.ConnectionCallbacks{
		OnConnect:    func(svc string) { atomic.AddInt32(&connected, 1) },
		OnDisconnect: func(svc string) { atomic.AddInt32(&disconnected, 1) },
	}

	backendURL := "ws" + strings.TrimPrefix(backend.URL, "http")
	wsHandler := httpapi.NewWSProxy(backendURL, []string{"/ws"}, "nexus", cb)

	proxy := httptest.NewServer(wsHandler)
	defer proxy.Close()

	proxyURL := "ws" + strings.TrimPrefix(proxy.URL, "http") + "/api/nexus/ws"
	conn, _, err := websocket.DefaultDialer.Dial(proxyURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if c := atomic.LoadInt32(&connected); c != 1 {
		t.Fatalf("expected 1 connect callback, got %d", c)
	}

	conn.Close()
	time.Sleep(50 * time.Millisecond)
	if d := atomic.LoadInt32(&disconnected); d != 1 {
		t.Fatalf("expected 1 disconnect callback, got %d", d)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/infra/httpapi/ -run TestWSProxy_ConnectionCallbacks -v`
Expected: FAIL — `NewWSProxy` doesn't accept 4th argument.

- [ ] **Step 3: Add ConnectionCallbacks to NewWSProxy**

In `ws.go`, add the type and update the signature:

```go
// ConnectionCallbacks notifies the caller of WebSocket connection lifecycle events.
type ConnectionCallbacks struct {
	OnConnect    func(service string)
	OnDisconnect func(service string)
}

func NewWSProxy(backendURL string, wsPaths []string, serviceName string, cb *ConnectionCallbacks) http.Handler {
```

In the handler func, after successfully upgrading the client connection (`defer clientConn.Close()`), add:

```go
		if cb != nil && cb.OnConnect != nil {
			cb.OnConnect(serviceName)
		}
		if cb != nil && cb.OnDisconnect != nil {
			defer cb.OnDisconnect(serviceName)
		}
```

- [ ] **Step 4: Update existing WS tests for new signature**

Pass `nil` as the 4th argument:

```go
httpapi.NewWSProxy(backendURL, []string{"/ws", "/presence"}, "nexus", nil)
httpapi.NewWSProxy("ws://localhost:0", []string{"/ws"}, "nexus", nil)
```

- [ ] **Step 5: Run all WS tests**

Run: `go test ./internal/infra/httpapi/ -run TestWSProxy -v`
Expected: All PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/infra/httpapi/ws.go internal/infra/httpapi/ws_test.go
git commit -m "feat(ws): add ConnectionCallbacks to NewWSProxy"
```

### Task 4: Implement ServiceTracker

**Files:**
- Create: `internal/infra/httpapi/tracker.go`
- Create: `internal/infra/httpapi/tracker_test.go`

- [ ] **Step 1: Write the failing test for tracker initial probe**

In `internal/infra/httpapi/tracker_test.go`:

```go
package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Work-Fort/Scope/internal/infra/httpapi"
)

func TestTracker_InitialProbe(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":   "ok",
			"name":     "sharkfin",
			"label":    "Chat",
			"route":    "/chat",
			"ws_paths": []string{"/ws"},
		})
	}))
	defer svc.Close()

	tracker := httpapi.NewServiceTracker([]string{svc.URL})
	tracker.InitialProbe(context.Background())

	services := tracker.Services()
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
	if services[0].Name != "sharkfin" {
		t.Fatalf("expected name sharkfin, got %q", services[0].Name)
	}
	if services[0].Label != "Chat" {
		t.Fatalf("expected label Chat, got %q", services[0].Label)
	}
	if !services[0].UI {
		t.Fatal("expected ui=true")
	}
	// WS service starts disconnected (no active connections yet).
	if services[0].Connected {
		t.Fatal("expected connected=false for WS service before any connections")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/infra/httpapi/ -run TestTracker_InitialProbe -v`
Expected: FAIL — `NewServiceTracker` doesn't exist.

- [ ] **Step 3: Implement ServiceTracker core**

In `internal/infra/httpapi/tracker.go`:

```go
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
	byName    map[string]int // name → index into services
	byRoute   map[string]int // route → index into services

	// OnServiceDiscovered is called when a new service is discovered after
	// the initial probe (i.e., a service that was down at startup comes up).
	// Called WITHOUT holding the mutex — safe to call back into the tracker.
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
		// Service unreachable — mark disconnected if already known.
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

	// Both 200 and 503 include the manifest (identity is always returned).
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
			t.services[idx].Connected = true // Non-WS: connected if reachable.
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
		Connected: !hasWS, // Non-WS: connected if reachable. WS: starts false.
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
```

- [ ] **Step 4: Run initial probe test**

Run: `go test ./internal/infra/httpapi/ -run TestTracker_InitialProbe -v`
Expected: PASS.

- [ ] **Step 5: Write and run conflict detection test**

```go
func TestTracker_ConflictDetection(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"name":   "sharkfin",
			"label":  "Chat",
			"route":  "/chat",
		})
	})

	svc1 := httptest.NewServer(handler)
	defer svc1.Close()
	svc2 := httptest.NewServer(handler)
	defer svc2.Close()

	// Sequential probe — svc1 wins because it's first in the list.
	tracker := httpapi.NewServiceTracker([]string{svc1.URL, svc2.URL})
	tracker.InitialProbe(context.Background())

	services := tracker.Services()
	if len(services) != 1 {
		t.Fatalf("expected 1 service (first wins), got %d", len(services))
	}
	if services[0].URL != svc1.URL {
		t.Fatalf("expected first URL to win, got %q", services[0].URL)
	}

	conflicts := tracker.Conflicts()
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].URL != svc2.URL {
		t.Fatalf("expected conflict URL to be svc2, got %q", conflicts[0].URL)
	}
}
```

Run: `go test ./internal/infra/httpapi/ -run TestTracker_ConflictDetection -v`
Expected: PASS.

- [ ] **Step 6: Write and run unreachable service test**

```go
func TestTracker_UnreachableService(t *testing.T) {
	tracker := httpapi.NewServiceTracker([]string{"http://127.0.0.1:1"})
	tracker.InitialProbe(context.Background())

	if len(tracker.Services()) != 0 {
		t.Fatalf("expected 0 services for unreachable URL, got %d", len(tracker.Services()))
	}
}
```

Run: `go test ./internal/infra/httpapi/ -run TestTracker_UnreachableService -v`
Expected: PASS.

- [ ] **Step 7: Write and run 503 (no UI) service test**

```go
func TestTracker_NoUIService(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "unavailable",
			"name":   "auth",
			"label":  "Auth",
			"route":  "/auth",
		})
	}))
	defer svc.Close()

	tracker := httpapi.NewServiceTracker([]string{svc.URL})
	tracker.InitialProbe(context.Background())

	services := tracker.Services()
	if len(services) != 1 {
		t.Fatalf("expected 1 service (503 still registers), got %d", len(services))
	}
	if services[0].Name != "auth" {
		t.Fatalf("expected name auth, got %q", services[0].Name)
	}
	if services[0].UI {
		t.Fatal("expected ui=false for 503 service")
	}
	if !services[0].Connected {
		t.Fatal("expected connected=true for non-WS reachable service")
	}
}
```

Run: `go test ./internal/infra/httpapi/ -run TestTracker_NoUIService -v`
Expected: PASS.

- [ ] **Step 8: Write and run WS connection ref-counting test**

```go
func TestTracker_WSConnectionTracking(t *testing.T) {
	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":   "ok",
			"name":     "sharkfin",
			"label":    "Chat",
			"route":    "/chat",
			"ws_paths": []string{"/ws"},
		})
	}))
	defer svc.Close()

	tracker := httpapi.NewServiceTracker([]string{svc.URL})
	tracker.InitialProbe(context.Background())

	// WS service starts disconnected.
	if tracker.Services()[0].Connected {
		t.Fatal("WS service should start disconnected")
	}

	// First connection.
	tracker.OnConnect("sharkfin")
	if !tracker.Services()[0].Connected {
		t.Fatal("expected connected after OnConnect")
	}

	// Second connection.
	tracker.OnConnect("sharkfin")
	// Disconnect one — still connected (ref count > 0).
	tracker.OnDisconnect("sharkfin")
	if !tracker.Services()[0].Connected {
		t.Fatal("expected still connected (ref count = 1)")
	}

	// Disconnect last.
	tracker.OnDisconnect("sharkfin")
	if tracker.Services()[0].Connected {
		t.Fatal("expected disconnected (ref count = 0)")
	}
}
```

Run: `go test ./internal/infra/httpapi/ -run TestTracker_WSConnectionTracking -v`
Expected: PASS.

- [ ] **Step 9: Write and run background polling test**

```go
func TestTracker_BackgroundPolling(t *testing.T) {
	var probeCount int32
	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&probeCount, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"name":   "nexus",
			"label":  "Nexus",
			"route":  "/nexus",
		})
	}))
	defer svc.Close()

	tracker := httpapi.NewServiceTracker([]string{svc.URL})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracker.InitialProbe(ctx)
	tracker.StartPolling(ctx, 50*time.Millisecond)

	time.Sleep(200 * time.Millisecond)
	cancel()

	if c := atomic.LoadInt32(&probeCount); c < 3 {
		t.Fatalf("expected at least 3 probes, got %d", c)
	}
}
```

Run: `go test ./internal/infra/httpapi/ -run TestTracker_BackgroundPolling -v`
Expected: PASS.

- [ ] **Step 10: Write and run "service comes back up" test**

```go
func TestTracker_ServiceComesBackUp(t *testing.T) {
	var respondOK int32 // 0 = fail, 1 = respond OK

	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&respondOK) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"name":   "nexus",
			"label":  "Nexus",
			"route":  "/nexus",
		})
	}))
	defer svc.Close()

	var discovered int32
	tracker := httpapi.NewServiceTracker([]string{svc.URL})
	tracker.OnServiceDiscovered = func(svc httpapi.TrackedService) {
		atomic.AddInt32(&discovered, 1)
	}

	// Initial probe — service is down.
	tracker.InitialProbe(context.Background())
	if len(tracker.Services()) != 0 {
		t.Fatal("expected 0 services while down")
	}

	// Bring service up.
	atomic.StoreInt32(&respondOK, 1)

	// Start polling — should discover the service.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tracker.StartPolling(ctx, 50*time.Millisecond)

	time.Sleep(200 * time.Millisecond)
	cancel()

	if len(tracker.Services()) != 1 {
		t.Fatalf("expected 1 service after coming up, got %d", len(tracker.Services()))
	}
	if atomic.LoadInt32(&discovered) != 1 {
		t.Fatal("expected OnServiceDiscovered to fire once")
	}
}
```

Run: `go test ./internal/infra/httpapi/ -run TestTracker_ServiceComesBackUp -v`
Expected: PASS.

- [ ] **Step 11: Commit tracker**

```bash
git add internal/infra/httpapi/tracker.go internal/infra/httpapi/tracker_test.go
git commit -m "feat(tracker): ServiceTracker with health probing, WS ref-counting, conflict detection"
```

### Task 5: Wire ServiceTracker into handler

**Files:**
- Modify: `internal/infra/httpapi/handler.go`
- Modify: `internal/infra/httpapi/handler_test.go`
- Modify: `cmd/web/web.go`

- [ ] **Step 1: Rewrite handler.go**

Delete `serviceMetadata` entirely. Update `NewHandler` signature:

```go
func NewHandler(fort domain.Fort, tracker *ServiceTracker, tc *TokenConverter, spaFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	// Shell endpoints.
	mux.HandleFunc("GET /api/services", servicesHandler(fort.Name, tracker))
	mux.HandleFunc("GET /api/config", configHandler(fort.Name))

	// Register proxy routes for discovered services.
	registerServiceRoutes(mux, tracker, tc, fort)

	// Wire up dynamic route registration for late-arriving services.
	tracker.OnServiceDiscovered = func(svc TrackedService) {
		registerOneServiceRoute(mux, svc, tracker, tc, fort)
	}

	// SPA fallback.
	if spaFS != nil {
		mux.Handle("/", NewSPAHandler(spaFS))
	}

	return mux
}

func registerServiceRoutes(mux *http.ServeMux, tracker *ServiceTracker, tc *TokenConverter, fort domain.Fort) {
	for _, svc := range tracker.Services() {
		registerOneServiceRoute(mux, svc, tracker, tc, fort)
	}
}

func registerOneServiceRoute(mux *http.ServeMux, svc TrackedService, tracker *ServiceTracker, tc *TokenConverter, fort domain.Fort) {
	prefix := "/api/" + svc.Name + "/"

	if svc.Name == "auth" {
		proxy := NewServiceProxy(svc.Name, svc.URL, fort.Local, fort.Gateway)
		mux.Handle(prefix, proxy)
		return
	}

	proxy := NewServiceProxy(svc.Name, svc.URL, fort.Local, fort.Gateway)

	var wsHandler http.Handler
	if svc.hasWS {
		wsURL := "ws" + svc.URL[4:]
		if !fort.Local {
			wsURL = "ws" + fort.Gateway[4:]
		}
		wsHandler = NewWSProxy(wsURL, svc.WSPaths, svc.Name, &ConnectionCallbacks{
			OnConnect:    tracker.OnConnect,
			OnDisconnect: tracker.OnDisconnect,
		})
	}

	mux.Handle(prefix, bffMiddleware(tc, proxy, wsHandler))
}
```

Update `bffMiddleware` to remove `domain.Service` parameter:

```go
func bffMiddleware(tc *TokenConverter, proxy http.Handler, wsHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wsHandler != nil && isWebSocketUpgrade(r) {
			if tc != nil {
				token, err := tc.Token(r)
				if err != nil {
					writeAuthError(w, err)
					return
				}
				r.Header.Set("Authorization", "Bearer "+token)
			}
			wsHandler.ServeHTTP(w, r)
			return
		}

		if tc == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "auth: no token converter configured"})
			return
		}

		token, err := tc.Token(r)
		if err != nil {
			writeAuthError(w, err)
			return
		}

		r.Header.Set("Authorization", "Bearer "+token)
		proxy.ServeHTTP(w, r)
	})
}
```

Update `servicesHandler`:

```go
func servicesHandler(fortName string, tracker *ServiceTracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"fort":      fortName,
			"services":  tracker.Services(),
			"conflicts": tracker.Conflicts(),
		})
	}
}
```

Update `configHandler`:

```go
func configHandler(fortName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"fort": fortName,
		})
	}
}
```

Remove `domain` import if no longer used (only `domain.Fort` is used in function parameters — keep import).

- [ ] **Step 2: Rewrite handler tests**

Update `handler_test.go` to create fake health servers, build trackers, and pass them to `NewHandler`:

```go
func newTestTracker(t *testing.T) (*httpapi.ServiceTracker, func()) {
	t.Helper()

	sharkfin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":   "ok",
			"name":     "sharkfin",
			"label":    "Chat",
			"route":    "/chat",
			"ws_paths": []string{"/ws", "/presence"},
		})
	}))

	nexus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"name":   "nexus",
			"label":  "Nexus",
			"route":  "/nexus",
		})
	}))

	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ui/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]any{
				"status": "unavailable",
				"name":   "auth",
				"label":  "Auth",
				"route":  "/auth",
			})
			return
		}
		// Auth service proxy handler for BFF tests.
		w.WriteHeader(http.StatusOK)
	}))

	tracker := httpapi.NewServiceTracker([]string{sharkfin.URL, nexus.URL, auth.URL})
	tracker.InitialProbe(context.Background())

	cleanup := func() {
		sharkfin.Close()
		nexus.Close()
		auth.Close()
	}

	return tracker, cleanup
}

func newTestFort(tracker *httpapi.ServiceTracker) domain.Fort {
	svcs := tracker.Services()
	configSvcs := make([]domain.ConfigService, len(svcs))
	for i, s := range svcs {
		configSvcs[i] = domain.ConfigService{URL: s.URL}
	}
	return domain.Fort{
		Name:     "local",
		Local:    true,
		Services: configSvcs,
	}
}

func TestHandler_ServicesEndpoint(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()
	fort := newTestFort(tracker)

	handler := httpapi.NewHandler(fort, tracker, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Fort     string `json:"fort"`
		Services []struct {
			Name      string `json:"name"`
			Connected bool   `json:"connected"`
			UI        bool   `json:"ui"`
		} `json:"services"`
		Conflicts []any `json:"conflicts"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Fort != "local" {
		t.Fatalf("expected fort 'local', got %q", resp.Fort)
	}
	if len(resp.Services) != 3 {
		t.Fatalf("expected 3 services, got %d", len(resp.Services))
	}
}
```

Add `"context"` to imports. Update all other handler tests similarly.

- [ ] **Step 3: Update cmd/web/web.go**

In `cmd/web/web.go`, update the `run` function:

```go
func run(cmd *cobra.Command, args []string) error {
	registry := fortconfig.New()
	fort := registry.Active()

	log.Info("starting web server",
		"fort", fort.Name,
		"local", fort.Local,
		"services", len(fort.Services),
	)

	// Create service tracker and run initial probe.
	urls := make([]string, len(fort.Services))
	for i, svc := range fort.Services {
		urls[i] = svc.URL
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	tracker := httpapi.NewServiceTracker(urls)
	tracker.InitialProbe(ctx)
	tracker.StartPolling(ctx, 10*time.Second)

	// Find the auth service URL from tracker (discovered via /ui/health).
	var tc *httpapi.TokenConverter
	if authSvc, ok := tracker.ServiceByName("auth"); ok {
		tc = httpapi.NewTokenConverter(authSvc.URL)
	} else {
		log.Warn("auth service not discovered — BFF token conversion disabled")
	}

	// SPA handler.
	var spaFS fs.FS
	if !dev {
		sub, err := fs.Sub(webFS, "dist")
		if err != nil {
			return fmt.Errorf("embedded SPA: %w", err)
		}
		spaFS = sub
	}

	handler := httpapi.NewHandler(fort, tracker, tc, spaFS)

	// ...rest unchanged (dev proxy, server setup, etc.)
```

Move the `signal.NotifyContext` call earlier (before tracker creation) so the context is available for probing. The rest of the function (dev proxy, server, shutdown) stays the same.

- [ ] **Step 4: Run all Go tests**

Run: `go test ./... -v`
Expected: All PASS.

- [ ] **Step 5: Build**

Run: `mise run build`
Expected: Build succeeds (note: needs `mise run build:web` first, which is chained).

- [ ] **Step 6: Commit**

```bash
git add internal/infra/httpapi/handler.go internal/infra/httpapi/handler_test.go cmd/web/web.go
git commit -m "feat(handler): wire ServiceTracker, delete hardcoded serviceMetadata, dynamic route registration"
```

---

## Chunk 3: Design Tokens

### Task 6: Add status design tokens

**Files:**
- Modify: `web/packages/ui/src/styles/tokens.css`
- Modify: `web/packages/ui/src/styles/components.css`

- [ ] **Step 1: Add status tokens to tokens.css**

In the `:root` block, append after the existing tokens:

```css
  --wf-error: #ef4444;
  --wf-error-subtle: rgba(239, 68, 68, 0.12);
  --wf-warning: #f59e0b;
  --wf-warning-subtle: rgba(245, 158, 11, 0.12);
  --wf-success: #22c55e;
  --wf-success-subtle: rgba(34, 197, 94, 0.12);
```

In the `[data-theme="light"]` block, append:

```css
  --wf-error: #dc2626;
  --wf-error-subtle: rgba(220, 38, 38, 0.08);
  --wf-warning: #d97706;
  --wf-warning-subtle: rgba(217, 119, 6, 0.08);
  --wf-success: #16a34a;
  --wf-success-subtle: rgba(22, 163, 74, 0.08);
```

- [ ] **Step 2: Update status-dot colors to use semantic tokens**

In `components.css`, replace the hardcoded status-dot colors:

```css
.wf-status-dot--online { background: var(--wf-success); }
.wf-status-dot--away { background: var(--wf-warning); }
.wf-status-dot--offline { background: var(--wf-error); }
```

- [ ] **Step 3: Commit**

```bash
git add web/packages/ui/src/styles/tokens.css web/packages/ui/src/styles/components.css
git commit -m "feat(ui): add semantic status tokens, update status-dot colors"
```

---

## Chunk 4: `wf-banner` Component

### Task 7: Implement wf-banner web component

**Files:**
- Create: `web/packages/ui/src/components/banner.ts`
- Create: `web/packages/ui/src/styles/banner.css`
- Modify: `web/packages/ui/src/styles/components.css`
- Modify: `web/packages/ui/src/index.ts`

**Design note:** Existing `@workfort/ui` components use light DOM (no Shadow DOM) and imperative DOM construction — no `render()` method, no `<slot>` elements. `wf-banner` follows this pattern. Content is set via properties, not slots.

- [ ] **Step 1: Write the banner component**

In `web/packages/ui/src/components/banner.ts`:

```typescript
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfBanner extends WfElement {
  @property({ type: String }) variant: 'error' | 'warning' | 'info' = 'info';
  @property({ type: Boolean }) dismissible = false;
  @property({ type: String }) headline = '';
  @property({ type: String }) details = '';

  private _expanded = false;
  private _headlineEl: HTMLSpanElement | null = null;
  private _detailsEl: HTMLDivElement | null = null;
  private _toggleBtn: HTMLButtonElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-banner');
    this._buildDOM();
    this._applyVariant();
    this._sync();
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('variant')) this._applyVariant();
    if (changed.has('headline') || changed.has('details') || changed.has('dismissible')) {
      this._sync();
    }
  }

  private _buildDOM(): void {
    const content = document.createElement('div');
    content.className = 'wf-banner__content';

    const icon = document.createElement('span');
    icon.className = 'wf-banner__icon';
    icon.setAttribute('aria-hidden', 'true');
    icon.textContent = '●';
    content.appendChild(icon);

    this._headlineEl = document.createElement('span');
    this._headlineEl.className = 'wf-banner__headline';
    content.appendChild(this._headlineEl);

    const actions = document.createElement('span');
    actions.className = 'wf-banner__actions';

    this._toggleBtn = document.createElement('button');
    this._toggleBtn.className = 'wf-banner__toggle';
    this._toggleBtn.setAttribute('aria-label', 'Toggle details');
    this._toggleBtn.textContent = '▾';
    this._toggleBtn.addEventListener('click', this._toggle);
    actions.appendChild(this._toggleBtn);

    if (this.dismissible) {
      const closeBtn = document.createElement('button');
      closeBtn.className = 'wf-banner__close';
      closeBtn.setAttribute('aria-label', 'Dismiss');
      closeBtn.textContent = '✕';
      closeBtn.addEventListener('click', this._dismiss);
      actions.appendChild(closeBtn);
    }

    content.appendChild(actions);
    this.appendChild(content);

    this._detailsEl = document.createElement('div');
    this._detailsEl.className = 'wf-banner__details';
    this._detailsEl.style.display = 'none';
    this.appendChild(this._detailsEl);
  }

  private _sync(): void {
    if (this._headlineEl) this._headlineEl.textContent = this.headline;
    if (this._detailsEl) this._detailsEl.textContent = this.details;
    if (this._toggleBtn) {
      this._toggleBtn.style.display = this.details ? '' : 'none';
    }
  }

  private _applyVariant(): void {
    this.classList.remove('wf-banner--error', 'wf-banner--warning', 'wf-banner--info');
    this.classList.add(`wf-banner--${this.variant}`);
  }

  private _toggle = (): void => {
    this._expanded = !this._expanded;
    if (this._detailsEl) {
      this._detailsEl.style.display = this._expanded ? '' : 'none';
    }
    if (this._toggleBtn) {
      this._toggleBtn.textContent = this._expanded ? '▴' : '▾';
    }
  };

  private _dismiss = (): void => {
    this.style.display = 'none';
    this.dispatchEvent(new CustomEvent('wf-dismiss', { bubbles: true, composed: true }));
  };

  show(): void {
    this.style.display = '';
  }
}

customElements.define('wf-banner', WfBanner);

declare global {
  interface HTMLElementTagNameMap {
    'wf-banner': WfBanner;
  }
}
```

- [ ] **Step 2: Write the banner styles**

In `web/packages/ui/src/styles/banner.css`:

```css
.wf-banner {
  display: block;
  padding: var(--wf-space-sm) var(--wf-space-lg);
  border-left: 4px solid var(--wf-border);
  font-family: var(--wf-font-sans);
  font-size: var(--wf-font-size-sm);
}

.wf-banner--error {
  background: var(--wf-error-subtle);
  border-left-color: var(--wf-error);
}
.wf-banner--warning {
  background: var(--wf-warning-subtle);
  border-left-color: var(--wf-warning);
}
.wf-banner--info {
  background: var(--wf-bg-secondary);
  border-left-color: var(--wf-accent);
}

.wf-banner__content {
  display: flex;
  align-items: center;
  gap: var(--wf-space-sm);
}

.wf-banner__icon {
  flex-shrink: 0;
  font-size: var(--wf-font-size-xs);
}
.wf-banner--error .wf-banner__icon { color: var(--wf-error); }
.wf-banner--warning .wf-banner__icon { color: var(--wf-warning); }
.wf-banner--info .wf-banner__icon { color: var(--wf-accent); }

.wf-banner__headline {
  flex: 1;
  font-weight: 600;
  color: var(--wf-text);
}

.wf-banner__actions {
  display: flex;
  align-items: center;
  gap: var(--wf-space-xs);
}

.wf-banner__toggle,
.wf-banner__close {
  background: none;
  border: none;
  color: var(--wf-text-secondary);
  cursor: pointer;
  padding: var(--wf-space-xs);
  font-size: var(--wf-font-size-sm);
  line-height: 1;
}
.wf-banner__toggle:hover,
.wf-banner__close:hover {
  color: var(--wf-text);
}

.wf-banner__details {
  margin-top: var(--wf-space-sm);
  padding-left: calc(var(--wf-space-sm) + 0.5rem + var(--wf-space-sm));
  font-family: var(--wf-font-mono);
  font-size: var(--wf-font-size-xs);
  color: var(--wf-text-secondary);
  white-space: pre-wrap;
}
```

- [ ] **Step 3: Import banner styles in components.css**

Append to `web/packages/ui/src/styles/components.css`:

```css
@import './banner.css';
```

- [ ] **Step 4: Register in index.ts**

Add to `web/packages/ui/src/index.ts`:

```typescript
import './components/banner.js';
export { WfBanner } from './components/banner.js';
```

- [ ] **Step 5: Verify the build**

Run: `cd web && pnpm --filter @workfort/ui build`
Expected: Build succeeds.

- [ ] **Step 6: Commit**

```bash
git add web/packages/ui/src/components/banner.ts web/packages/ui/src/styles/banner.css web/packages/ui/src/styles/components.css web/packages/ui/src/index.ts
git commit -m "feat(ui): add wf-banner component"
```

---

## Chunk 5: `wf-toast` Component

### Task 8: Implement wf-toast and wf-toast-container

**Files:**
- Create: `web/packages/ui/src/components/toast.ts`
- Create: `web/packages/ui/src/components/toast-container.ts`
- Create: `web/packages/ui/src/styles/toast.css`
- Modify: `web/packages/ui/src/styles/components.css`
- Modify: `web/packages/ui/src/index.ts`

- [ ] **Step 1: Write the toast component**

In `web/packages/ui/src/components/toast.ts` (imperative DOM, same pattern as existing components):

```typescript
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfToast extends WfElement {
  @property({ type: String }) variant: 'error' | 'warning' | 'info' | 'success' = 'info';
  @property({ type: Boolean }) sticky = false;
  @property({ type: Number }) duration = 5000;

  private _timer: ReturnType<typeof setTimeout> | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-toast');
    this._applyVariant();
    this._ensureCloseButton();

    if (!this.sticky) {
      this._timer = setTimeout(() => this._dismiss(), this.duration);
    }
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    if (this._timer) {
      clearTimeout(this._timer);
      this._timer = null;
    }
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('variant')) this._applyVariant();
  }

  private _applyVariant(): void {
    this.classList.remove('wf-toast--error', 'wf-toast--warning', 'wf-toast--info', 'wf-toast--success');
    this.classList.add(`wf-toast--${this.variant}`);
  }

  private _ensureCloseButton(): void {
    if (this.querySelector('.wf-toast__close')) return;
    const btn = document.createElement('button');
    btn.className = 'wf-toast__close';
    btn.setAttribute('aria-label', 'Dismiss');
    btn.textContent = '✕';
    btn.addEventListener('click', () => this._dismiss());
    this.appendChild(btn);
  }

  private _dismiss(): void {
    this.dispatchEvent(new CustomEvent('wf-dismiss', { bubbles: true, composed: true }));
    this.remove();
  }
}

customElements.define('wf-toast', WfToast);

declare global {
  interface HTMLElementTagNameMap {
    'wf-toast': WfToast;
  }
}
```

- [ ] **Step 2: Write the toast container component**

In `web/packages/ui/src/components/toast-container.ts`:

```typescript
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfToastContainer extends WfElement {
  @property({ type: String }) position: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left' = 'top-right';

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-toast-container');
    this._applyPosition();
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('position')) this._applyPosition();
  }

  private _applyPosition(): void {
    this.classList.remove(
      'wf-toast-container--top-right',
      'wf-toast-container--top-left',
      'wf-toast-container--bottom-right',
      'wf-toast-container--bottom-left',
    );
    this.classList.add(`wf-toast-container--${this.position}`);
  }
}

customElements.define('wf-toast-container', WfToastContainer);

declare global {
  interface HTMLElementTagNameMap {
    'wf-toast-container': WfToastContainer;
  }
}
```

- [ ] **Step 3: Write the toast styles**

In `web/packages/ui/src/styles/toast.css`:

```css
.wf-toast {
  display: flex;
  align-items: center;
  gap: var(--wf-space-sm);
  padding: var(--wf-space-sm) var(--wf-space-md);
  border-left: 4px solid var(--wf-border);
  border-radius: var(--wf-radius-md);
  background: var(--wf-bg);
  color: var(--wf-text);
  font-family: var(--wf-font-sans);
  font-size: var(--wf-font-size-sm);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 240px;
  max-width: 400px;
  animation: wf-toast-in 200ms ease-out;
}

@keyframes wf-toast-in {
  from { opacity: 0; transform: translateX(16px); }
  to { opacity: 1; transform: translateX(0); }
}

.wf-toast--error { border-left-color: var(--wf-error); background: var(--wf-error-subtle); }
.wf-toast--warning { border-left-color: var(--wf-warning); background: var(--wf-warning-subtle); }
.wf-toast--info { border-left-color: var(--wf-accent); }
.wf-toast--success { border-left-color: var(--wf-success); background: var(--wf-success-subtle); }

.wf-toast__close {
  background: none;
  border: none;
  color: var(--wf-text-secondary);
  cursor: pointer;
  padding: var(--wf-space-xs);
  font-size: var(--wf-font-size-sm);
  line-height: 1;
  margin-left: auto;
}
.wf-toast__close:hover { color: var(--wf-text); }

.wf-toast-container {
  position: fixed;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: var(--wf-space-sm);
  pointer-events: none;
}
.wf-toast-container > * { pointer-events: auto; }

.wf-toast-container--top-right { top: var(--wf-space-lg); right: var(--wf-space-lg); }
.wf-toast-container--top-left { top: var(--wf-space-lg); left: var(--wf-space-lg); }
.wf-toast-container--bottom-right { bottom: var(--wf-space-lg); right: var(--wf-space-lg); }
.wf-toast-container--bottom-left { bottom: var(--wf-space-lg); left: var(--wf-space-lg); }
```

- [ ] **Step 4: Import toast styles and register components**

In `components.css`, append:

```css
@import './toast.css';
```

In `index.ts`, add:

```typescript
import './components/toast.js';
import './components/toast-container.js';
export { WfToast } from './components/toast.js';
export { WfToastContainer } from './components/toast-container.js';
```

- [ ] **Step 5: Verify the build**

Run: `cd web && pnpm --filter @workfort/ui build`
Expected: Build succeeds.

- [ ] **Step 6: Commit**

```bash
git add web/packages/ui/src/components/toast.ts web/packages/ui/src/components/toast-container.ts web/packages/ui/src/styles/toast.css web/packages/ui/src/styles/components.css web/packages/ui/src/index.ts
git commit -m "feat(ui): add wf-toast and wf-toast-container components"
```

---

## Chunk 6: Shell SPA Stores & Types

### Task 9: Update API types

**Files:**
- Modify: `web/shell/src/lib/api.ts`

- [ ] **Step 1: Add connected, conflicts to types**

```typescript
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
```

- [ ] **Step 2: Commit**

```bash
git add web/shell/src/lib/api.ts
git commit -m "feat(shell): add connected and conflicts to API types"
```

### Task 10: Update remotes.ts

**Files:**
- Modify: `web/shell/src/lib/remotes.ts`

- [ ] **Step 1: Update ServiceModule interface and make registration incremental**

```typescript
import { init, registerRemotes, loadRemote } from '@module-federation/runtime';
import type { ServiceInfo } from './api';

// Bootstrap MF runtime once.
init({ name: 'shell', remotes: [] });

const registeredNames = new Set<string>();

export function registerNewRemotes(services: ServiceInfo[]): void {
  const newRemotes = services
    .filter((s) => s.enabled && s.ui && !registeredNames.has(s.name))
    .map((s) => ({
      name: s.name,
      entry: `/api/${s.name}/ui/remoteEntry.js`,
    }));

  if (newRemotes.length > 0) {
    registerRemotes(newRemotes);
    newRemotes.forEach((r) => registeredNames.add(r.name));
  }
}

export interface ServiceModule {
  default: (props: { connected: boolean }) => any;
  manifest: { name: string; label: string; route: string; minWidth?: number };
  SidebarContent?: () => any;
  HeaderActions?: () => any;
}

export async function loadServiceModule(
  serviceName: string,
): Promise<ServiceModule> {
  const mod = await loadRemote<ServiceModule>(`${serviceName}/index`);
  if (!mod || !mod.default || !mod.manifest) {
    throw new Error(
      `Remote "${serviceName}" did not export required fields (default, manifest)`,
    );
  }
  return mod;
}
```

Remove the old `initRemotes` function and `initialized` flag.

- [ ] **Step 2: Commit**

```bash
git add web/shell/src/lib/remotes.ts
git commit -m "feat(shell): incremental remote registration, connected prop on ServiceModule"
```

### Task 11: Implement banner store

**Files:**
- Create: `web/shell/src/stores/banners.ts`

- [ ] **Step 1: Write the banner store**

```typescript
import { createSignal } from 'solid-js';

export interface BannerEntry {
  key: string;
  variant: 'error' | 'warning' | 'info';
  headline: string;
  details?: string;
  source: 'system' | 'app';
}

const [banners, setBanners] = createSignal<BannerEntry[]>([]);
const dismissed = new Set<string>();

export { banners };

export function addBanner(
  key: string,
  variant: BannerEntry['variant'],
  headline: string,
  details?: string,
  source: BannerEntry['source'] = 'app',
): void {
  setBanners((prev) => {
    if (prev.find((b) => b.key === key)) return prev;
    dismissed.delete(key);
    return [...prev, { key, variant, headline, details, source }];
  });
}

export function removeBanner(key: string): void {
  setBanners((prev) => prev.filter((b) => b.key !== key));
}

export function dismissBanner(key: string): void {
  dismissed.add(key);
}

export function isBannerDismissed(key: string): boolean {
  return dismissed.has(key);
}

/** System banners first, then app banners. Errors before warnings. */
export function sortedBanners(): BannerEntry[] {
  const variantOrder = { error: 0, warning: 1, info: 2 };
  return banners()
    .filter((b) => !dismissed.has(b.key))
    .sort((a, b) => {
      if (a.source !== b.source) return a.source === 'system' ? -1 : 1;
      return variantOrder[a.variant] - variantOrder[b.variant];
    });
}
```

- [ ] **Step 2: Commit**

```bash
git add web/shell/src/stores/banners.ts
git commit -m "feat(shell): add banner store"
```

### Task 12: Implement toast store

**Files:**
- Create: `web/shell/src/stores/toasts.ts`

- [ ] **Step 1: Write the toast store**

```typescript
import { createSignal } from 'solid-js';

export interface ToastEntry {
  id: string;
  variant: 'error' | 'warning' | 'info' | 'success';
  message: string;
  sticky: boolean;
  duration: number;
}

let nextId = 0;
const [toasts, setToasts] = createSignal<ToastEntry[]>([]);

export { toasts };

export interface ToastOptions {
  sticky?: boolean;
  duration?: number;
}

export function addToast(
  variant: ToastEntry['variant'],
  message: string,
  options: ToastOptions = {},
): string {
  const id = `toast-${++nextId}`;
  const entry: ToastEntry = {
    id,
    variant,
    message,
    sticky: options.sticky ?? false,
    duration: options.duration ?? 5000,
  };
  setToasts((prev) => [...prev, entry]);
  return id;
}

export function dismissToast(id: string): void {
  setToasts((prev) => prev.filter((t) => t.id !== id));
}
```

- [ ] **Step 2: Commit**

```bash
git add web/shell/src/stores/toasts.ts
git commit -m "feat(shell): add toast store"
```

### Task 13: Rewrite services store with polling

**Files:**
- Modify: `web/shell/src/stores/services.ts`

- [ ] **Step 1: Rewrite with polling loop and state transition detection**

```typescript
import { createSignal } from 'solid-js';
import { fetchServices, type ServiceInfo, type Conflict, type ServicesResponse } from '../lib/api';
import { registerNewRemotes } from '../lib/remotes';
import { addBanner, removeBanner, banners } from './banners';
import { addToast } from './toasts';

const POLL_INTERVAL = 30_000;

const [serviceList, setServiceList] = createSignal<ServiceInfo[]>([]);
const [conflictList, setConflictList] = createSignal<Conflict[]>([]);
const [fort, setFort] = createSignal('');

let prevConnected = new Map<string, boolean>();

function handlePollResult(res: ServicesResponse): void {
  setFort(res.fort);
  setConflictList(res.conflicts ?? []);

  // Detect state transitions for toasts.
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

  // Register any newly discovered remotes.
  registerNewRemotes(res.services);

  // Update conflict banners — add new, remove stale.
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
  // Remove conflict banners no longer present.
  for (const b of banners()) {
    if (b.key.startsWith('conflict:') && !activeConflictKeys.has(b.key)) {
      removeBanner(b.key);
    }
  }

  // Remove disconnected banners for services that are now connected.
  for (const svc of res.services) {
    if (svc.connected) {
      removeBanner(`disconnected:${svc.name}`);
    }
  }

  setServiceList(res.services);
}

let intervalId: ReturnType<typeof setInterval> | null = null;

export function startPolling(): void {
  fetchServices().then(handlePollResult).catch(console.error);
  intervalId = setInterval(() => {
    fetchServices().then(handlePollResult).catch(console.error);
  }, POLL_INTERVAL);
}

export function stopPolling(): void {
  if (intervalId) {
    clearInterval(intervalId);
    intervalId = null;
  }
}

export const services = serviceList;
export const conflicts = conflictList;
export const fortName = fort;
```

- [ ] **Step 2: Start/stop polling from the app root**

Find the app entry point (likely `App.tsx` or `index.tsx`) and call `startPolling()` on mount, `stopPolling()` on cleanup via `onMount`/`onCleanup`.

- [ ] **Step 3: Commit**

```bash
git add web/shell/src/stores/services.ts
git commit -m "feat(shell): services store with 30s polling, state transitions, stale banner cleanup"
```

---

## Chunk 7: Shell Layout & Nav Updates

### Task 14: Update shell layout for banners and toasts

**Files:**
- Modify: `web/shell/src/components/shell-layout.tsx`
- Modify: `web/shell/src/global.css`

- [ ] **Step 1: Update CSS grid**

In `global.css`, change `.shell-layout`:

```css
.shell-layout {
  display: grid;
  grid-template-rows: auto auto 1fr;
  grid-template-columns: 240px 1fr;
  grid-template-areas:
    "banners banners"
    "nav     nav"
    "sidebar content";
  height: 100%;
}

.shell-layout--no-sidebar {
  grid-template-columns: 1fr;
  grid-template-areas:
    "banners"
    "nav"
    "content";
}

.shell-banners {
  grid-area: banners;
}
```

- [ ] **Step 2: Render banners and toasts in shell-layout.tsx**

```tsx
import { type Component, type JSX, Show, For } from 'solid-js';
import NavBar from './nav-bar';
import { sortedBanners, dismissBanner } from '../stores/banners';
import { toasts, dismissToast } from '../stores/toasts';

const ShellLayout: Component<{
  sidebar?: () => JSX.Element;
  children: JSX.Element;
}> = (props) => {
  return (
    <div
      class="shell-layout"
      classList={{ 'shell-layout--no-sidebar': !props.sidebar }}
    >
      <div class="shell-banners">
        <For each={sortedBanners()}>
          {(banner) => (
            <wf-banner
              variant={banner.variant}
              headline={banner.headline}
              details={banner.details ?? ''}
              dismissible
              on:wf-dismiss={() => dismissBanner(banner.key)}
            />
          )}
        </For>
      </div>
      <NavBar />
      <Show when={props.sidebar}>
        <aside class="shell-sidebar">{props.sidebar!()}</aside>
      </Show>
      <main class="shell-content">{props.children}</main>
      <wf-toast-container position="top-right">
        <For each={toasts()}>
          {(toast) => (
            <wf-toast
              variant={toast.variant}
              sticky={toast.sticky}
              duration={toast.duration}
              on:wf-dismiss={() => dismissToast(toast.id)}
            >
              {toast.message}
            </wf-toast>
          )}
        </For>
      </wf-toast-container>
    </div>
  );
};

export default ShellLayout;
```

Note: `wf-banner` uses `headline` and `details` as properties (not slots) because the component uses light DOM with imperative DOM construction.

- [ ] **Step 3: Commit**

```bash
git add web/shell/src/components/shell-layout.tsx web/shell/src/global.css
git commit -m "feat(shell): banner bar above nav, toast container in layout"
```

### Task 15: Add status dots to nav tabs

**Files:**
- Modify: `web/shell/src/components/nav-bar.tsx`
- Modify: `web/shell/src/global.css`

- [ ] **Step 1: Add status dots**

Update `nav-bar.tsx`, adding `Show` to the solid-js import:

```tsx
import { For, Show, type Component } from 'solid-js';
```

Update the `For` loop:

```tsx
<For each={services().filter((s) => s.enabled)}>
  {(svc) => (
    <wf-list-item
      active={location.pathname.startsWith(svc.route)}
      class={!svc.ui ? 'shell-nav__tab--disabled' : ''}
      on:wf-select={() => navigate(svc.route)}
    >
      <Show when={svc.ui}>
        <wf-status-dot status={svc.connected ? 'online' : 'offline'} />
      </Show>
      {svc.label}
    </wf-list-item>
  )}
</For>
```

- [ ] **Step 2: Add spacing for status dot**

In `global.css`:

```css
.shell-nav__tabs wf-status-dot {
  margin-right: var(--wf-space-xs);
}
```

- [ ] **Step 3: Commit**

```bash
git add web/shell/src/components/nav-bar.tsx web/shell/src/global.css
git commit -m "feat(shell): status dots on nav tabs"
```

### Task 16: Pass connected prop to service-mount

**Files:**
- Modify: `web/shell/src/components/service-mount.tsx`

- [ ] **Step 1: Add connected prop with "never loaded + disconnected" fallback**

```tsx
import { createResource, Suspense, ErrorBoundary, Show, type Component } from 'solid-js';
import { Dynamic } from 'solid-js/web';
import { loadServiceModule, type ServiceModule } from '../lib/remotes';
import Unavailable from './unavailable';

const ServiceMount: Component<{
  name: string;
  label: string;
  connected: boolean;
  onModule?: (mod: ServiceModule | null) => void;
}> = (props) => {
  const [mod] = createResource(
    () => props.name,
    async (name) => {
      const m = await loadServiceModule(name);
      props.onModule?.(m);
      return m;
    },
  );

  return (
    <ErrorBoundary fallback={<Unavailable label={props.label} />}>
      <Suspense fallback={<wf-skeleton width="100%" height="200px" />}>
        <Show
          when={mod() || props.connected}
          fallback={
            <wf-banner
              variant="warning"
              headline={`${props.label} is starting up or temporarily unavailable. This page will update automatically when it's ready.`}
            />
          }
        >
          <Show when={mod()}>
            <Dynamic component={mod()!.default} connected={props.connected} />
          </Show>
        </Show>
      </Suspense>
    </ErrorBoundary>
  );
};

export default ServiceMount;
```

- [ ] **Step 2: Update all call sites to pass connected prop**

Find where `ServiceMount` is rendered (likely in the router setup) and pass `connected` from the services store lookup.

- [ ] **Step 3: Commit**

```bash
git add web/shell/src/components/service-mount.tsx
git commit -m "feat(shell): pass connected prop to service remotes, disconnected fallback banner"
```

### Task 17: Final verification

- [ ] **Step 1: Run all Go tests**

Run: `mise run test`
Expected: All PASS.

- [ ] **Step 2: Build frontend**

Run: `cd web && pnpm build`
Expected: Build succeeds.

- [ ] **Step 3: Full build**

Run: `mise run build`
Expected: Build succeeds.

- [ ] **Step 4: Manual smoke test**

Start the dev server and verify:
- `/api/services` returns `connected` and `conflicts` fields
- Nav tabs show status dots (green = connected, red = disconnected)
- Services with `ui: false` have no status dot and are grayed out
- Banners render above the nav bar when conflicts exist
- Banners show headline + expandable technical details
- Dismissing a banner hides it; it reappears if the condition changes
- Toast notifications appear on state transitions (service disconnect/reconnect)
- Toasts stack in top-right corner
- Sticky toasts persist until dismissed; non-sticky auto-dismiss after 5s
- A service that was down at startup appears when it comes up (next poll cycle)
