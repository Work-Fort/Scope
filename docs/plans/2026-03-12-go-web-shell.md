# Go Web Shell — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Go-side infrastructure for serving the WorkFort web UI — shared frontend serving package, BFF proxy, fort config, and the `workfort web` command.

**Architecture:** Bottom-up build order. `pkg/frontend/` (shared embed+serve) first, then domain types, fort config adapter, HTTP handler layer (proxy, BFF, SPA), composition root (`cmd/web/`), and finally CGo build gating. Each layer is independently testable.

**Tech Stack:** Go 1.25, `net/http`, `httputil.ReverseProxy`, `gorilla/websocket`, `lestrrat-go/jwx/v2`, `spf13/cobra`, `spf13/viper`

**Specs:**
- `docs/2026-03-12-go-web-shell-design.md` — primary spec
- `docs/2026-03-11-web-ui-design.md` — full web UI architecture
- `docs/2026-03-11-service-auth-design.md` — auth design (BFF flow)

---

## File Structure

**Intentional deviations from spec:**
- `spa.go` instead of `embed.go` — clearer name since `embed.go` now exists in `cmd/web/` for the embed directive
- `ws.go` split from `proxy.go` — WebSocket proxying is a distinct concern with different dependencies (gorilla/websocket)
- `NewServiceProxy(svc, local, gatewayURL)` exported with extra params — spec shows `newServiceProxy(service)` but the implementation needs `local` and `gatewayURL` to handle both routing modes, and it's exported because tests in `httpapi_test` call it directly

```
pkg/
  frontend/
    frontend.go            # Handler(fsys) — serves embedded Vite build with cache headers + health probe
    frontend_test.go       # Tests: cache headers, health probe, 404 handling

internal/
  domain/
    web.go                 # Fort, Service structs + FortRegistry interface

  infra/
    fortconfig/
      registry.go          # Viper-backed FortRegistry implementation
      registry_test.go     # Tests: reads Viper config, SetActive validation

    httpapi/
      handler.go           # NewHandler() — top-level mux wiring, /api/services, /api/config endpoints
      proxy.go             # NewServiceProxy() — reverse proxy construction, path stripping
      bff.go               # tokenConverter — cookie-to-JWT conversion with caching
      ws.go                # WebSocket upgrade proxying with path whitelist
      spa.go               # SPA serving — embedded file server with index.html fallback
      handler_test.go      # Integration: full mux routing
      proxy_test.go        # Tests: path stripping, disabled service 503, gateway forwarding
      bff_test.go          # Tests: token conversion, caching, error cases
      ws_test.go           # Tests: WebSocket whitelist, upgrade proxying
      spa_test.go          # Tests: static file serving, SPA fallback, dev mode proxy

cmd/
  web/
    web.go                 # New() *cobra.Command — composition root, flags, server lifecycle
  chat/
    chat.go                # Rename: NewChatCmd() → New()
  root.go                  # Remove chat import, add web import
  root_cgo.go              # //go:build cgo — imports and registers chat command
```

---

## Chunk 1: `pkg/frontend/` — Shared Frontend Serving

### Task 1: Health Probe and File Existence Check

**Files:**
- Create: `pkg/frontend/frontend.go`
- Create: `pkg/frontend/frontend_test.go`

**Context:** This package lets any Go service embed and serve its Vite-built Module Federation remote. A service calls `frontend.Handler(fsys)` with an `fs.FS` rooted at the build output directory. The handler serves files under `/ui/` with appropriate cache headers and provides a `/ui/health` endpoint that checks if `remoteEntry.js` exists.

- [ ] **Step 1: Write test for health probe — remoteEntry.js present**

```go
// pkg/frontend/frontend_test.go
package frontend_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/Work-Fort/WorkFort/pkg/frontend"
)

func TestHealthProbe_OK(t *testing.T) {
	fsys := fstest.MapFS{
		"remoteEntry.js": &fstest.MapFile{Data: []byte("// entry")},
	}

	handler := frontend.Handler(fsys)
	req := httptest.NewRequest(http.MethodGet, "/ui/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}

	body := rec.Body.String()
	if body != `{"status":"ok"}`+"\n" {
		t.Fatalf("unexpected body: %q", body)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/frontend/ -run TestHealthProbe_OK -v`
Expected: FAIL — package does not exist yet

- [ ] **Step 3: Write test for health probe — remoteEntry.js missing**

```go
// pkg/frontend/frontend_test.go (append)

func TestHealthProbe_Unavailable(t *testing.T) {
	fsys := fstest.MapFS{
		"other.js": &fstest.MapFile{Data: []byte("// other")},
	}

	handler := frontend.Handler(fsys)
	req := httptest.NewRequest(http.MethodGet, "/ui/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	body := rec.Body.String()
	if body != `{"status":"unavailable"}`+"\n" {
		t.Fatalf("unexpected body: %q", body)
	}
}
```

- [ ] **Step 4: Write minimal Handler implementation with health probe**

```go
// pkg/frontend/frontend.go
package frontend

import (
	"encoding/json"
	"io/fs"
	"net/http"
)

// Handler returns an http.Handler that serves an embedded Vite build
// as a Module Federation remote.
//
// Routes registered under /ui/:
//   /ui/health           — 200 if remoteEntry.js exists, 503 if not
//   /ui/assets/*         — immutable content-hashed chunks (1yr cache)
//   /ui/remoteEntry.js   — federation entry point (no-cache)
//   /ui/*                — everything else (no-cache)
//
// The fsys must be rooted at the Vite build output directory
// (e.g., the result of fs.Sub(embedFS, "web/dist")).
func Handler(fsys fs.FS) http.Handler {
	hasRemoteEntry := fileExists(fsys, "remoteEntry.js")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ui/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if hasRemoteEntry {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
		}
	})

	// File serving will be added in Task 2.
	return mux
}

func fileExists(fsys fs.FS, name string) bool {
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	f.Close()
	return true
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./pkg/frontend/ -v`
Expected: PASS — both health probe tests pass

- [ ] **Step 6: Commit**

```bash
git add pkg/frontend/frontend.go pkg/frontend/frontend_test.go
git commit -m "feat(frontend): add health probe for Module Federation remote"
```

### Task 2: Static File Serving with Cache Headers

**Files:**
- Modify: `pkg/frontend/frontend.go`
- Modify: `pkg/frontend/frontend_test.go`

**Context:** Vite outputs content-hashed files under `assets/` (safe to cache forever) and `remoteEntry.js` (must revalidate). The handler sets appropriate `Cache-Control` headers based on path pattern.

- [ ] **Step 1: Write test for immutable cache on assets/**

```go
// pkg/frontend/frontend_test.go (append)

func TestCacheHeaders_Assets(t *testing.T) {
	fsys := fstest.MapFS{
		"remoteEntry.js":       &fstest.MapFile{Data: []byte("// entry")},
		"assets/chunk-abc.js":  &fstest.MapFile{Data: []byte("// chunk")},
	}

	handler := frontend.Handler(fsys)
	req := httptest.NewRequest(http.MethodGet, "/ui/assets/chunk-abc.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	cc := rec.Header().Get("Cache-Control")
	expected := "public, max-age=31536000, immutable"
	if cc != expected {
		t.Fatalf("expected Cache-Control %q, got %q", expected, cc)
	}
}
```

- [ ] **Step 2: Write test for no-cache on remoteEntry.js**

```go
// pkg/frontend/frontend_test.go (append)

func TestCacheHeaders_RemoteEntry(t *testing.T) {
	fsys := fstest.MapFS{
		"remoteEntry.js": &fstest.MapFile{Data: []byte("// entry")},
	}

	handler := frontend.Handler(fsys)
	req := httptest.NewRequest(http.MethodGet, "/ui/remoteEntry.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	cc := rec.Header().Get("Cache-Control")
	if cc != "no-cache" {
		t.Fatalf("expected Cache-Control no-cache, got %q", cc)
	}
}
```

- [ ] **Step 3: Write test for no-cache on other files**

```go
// pkg/frontend/frontend_test.go (append)

func TestCacheHeaders_OtherFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"remoteEntry.js": &fstest.MapFile{Data: []byte("// entry")},
		"manifest.json":  &fstest.MapFile{Data: []byte(`{}`)},
	}

	handler := frontend.Handler(fsys)
	req := httptest.NewRequest(http.MethodGet, "/ui/manifest.json", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	cc := rec.Header().Get("Cache-Control")
	if cc != "no-cache" {
		t.Fatalf("expected Cache-Control no-cache, got %q", cc)
	}
}
```

- [ ] **Step 4: Write test for 404 on missing file**

```go
// pkg/frontend/frontend_test.go (append)

func TestFileNotFound(t *testing.T) {
	fsys := fstest.MapFS{
		"remoteEntry.js": &fstest.MapFile{Data: []byte("// entry")},
	}

	handler := frontend.Handler(fsys)
	req := httptest.NewRequest(http.MethodGet, "/ui/nonexistent.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	// Go 1.23+ strips Cache-Control on error responses. Verify it's absent
	// so 404s are never cached by browsers or CDNs.
	cc := rec.Header().Get("Cache-Control")
	if cc != "" {
		t.Fatalf("expected no Cache-Control on 404, got %q", cc)
	}
}
```

- [ ] **Step 5: Run tests to verify they fail**

Run: `go test ./pkg/frontend/ -v`
Expected: FAIL — file serving not implemented yet (health probe tests still pass)

- [ ] **Step 6: Implement file serving with cache headers**

Replace the `Handler` function in `pkg/frontend/frontend.go`:

```go
// pkg/frontend/frontend.go
package frontend

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"
)

// Handler returns an http.Handler that serves an embedded Vite build
// as a Module Federation remote.
//
// Routes registered under /ui/:
//   /ui/health           — 200 if remoteEntry.js exists, 503 if not
//   /ui/assets/*         — immutable content-hashed chunks (1yr cache)
//   /ui/remoteEntry.js   — federation entry point (no-cache)
//   /ui/*                — everything else (no-cache)
//
// The fsys must be rooted at the Vite build output directory
// (e.g., the result of fs.Sub(embedFS, "web/dist")).
func Handler(fsys fs.FS) http.Handler {
	hasRemoteEntry := fileExists(fsys, "remoteEntry.js")
	fileServer := http.StripPrefix("/ui/", http.FileServer(http.FS(fsys)))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ui/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if hasRemoteEntry {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
		}
	})

	mux.HandleFunc("/ui/", func(w http.ResponseWriter, r *http.Request) {
		// Set cache headers based on path pattern BEFORE calling the file server.
		// http.FileServer does not set or overwrite Cache-Control on success.
		// On error (404, 416), Go 1.23+ strips Cache-Control, which is correct —
		// we don't want to cache error responses.
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

func fileExists(fsys fs.FS, name string) bool {
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	f.Close()
	return true
}
```

- [ ] **Step 7: Run all tests to verify they pass**

Run: `go test ./pkg/frontend/ -v`
Expected: PASS — all 6 tests pass

- [ ] **Step 8: Commit**

```bash
git add pkg/frontend/frontend.go pkg/frontend/frontend_test.go
git commit -m "feat(frontend): add static file serving with cache headers"
```

---

## Chunk 2: Domain Layer and Fort Config

### Task 3: Domain Types

**Files:**
- Create: `internal/domain/web.go`

**Context:** Pure types and a port interface for fort configuration. No dependencies, no tests needed — these are plain data structs and an interface.

- [ ] **Step 1: Create domain types**

```go
// internal/domain/web.go
package domain

// Fort is a named collection of services. Users can belong to multiple
// forts and switch between them.
type Fort struct {
	// Name is the fort identifier (e.g., "local", "acme-corp").
	Name string

	// Local controls how the proxy routes traffic.
	// true = proxy directly to each service's URL.
	// false = proxy through Gateway.
	Local bool

	// Gateway is the single origin URL for remote forts.
	// Only used when Local is false.
	Gateway string

	// Services lists the backend services in this fort.
	Services []Service
}

// Service is a backend service in a fort.
type Service struct {
	// Name is the service identifier (e.g., "auth", "sharkfin", "nexus", "hive").
	Name string

	// URL is the direct backend URL (e.g., "http://127.0.0.1:16000").
	// Only used when the fort's Local flag is true.
	URL string

	// WSPaths is a whitelist of paths that accept WebSocket upgrade.
	// Matched against the path suffix after the /api/{service} prefix is stripped.
	// Example: ["/ws", "/presence"]
	WSPaths []string

	// Enabled controls whether the proxy accepts requests for this service.
	// Disabled services return 503.
	Enabled bool
}

// FortRegistry reads fort configuration.
type FortRegistry interface {
	// Forts returns all configured forts.
	Forts() []Fort

	// Active returns the currently active fort.
	Active() Fort

	// SetActive switches the active fort.
	// Returns an error if the fort name does not exist.
	SetActive(name string) error
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/domain/`
Expected: Success, no errors

- [ ] **Step 3: Commit**

```bash
git add internal/domain/web.go
git commit -m "feat(domain): add Fort, Service, and FortRegistry types"
```

### Task 4: Viper-backed Fort Registry

**Files:**
- Create: `internal/infra/fortconfig/registry.go`
- Create: `internal/infra/fortconfig/registry_test.go`
- Modify: `pkg/config/viper.go` (add fort defaults)

**Context:** The registry reads fort config from Viper. Viper is already initialized in `pkg/config/viper.go:InitViper()`. The registry walks `forts.<name>.services.<svc>` keys. The config YAML format is defined in the spec (see `docs/2026-03-12-go-web-shell-design.md`, "Config file format" section).

**Reference:** Read `pkg/config/viper.go` (lines 14-32) for how defaults are set.

- [ ] **Step 1: Add fort config defaults to `pkg/config/viper.go`**

Add these lines inside `InitViper()`, after the existing `viper.SetDefault` calls (after line 27):

```go
	// Fort defaults
	viper.SetDefault("active-fort", "local")
	viper.SetDefault("forts.local.local", true)
	viper.SetDefault("forts.local.services.auth.url", "http://127.0.0.1:3000")
	viper.SetDefault("forts.local.services.auth.enabled", true)
	viper.SetDefault("forts.local.services.sharkfin.url", "http://127.0.0.1:16000")
	viper.SetDefault("forts.local.services.sharkfin.enabled", true)
	viper.SetDefault("forts.local.services.sharkfin.ws-paths", []string{"/ws", "/presence"})
	viper.SetDefault("forts.local.services.nexus.url", "http://127.0.0.1:9600")
	viper.SetDefault("forts.local.services.nexus.enabled", true)
	viper.SetDefault("forts.local.services.hive.url", "http://127.0.0.1:17000")
	viper.SetDefault("forts.local.services.hive.enabled", false)
```

- [ ] **Step 2: Write test for reading the active fort**

```go
// internal/infra/fortconfig/registry_test.go
package fortconfig_test

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/Work-Fort/WorkFort/internal/domain"
	"github.com/Work-Fort/WorkFort/internal/infra/fortconfig"
)

func setupViper(t *testing.T) {
	t.Helper()
	viper.Reset()

	viper.Set("active-fort", "local")
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services.sharkfin.url", "http://127.0.0.1:16000")
	viper.Set("forts.local.services.sharkfin.enabled", true)
	viper.Set("forts.local.services.sharkfin.ws-paths", []string{"/ws", "/presence"})
	viper.Set("forts.local.services.nexus.url", "http://127.0.0.1:9600")
	viper.Set("forts.local.services.nexus.enabled", true)
	viper.Set("forts.local.services.auth.url", "http://127.0.0.1:3000")
	viper.Set("forts.local.services.auth.enabled", true)
}

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

	// Find sharkfin
	var sf *domain.Service
	for i := range fort.Services {
		if fort.Services[i].Name == "sharkfin" {
			sf = &fort.Services[i]
			break
		}
	}
	if sf == nil {
		t.Fatal("sharkfin service not found")
	}
	if sf.URL != "http://127.0.0.1:16000" {
		t.Fatalf("expected sharkfin URL http://127.0.0.1:16000, got %q", sf.URL)
	}
	if !sf.Enabled {
		t.Fatal("expected sharkfin to be enabled")
	}
	if len(sf.WSPaths) != 2 || sf.WSPaths[0] != "/ws" || sf.WSPaths[1] != "/presence" {
		t.Fatalf("unexpected ws-paths: %v", sf.WSPaths)
	}
}
```

- [ ] **Step 3: Write test for Forts() listing all forts**

```go
// internal/infra/fortconfig/registry_test.go (append)

func TestForts(t *testing.T) {
	viper.Reset()

	viper.Set("active-fort", "local")
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services.auth.url", "http://127.0.0.1:3000")
	viper.Set("forts.local.services.auth.enabled", true)

	viper.Set("forts.remote.local", false)
	viper.Set("forts.remote.gateway", "https://fort.acme.com")
	viper.Set("forts.remote.services.auth.enabled", true)

	reg := fortconfig.New()
	forts := reg.Forts()

	if len(forts) != 2 {
		t.Fatalf("expected 2 forts, got %d", len(forts))
	}

	// Find the remote fort
	var remote *domain.Fort
	for i := range forts {
		if forts[i].Name == "remote" {
			remote = &forts[i]
			break
		}
	}
	if remote == nil {
		t.Fatal("remote fort not found")
	}
	if remote.Local {
		t.Fatal("expected remote fort to not be local")
	}
	if remote.Gateway != "https://fort.acme.com" {
		t.Fatalf("expected gateway https://fort.acme.com, got %q", remote.Gateway)
	}
}
```

- [ ] **Step 4: Write test for SetActive — valid and invalid names**

```go
// internal/infra/fortconfig/registry_test.go (append)

func TestSetActive_Valid(t *testing.T) {
	viper.Reset()

	viper.Set("active-fort", "local")
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services.auth.url", "http://127.0.0.1:3000")
	viper.Set("forts.local.services.auth.enabled", true)

	viper.Set("forts.remote.local", false)
	viper.Set("forts.remote.gateway", "https://fort.acme.com")
	viper.Set("forts.remote.services.auth.enabled", true)

	reg := fortconfig.New()
	if err := reg.SetActive("remote"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fort := reg.Active()
	if fort.Name != "remote" {
		t.Fatalf("expected active fort 'remote', got %q", fort.Name)
	}
}

func TestSetActive_Invalid(t *testing.T) {
	viper.Reset()

	viper.Set("active-fort", "local")
	viper.Set("forts.local.local", true)
	viper.Set("forts.local.services.auth.url", "http://127.0.0.1:3000")
	viper.Set("forts.local.services.auth.enabled", true)

	reg := fortconfig.New()
	err := reg.SetActive("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent fort")
	}
}
```

- [ ] **Step 5: Run tests to verify they fail**

Run: `go test ./internal/infra/fortconfig/ -v`
Expected: FAIL — package does not exist yet

- [ ] **Step 6: Implement the registry**

```go
// internal/infra/fortconfig/registry.go
package fortconfig

import (
	"fmt"
	"sort"

	"github.com/spf13/viper"

	"github.com/Work-Fort/WorkFort/internal/domain"
)

// Compile-time interface check.
var _ domain.FortRegistry = (*Registry)(nil)

// Registry reads fort configuration from Viper.
type Registry struct{}

// New creates a new fort config registry.
func New() *Registry {
	return &Registry{}
}

// Forts returns all configured forts, sorted by name.
func (r *Registry) Forts() []domain.Fort {
	fortsMap := viper.GetStringMap("forts")
	forts := make([]domain.Fort, 0, len(fortsMap))
	for name := range fortsMap {
		forts = append(forts, r.readFort(name))
	}
	sort.Slice(forts, func(i, j int) bool {
		return forts[i].Name < forts[j].Name
	})
	return forts
}

// Active returns the currently active fort.
func (r *Registry) Active() domain.Fort {
	name := viper.GetString("active-fort")
	return r.readFort(name)
}

// SetActive switches the active fort. Returns an error if the fort does not exist.
func (r *Registry) SetActive(name string) error {
	fortsMap := viper.GetStringMap("forts")
	if _, ok := fortsMap[name]; !ok {
		return fmt.Errorf("fortconfig: fort %q not found", name)
	}
	viper.Set("active-fort", name)
	return nil
}

func (r *Registry) readFort(name string) domain.Fort {
	prefix := "forts." + name

	fort := domain.Fort{
		Name:    name,
		Local:   viper.GetBool(prefix + ".local"),
		Gateway: viper.GetString(prefix + ".gateway"),
	}

	svcsMap := viper.GetStringMap(prefix + ".services")
	for svcName := range svcsMap {
		svcPrefix := prefix + ".services." + svcName
		svc := domain.Service{
			Name:    svcName,
			URL:     viper.GetString(svcPrefix + ".url"),
			WSPaths: viper.GetStringSlice(svcPrefix + ".ws-paths"),
			Enabled: viper.GetBool(svcPrefix + ".enabled"),
		}
		fort.Services = append(fort.Services, svc)
	}

	sort.Slice(fort.Services, func(i, j int) bool {
		return fort.Services[i].Name < fort.Services[j].Name
	})

	return fort
}
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `go test ./internal/infra/fortconfig/ -v`
Expected: PASS — all 4 tests pass

- [ ] **Step 8: Commit**

```bash
git add pkg/config/viper.go internal/domain/web.go internal/infra/fortconfig/registry.go internal/infra/fortconfig/registry_test.go
git commit -m "feat(fortconfig): add domain types and Viper-backed fort registry"
```

---

## Chunk 3: BFF Token Conversion

### Task 5: Cookie-to-JWT Conversion with Caching

**Files:**
- Create: `internal/infra/httpapi/bff.go`
- Create: `internal/infra/httpapi/bff_test.go`

**Context:** The BFF converts session cookies to JWTs by calling the auth service's `GET /v1/token` endpoint. It caches the JWT keyed by session cookie value. JWTs have a 15-minute lifetime; the cache refreshes at 14 minutes. On 401 from auth, the cache entry is evicted and the session cookie is cleared.

The better-auth session cookie name is `better-auth.session_token`.

**Reference:** Read the "BFF token conversion" section of `docs/2026-03-12-go-web-shell-design.md`.

- [ ] **Step 1: Write test for successful token conversion**

```go
// internal/infra/httpapi/bff_test.go
package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Work-Fort/WorkFort/internal/infra/httpapi"
)

func TestTokenConverter_Success(t *testing.T) {
	// Mock auth service that returns a JWT when given a valid session cookie.
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("better-auth.session_token")
		if err != nil || cookie.Value != "valid-session" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "jwt-token-123"})
	}))
	defer authServer.Close()

	tc := httpapi.NewTokenConverter(authServer.URL)

	// Build a request with a session cookie.
	req := httptest.NewRequest(http.MethodGet, "/api/sharkfin/v1/channels", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "valid-session"})

	token, err := tc.Token(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "jwt-token-123" {
		t.Fatalf("expected jwt-token-123, got %q", token)
	}
}
```

- [ ] **Step 2: Write test for missing session cookie**

```go
// internal/infra/httpapi/bff_test.go (append)

func TestTokenConverter_NoCookie(t *testing.T) {
	tc := httpapi.NewTokenConverter("http://localhost:0")

	req := httptest.NewRequest(http.MethodGet, "/api/sharkfin/v1/channels", nil)
	_, err := tc.Token(req)
	if err == nil {
		t.Fatal("expected error for missing cookie")
	}
}
```

- [ ] **Step 3: Write test for auth service returning 401 (expired session) — verifies cache eviction**

This test uses `NewTokenConverterForTest` (a test-only constructor with zero cache TTL) to verify that:
1. A 401 from auth evicts the cache entry
2. The next call re-contacts auth instead of using a stale cached token

```go
// internal/infra/httpapi/bff_test.go (append)

func TestTokenConverter_ExpiredSession_EvictsCache(t *testing.T) {
	calls := 0
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch calls {
		case 1:
			// First: return valid token.
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "jwt-first"})
		case 2:
			// Second: session expired.
			w.WriteHeader(http.StatusUnauthorized)
		case 3:
			// Third: new token after re-auth.
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "jwt-refreshed"})
		}
	}))
	defer authServer.Close()

	// Zero TTL and zero refreshBefore means cached entries expire immediately,
	// so every Token() call hits auth. This lets us test the 401 eviction
	// path without time manipulation.
	tc := httpapi.NewTokenConverterForTest(authServer.URL, 0, 0)

	req := httptest.NewRequest(http.MethodGet, "/api/sharkfin/v1/channels", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "evict-test"})

	// Call 1 — succeeds, caches token.
	tok, err := tc.Token(req)
	if err != nil {
		t.Fatalf("call 1: %v", err)
	}
	if tok != "jwt-first" {
		t.Fatalf("call 1: expected jwt-first, got %q", tok)
	}

	// Call 2 — cache expired (TTL=0), auth returns 401, evicts cache entry.
	_, err = tc.Token(req)
	if err == nil {
		t.Fatal("call 2: expected error for expired session")
	}

	// Call 3 — cache was evicted, hits auth again, gets new token.
	tok, err = tc.Token(req)
	if err != nil {
		t.Fatalf("call 3: %v", err)
	}
	if tok != "jwt-refreshed" {
		t.Fatalf("call 3: expected jwt-refreshed, got %q", tok)
	}
	if calls != 3 {
		t.Fatalf("expected 3 auth calls, got %d", calls)
	}
}
```

- [ ] **Step 4: Write test for caching — second call does not hit auth service**

```go
// internal/infra/httpapi/bff_test.go (append)

func TestTokenConverter_CacheHit(t *testing.T) {
	calls := 0
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "jwt-cached"})
	}))
	defer authServer.Close()

	tc := httpapi.NewTokenConverter(authServer.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/sharkfin/v1/channels", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "cache-test"})

	// First call — hits auth service
	token1, err := tc.Token(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call — should use cache
	token2, err := tc.Token(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token1 != token2 {
		t.Fatalf("expected same token from cache")
	}
	if calls != 1 {
		t.Fatalf("expected 1 auth call (cached), got %d", calls)
	}
}
```

- [ ] **Step 5: Write test for auth service unreachable**

```go
// internal/infra/httpapi/bff_test.go (append)

func TestTokenConverter_AuthUnreachable(t *testing.T) {
	// Point at a URL that won't connect.
	tc := httpapi.NewTokenConverter("http://127.0.0.1:1")

	req := httptest.NewRequest(http.MethodGet, "/api/sharkfin/v1/channels", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "some-session"})

	_, err := tc.Token(req)
	if err == nil {
		t.Fatal("expected error for unreachable auth service")
	}
}
```

- [ ] **Step 6: Run tests to verify they fail**

Run: `go test ./internal/infra/httpapi/ -run TestTokenConverter -v`
Expected: FAIL — package does not exist yet

- [ ] **Step 7: Implement token converter**

Note: `TokenConverter` is exported from the start — `cmd/web/` references it directly.

```go
// internal/infra/httpapi/bff.go
package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	sessionCookieName = "better-auth.session_token"
	tokenLifetime     = 15 * time.Minute
	refreshBefore     = 1 * time.Minute
)

var (
	errNoSession      = errors.New("bff: no session cookie")
	errSessionExpired = errors.New("bff: session expired")
	errAuthDown       = errors.New("bff: auth service unavailable")
)

type cachedToken struct {
	jwt    string
	expiry time.Time
}

// TokenConverter converts session cookies to JWTs by calling the auth service.
type TokenConverter struct {
	authURL       string
	client        *http.Client
	tokenLifetime time.Duration
	refreshBefore time.Duration

	mu     sync.RWMutex
	tokens map[string]cachedToken
}

// NewTokenConverter creates a token converter that calls authServiceURL
// to exchange session cookies for JWTs.
func NewTokenConverter(authServiceURL string) *TokenConverter {
	return &TokenConverter{
		authURL:       authServiceURL + "/v1/token",
		client:        &http.Client{Timeout: 5 * time.Second},
		tokenLifetime: tokenLifetime,
		refreshBefore: refreshBefore,
		tokens:        make(map[string]cachedToken),
	}
}

// NewTokenConverterForTest creates a converter with custom timing parameters.
// Use ttl=0 and refresh=0 to force every call to hit auth (for testing
// cache eviction without time manipulation).
func NewTokenConverterForTest(authServiceURL string, ttl, refresh time.Duration) *TokenConverter {
	return &TokenConverter{
		authURL:       authServiceURL + "/v1/token",
		client:        &http.Client{Timeout: 5 * time.Second},
		tokenLifetime: ttl,
		refreshBefore: refresh,
		tokens:        make(map[string]cachedToken),
	}
}

// Token extracts the session cookie from the request and returns a JWT.
// Results are cached by session cookie value.
func (tc *TokenConverter) Token(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", errNoSession
	}
	sessionVal := cookie.Value

	// Check cache.
	tc.mu.RLock()
	cached, ok := tc.tokens[sessionVal]
	tc.mu.RUnlock()
	if ok && time.Until(cached.expiry) > tc.refreshBefore {
		return cached.jwt, nil
	}

	// Cache miss or near-expiry — call auth service.
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, tc.authURL, nil)
	if err != nil {
		return "", fmt.Errorf("bff: create request: %w", err)
	}
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionVal})

	resp, err := tc.client.Do(req)
	if err != nil {
		return "", errAuthDown
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Evict cache entry for this cookie value.
		tc.mu.Lock()
		delete(tc.tokens, sessionVal)
		tc.mu.Unlock()
		return "", errSessionExpired
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bff: auth returned status %d", resp.StatusCode)
	}

	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("bff: decode response: %w", err)
	}

	if body.Token == "" {
		return "", fmt.Errorf("bff: auth returned empty token")
	}

	// Cache the token.
	tc.mu.Lock()
	tc.tokens[sessionVal] = cachedToken{
		jwt:    body.Token,
		expiry: time.Now().Add(tc.tokenLifetime),
	}
	tc.mu.Unlock()

	return body.Token, nil
}
```

- [ ] **Step 8: Run tests to verify they pass**

Run: `go test ./internal/infra/httpapi/ -run TestTokenConverter -v`
Expected: PASS — all 5 tests pass

- [ ] **Step 9: Commit**

```bash
git add internal/infra/httpapi/bff.go internal/infra/httpapi/bff_test.go
git commit -m "feat(httpapi): add BFF token converter with session cookie caching"
```

---

## Chunk 4: Reverse Proxy and WebSocket Proxy

### Task 6: Reverse Proxy with Path Stripping

**Files:**
- Create: `internal/infra/httpapi/proxy.go`
- Create: `internal/infra/httpapi/proxy_test.go`

**Context:** Each enabled service gets a `httputil.ReverseProxy` that strips the `/api/{service}` prefix. For local forts, it forwards to the service's URL. For gateway forts, it preserves the `/api/{service}` prefix and forwards to the gateway. Disabled services return 503.

- [ ] **Step 1: Write test for path stripping on local fort**

```go
// internal/infra/httpapi/proxy_test.go
package httpapi_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Work-Fort/WorkFort/internal/domain"
	"github.com/Work-Fort/WorkFort/internal/infra/httpapi"
)

func TestProxy_PathStripping(t *testing.T) {
	// Backend service that echoes the request path.
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer backend.Close()

	svc := domain.Service{
		Name:    "nexus",
		URL:     backend.URL,
		Enabled: true,
	}

	proxy := httpapi.NewServiceProxy(svc, true, "")

	req := httptest.NewRequest(http.MethodGet, "/api/nexus/v1/vms", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	body, _ := io.ReadAll(rec.Body)
	if string(body) != "/v1/vms" {
		t.Fatalf("expected path /v1/vms, got %q", string(body))
	}
}
```

- [ ] **Step 2: Write test for disabled service returning 503**

```go
// internal/infra/httpapi/proxy_test.go (append)

func TestProxy_DisabledService(t *testing.T) {
	svc := domain.Service{
		Name:    "hive",
		Enabled: false,
	}

	proxy := httpapi.NewServiceProxy(svc, true, "")

	req := httptest.NewRequest(http.MethodGet, "/api/hive/v1/teams", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}
```

- [ ] **Step 3: Write test for gateway fort — path preserved**

```go
// internal/infra/httpapi/proxy_test.go (append)

func TestProxy_GatewayFort(t *testing.T) {
	// Gateway backend that echoes the request path.
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer gateway.Close()

	svc := domain.Service{
		Name:    "nexus",
		Enabled: true,
	}

	proxy := httpapi.NewServiceProxy(svc, false, gateway.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/nexus/v1/vms", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	body, _ := io.ReadAll(rec.Body)
	if string(body) != "/api/nexus/v1/vms" {
		t.Fatalf("expected path /api/nexus/v1/vms, got %q", string(body))
	}
}
```

- [ ] **Step 4: Run tests to verify they fail**

Run: `go test ./internal/infra/httpapi/ -run TestProxy -v`
Expected: FAIL — proxy.go does not exist yet

- [ ] **Step 5: Implement the reverse proxy**

```go
// internal/infra/httpapi/proxy.go
package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/Work-Fort/WorkFort/internal/domain"
)

// NewServiceProxy creates an http.Handler that proxies requests to a service.
//
// For local forts (local=true): strips the /api/{service} prefix and forwards
// to the service's URL (e.g., /api/nexus/v1/vms → http://service/v1/vms).
//
// For gateway forts (local=false): preserves the /api/{service} prefix and
// forwards to the gateway URL.
//
// Disabled services return 503.
func NewServiceProxy(svc domain.Service, local bool, gatewayURL string) http.Handler {
	if !svc.Enabled {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": svc.Name + " service is disabled",
			})
		})
	}

	prefix := "/api/" + svc.Name

	if local {
		target, _ := url.Parse(svc.URL)
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

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./internal/infra/httpapi/ -run TestProxy -v`
Expected: PASS — all 3 tests pass

- [ ] **Step 7: Commit**

```bash
git add internal/infra/httpapi/proxy.go internal/infra/httpapi/proxy_test.go
git commit -m "feat(httpapi): add reverse proxy with path stripping and gateway support"
```

### Task 7: WebSocket Proxy with Path Whitelist

**Files:**
- Create: `internal/infra/httpapi/ws.go`
- Create: `internal/infra/httpapi/ws_test.go`

**Context:** WebSocket upgrade requests are only accepted for paths in the service's `WSPaths` whitelist. The path is matched after stripping `/api/{service}`. Non-whitelisted upgrade requests return 400. The BFF token converter provides the JWT for the forwarded upgrade handshake.

**Reference:** `gorilla/websocket` is already in go.mod.

- [ ] **Step 1: Write test for whitelisted path upgrade**

```go
// internal/infra/httpapi/ws_test.go
package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/Work-Fort/WorkFort/internal/infra/httpapi"
)

func TestWSProxy_WhitelistedPath(t *testing.T) {
	// Backend WS server that accepts upgrades.
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("backend upgrade error: %v", err)
			return
		}
		defer conn.Close()
		// Echo one message back.
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		_ = conn.WriteMessage(mt, msg)
	}))
	defer backend.Close()

	backendURL := "ws" + strings.TrimPrefix(backend.URL, "http")
	wsHandler := httpapi.NewWSProxy(backendURL, []string{"/ws", "/presence"}, "nexus")

	// Wrap in a test server so we can dial it.
	proxy := httptest.NewServer(wsHandler)
	defer proxy.Close()

	proxyURL := "ws" + strings.TrimPrefix(proxy.URL, "http") + "/api/nexus/ws"
	conn, resp, err := websocket.DefaultDialer.Dial(proxyURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v (resp: %v)", err, resp)
	}
	defer conn.Close()

	// Send and receive a message.
	if err := conn.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		t.Fatalf("write error: %v", err)
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if string(msg) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(msg))
	}
}
```

- [ ] **Step 2: Write test for non-whitelisted path returning 400**

```go
// internal/infra/httpapi/ws_test.go (append)

func TestWSProxy_NonWhitelistedPath(t *testing.T) {
	wsHandler := httpapi.NewWSProxy("ws://localhost:0", []string{"/ws"}, "nexus")
	proxy := httptest.NewServer(wsHandler)
	defer proxy.Close()

	proxyURL := "ws" + strings.TrimPrefix(proxy.URL, "http") + "/api/nexus/not-allowed"
	_, resp, err := websocket.DefaultDialer.Dial(proxyURL, nil)
	if err == nil {
		t.Fatal("expected error for non-whitelisted path")
	}
	if resp != nil && resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/infra/httpapi/ -run TestWSProxy -v`
Expected: FAIL — ws.go does not exist yet

- [ ] **Step 4: Implement WebSocket proxy**

```go
// internal/infra/httpapi/ws.go
package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// NewWSProxy creates an http.Handler that proxies WebSocket connections
// to a backend service. Only paths in wsPaths are allowed — others return 400.
//
// The serviceName is used to strip the /api/{service} prefix before matching.
// The backendURL is the base WebSocket URL of the service (e.g., "ws://127.0.0.1:16000").
func NewWSProxy(backendURL string, wsPaths []string, serviceName string) http.Handler {
	pathSet := make(map[string]bool, len(wsPaths))
	for _, p := range wsPaths {
		pathSet[p] = true
	}
	prefix := "/api/" + serviceName

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strip the service prefix and check against the whitelist.
		stripped := strings.TrimPrefix(r.URL.Path, prefix)
		if stripped == "" {
			stripped = "/"
		}

		if !pathSet[stripped] {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "WebSocket upgrade not allowed on this path",
			})
			return
		}

		// Dial the backend.
		backendWSURL := backendURL + stripped
		header := http.Header{}
		// Forward Authorization header if present (JWT from BFF).
		if auth := r.Header.Get("Authorization"); auth != "" {
			header.Set("Authorization", auth)
		}

		backendConn, _, err := websocket.DefaultDialer.Dial(backendWSURL, header)
		if err != nil {
			http.Error(w, "backend unavailable", http.StatusBadGateway)
			return
		}
		defer backendConn.Close()

		// Upgrade the client connection.
		clientConn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer clientConn.Close()

		// Bidirectional proxying.
		done := make(chan struct{})

		// Client → Backend
		go func() {
			defer close(done)
			pumpMessages(clientConn, backendConn)
		}()

		// Backend → Client
		pumpMessages(backendConn, clientConn)
		<-done
	})
}

func pumpMessages(src, dst *websocket.Conn) {
	for {
		mt, r, err := src.NextReader()
		if err != nil {
			_ = dst.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
		w, err := dst.NextWriter(mt)
		if err != nil {
			return
		}
		if _, err := io.Copy(w, r); err != nil {
			return
		}
		if err := w.Close(); err != nil {
			return
		}
	}
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/infra/httpapi/ -run TestWSProxy -v`
Expected: PASS — both tests pass

- [ ] **Step 6: Commit**

```bash
git add internal/infra/httpapi/ws.go internal/infra/httpapi/ws_test.go
git commit -m "feat(httpapi): add WebSocket proxy with path whitelist"
```

---

## Chunk 5: SPA Serving, Handler Wiring, and Composition Root

### Task 8: SPA Serving (Embedded and Dev Mode)

**Files:**
- Create: `internal/infra/httpapi/spa.go`
- Create: `internal/infra/httpapi/spa_test.go`

**Context:** In production, the SPA is served from an embedded `fs.FS`. Any path that doesn't match a real file falls back to `index.html` (SPA client-side routing). In dev mode, all non-`/api/*` requests are proxied to Vite's dev server.

- [ ] **Step 1: Write test for serving a static file**

```go
// internal/infra/httpapi/spa_test.go
package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/Work-Fort/WorkFort/internal/infra/httpapi"
)

func TestSPA_StaticFile(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html":          &fstest.MapFile{Data: []byte("<html>shell</html>")},
		"assets/app-abc.js":   &fstest.MapFile{Data: []byte("// app")},
	}

	handler := httpapi.NewSPAHandler(fsys)

	req := httptest.NewRequest(http.MethodGet, "/assets/app-abc.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "// app" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}
```

- [ ] **Step 2: Write test for SPA fallback to index.html**

```go
// internal/infra/httpapi/spa_test.go (append)

func TestSPA_Fallback(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>shell</html>")},
	}

	handler := httpapi.NewSPAHandler(fsys)

	// Unknown path should serve index.html (SPA routing).
	req := httptest.NewRequest(http.MethodGet, "/chat/general", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "<html>shell</html>" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}
```

- [ ] **Step 3: Write test for dev mode proxying**

```go
// internal/infra/httpapi/spa_test.go (append)

func TestSPA_DevMode(t *testing.T) {
	// Mock Vite dev server.
	vite := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("vite:" + r.URL.Path))
	}))
	defer vite.Close()

	handler := httpapi.NewSPADevProxy(vite.URL)

	req := httptest.NewRequest(http.MethodGet, "/chat/general", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "vite:/chat/general" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}
```

- [ ] **Step 4: Run tests to verify they fail**

Run: `go test ./internal/infra/httpapi/ -run TestSPA -v`
Expected: FAIL — spa.go does not exist yet

- [ ] **Step 5: Implement SPA handlers**

```go
// internal/infra/httpapi/spa.go
package httpapi

import (
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewSPAHandler serves an embedded SPA filesystem. Requests for real files
// are served directly. All other paths fall back to index.html for
// client-side routing.
func NewSPAHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(fsys))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the file exists in the embedded FS.
		path := r.URL.Path
		if path == "/" {
			path = "index.html"
		} else if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}

		if _, err := fs.Stat(fsys, path); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback: serve index.html for SPA routing.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

// NewSPADevProxy returns a handler that proxies all requests to a Vite
// dev server. Used with the --dev flag.
func NewSPADevProxy(devURL string) http.Handler {
	target, _ := url.Parse(devURL)
	return httputil.NewSingleHostReverseProxy(target)
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./internal/infra/httpapi/ -run TestSPA -v`
Expected: PASS — all 3 tests pass

- [ ] **Step 7: Commit**

```bash
git add internal/infra/httpapi/spa.go internal/infra/httpapi/spa_test.go
git commit -m "feat(httpapi): add SPA serving with index.html fallback and dev proxy"
```

### Task 9: Handler Wiring — Mux, Shell Endpoints, Full Routing

**Files:**
- Create: `internal/infra/httpapi/handler.go`
- Create: `internal/infra/httpapi/handler_test.go`

**Context:** The handler wires everything together: service proxies (with BFF auth for non-auth services), shell endpoints (`/api/services`, `/api/config`), and SPA fallback. Auth routes (`/api/auth/*`) use the same strip-prefix pattern but skip JWT conversion.

**Reference:** Read the "Shell endpoints" and "Route behavior" sections of `docs/2026-03-12-go-web-shell-design.md`.

- [ ] **Step 1: Write test for /api/services endpoint**

```go
// internal/infra/httpapi/handler_test.go
package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Work-Fort/WorkFort/internal/domain"
	"github.com/Work-Fort/WorkFort/internal/infra/httpapi"
)

func newTestFort() domain.Fort {
	return domain.Fort{
		Name:  "local",
		Local: true,
		Services: []domain.Service{
			{Name: "auth", URL: "http://127.0.0.1:3000", Enabled: true},
			{Name: "sharkfin", URL: "http://127.0.0.1:16000", Enabled: true, WSPaths: []string{"/ws", "/presence"}},
			{Name: "nexus", URL: "http://127.0.0.1:9600", Enabled: true},
			{Name: "hive", URL: "http://127.0.0.1:17000", Enabled: false},
		},
	}
}

func TestHandler_ServicesEndpoint(t *testing.T) {
	fort := newTestFort()
	handler := httpapi.NewHandler(fort, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Fort     string `json:"fort"`
		Services []struct {
			Name    string `json:"name"`
			Label   string `json:"label"`
			Route   string `json:"route"`
			Enabled bool   `json:"enabled"`
		} `json:"services"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Fort != "local" {
		t.Fatalf("expected fort 'local', got %q", resp.Fort)
	}
	if len(resp.Services) != 4 {
		t.Fatalf("expected 4 services, got %d", len(resp.Services))
	}

	// Check hive is disabled.
	for _, svc := range resp.Services {
		if svc.Name == "hive" && svc.Enabled {
			t.Fatal("expected hive to be disabled")
		}
	}
}
```

- [ ] **Step 2: Write test for /api/config endpoint**

```go
// internal/infra/httpapi/handler_test.go (append)

func TestHandler_ConfigEndpoint(t *testing.T) {
	fort := newTestFort()
	handler := httpapi.NewHandler(fort, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Fort string `json:"fort"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Fort != "local" {
		t.Fatalf("expected fort 'local', got %q", resp.Fort)
	}
}
```

- [ ] **Step 3: Write test for BFF routing — mock auth + mock backend**

This test creates a mock auth service that returns JWTs, a mock nexus backend, and
verifies the full BFF flow: browser sends cookie → handler converts to JWT → backend
receives Bearer token with path stripped.

```go
// internal/infra/httpapi/handler_test.go (append)

func TestHandler_BFFProxyRouting(t *testing.T) {
	// Mock auth service: returns a JWT when given a valid session cookie.
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("better-auth.session_token")
		if err != nil || cookie.Value != "valid-session" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "jwt-for-nexus"})
	}))
	defer authServer.Close()

	// Mock nexus backend: echoes the Authorization header and path.
	nexusBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Header.Get("Authorization") + "|" + r.URL.Path))
	}))
	defer nexusBackend.Close()

	fort := domain.Fort{
		Name:  "local",
		Local: true,
		Services: []domain.Service{
			{Name: "auth", URL: authServer.URL, Enabled: true},
			{Name: "nexus", URL: nexusBackend.URL, Enabled: true},
		},
	}

	tc := httpapi.NewTokenConverter(authServer.URL)
	handler := httpapi.NewHandler(fort, tc, nil)

	// Request with valid session cookie.
	req := httptest.NewRequest(http.MethodGet, "/api/nexus/v1/vms", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "valid-session"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	// Verify Bearer token was attached and path was stripped.
	if body != "Bearer jwt-for-nexus|/v1/vms" {
		t.Fatalf("expected 'Bearer jwt-for-nexus|/v1/vms', got %q", body)
	}
}

func TestHandler_BFFProxyRouting_NoCookie(t *testing.T) {
	fort := domain.Fort{
		Name:  "local",
		Local: true,
		Services: []domain.Service{
			{Name: "auth", URL: "http://127.0.0.1:3000", Enabled: true},
			{Name: "nexus", URL: "http://127.0.0.1:9600", Enabled: true},
		},
	}

	tc := httpapi.NewTokenConverter("http://127.0.0.1:3000")
	handler := httpapi.NewHandler(fort, tc, nil)

	// Request without session cookie should get 401.
	req := httptest.NewRequest(http.MethodGet, "/api/nexus/v1/vms", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 (no cookie), got %d", rec.Code)
	}
}
```

- [ ] **Step 4: Write test for SPA fallback on non-API paths**

```go
// internal/infra/httpapi/handler_test.go (append)

func TestHandler_SPAFallback(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>shell</html>")},
	}

	fort := newTestFort()
	handler := httpapi.NewHandler(fort, nil, fsys)

	req := httptest.NewRequest(http.MethodGet, "/chat/general", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "<html>shell</html>" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}
```

- [ ] **Step 5: Run tests to verify they fail**

Run: `go test ./internal/infra/httpapi/ -run TestHandler -v`
Expected: FAIL — handler.go does not exist yet

Add the `"testing/fstest"` import to the test file.

- [ ] **Step 6: Implement the handler**

```go
// internal/infra/httpapi/handler.go
package httpapi

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"sort"

	"github.com/Work-Fort/WorkFort/internal/domain"
)

// Service metadata for nav tabs — presentation concern, not domain.
var serviceMetadata = map[string]struct{ Label, Route string }{
	"auth":     {"Auth", "/auth"},
	"sharkfin": {"Chat", "/chat"},
	"nexus":    {"Nexus", "/nexus"},
	"hive":     {"Hive", "/hive"},
}

// NewHandler creates the top-level HTTP handler for the web shell.
//
// Parameters:
//   - fort: the active fort configuration
//   - tc: token converter for BFF auth (nil disables BFF — only shell endpoints and SPA work)
//   - spaFS: embedded SPA filesystem (nil disables SPA serving — use NewSPADevProxy for dev mode)
func NewHandler(fort domain.Fort, tc *TokenConverter, spaFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	// Shell endpoints.
	mux.HandleFunc("GET /api/services", servicesHandler(fort))
	mux.HandleFunc("GET /api/config", configHandler(fort))

	// Service proxies.
	for _, svc := range fort.Services {
		prefix := "/api/" + svc.Name + "/"

		if svc.Name == "auth" {
			// Auth routes are pass-through — no BFF conversion.
			proxy := NewServiceProxy(svc, fort.Local, fort.Gateway)
			mux.Handle(prefix, proxy)
			continue
		}

		// Non-auth services get BFF conversion.
		proxy := NewServiceProxy(svc, fort.Local, fort.Gateway)

		// WebSocket handler for services with WS paths.
		var wsHandler http.Handler
		if len(svc.WSPaths) > 0 && svc.Enabled {
			// Convert http(s) → ws(s) by replacing the scheme prefix.
			// Both "http" and "https" are 4+ chars, so [4:] maps correctly:
			//   "http://..." → "ws://..."  |  "https://..." → "wss://..."
			wsURL := "ws" + svc.URL[4:]
			if !fort.Local {
				wsURL = "ws" + fort.Gateway[4:]
			}
			wsHandler = NewWSProxy(wsURL, svc.WSPaths, svc.Name)
		}

		mux.Handle(prefix, bffMiddleware(tc, svc, proxy, wsHandler))
	}

	// SPA fallback.
	if spaFS != nil {
		mux.Handle("/", NewSPAHandler(spaFS))
	}

	return mux
}

// bffMiddleware wraps a service proxy with BFF token conversion.
// WebSocket upgrade requests are routed to the wsHandler if available.
func bffMiddleware(tc *TokenConverter, svc domain.Service, proxy http.Handler, wsHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for WebSocket upgrade.
		if wsHandler != nil && isWebSocketUpgrade(r) {
			// BFF: convert cookie to JWT before WS upgrade.
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

		// Regular HTTP — BFF conversion.
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

		// Replace cookie auth with Bearer token for downstream.
		r.Header.Set("Authorization", "Bearer "+token)
		proxy.ServeHTTP(w, r)
	})
}

func writeAuthError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	switch err {
	case errNoSession, errSessionExpired:
		w.WriteHeader(http.StatusUnauthorized)
		// Clear session cookie on expiry.
		if err == errSessionExpired {
			http.SetCookie(w, &http.Cookie{
				Name:   sessionCookieName,
				Value:  "",
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

func isWebSocketUpgrade(r *http.Request) bool {
	for _, v := range r.Header.Values("Connection") {
		if v == "Upgrade" || v == "upgrade" {
			return true
		}
	}
	return false
}

func servicesHandler(fort domain.Fort) http.HandlerFunc {
	type serviceInfo struct {
		Name    string `json:"name"`
		Label   string `json:"label"`
		Route   string `json:"route"`
		Enabled bool   `json:"enabled"`
		UI      bool   `json:"ui"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		svcs := make([]serviceInfo, 0, len(fort.Services))
		for _, svc := range fort.Services {
			meta, ok := serviceMetadata[svc.Name]
			if !ok {
				meta = struct{ Label, Route string }{svc.Name, "/" + svc.Name}
			}
			svcs = append(svcs, serviceInfo{
				Name:    svc.Name,
				Label:   meta.Label,
				Route:   meta.Route,
				Enabled: svc.Enabled,
				UI:      false, // Set by probing /ui/health at startup
			})
		}
		sort.Slice(svcs, func(i, j int) bool { return svcs[i].Name < svcs[j].Name })

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"fort":     fort.Name,
			"services": svcs,
		})
	}
}

func configHandler(fort domain.Fort) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"fort": fort.Name,
		})
	}
}
```

- [ ] **Step 7: Run all handler tests to verify they pass**

Run: `go test ./internal/infra/httpapi/ -run TestHandler -v`
Expected: PASS — all 5 tests pass (ServicesEndpoint, ConfigEndpoint, BFFProxyRouting, BFFProxyRouting_NoCookie, SPAFallback)

- [ ] **Step 8: Run all httpapi tests together**

Run: `go test ./internal/infra/httpapi/ -v -race`
Expected: PASS — all tests in the package pass

- [ ] **Step 9: Commit**

```bash
git add internal/infra/httpapi/handler.go internal/infra/httpapi/handler_test.go
git commit -m "feat(httpapi): add handler wiring with shell endpoints and BFF middleware"
```

### Task 10: Composition Root — `cmd/web/`

**Files:**
- Create: `cmd/web/web.go`

**Context:** The `workfort web` command wires all layers: reads fort config, sets up BFF token converter, builds reverse proxies, registers SPA handler, and starts the HTTP server with graceful shutdown. No test file — this is a composition root wired in main; it's verified by building and running.

**Reference:** Read `cmd/root.go` (lines 27-51) for how the existing command structure works.

Note: The embedded SPA filesystem doesn't exist yet (the frontend hasn't been built). Use a placeholder `embed.go` with an empty directory. The embed directive will be updated when the frontend is built.

- [ ] **Step 1: Create placeholder embed file**

```go
// cmd/web/embed.go
package web

import "embed"

// webFS holds the embedded shell SPA. In production builds, this contains
// the Vite build output from web/dist/. During development, use --dev
// to proxy to Vite's dev server instead.
//
//go:embed all:placeholder
var webFS embed.FS
```

Create the placeholder directory:

```bash
mkdir -p cmd/web/placeholder
echo '<!DOCTYPE html><html><body><p>Shell SPA not built yet. Run with --dev to proxy to Vite dev server.</p></body></html>' > cmd/web/placeholder/index.html
```

- [ ] **Step 2: Implement the web command**

```go
// cmd/web/web.go
package web

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"charm.land/log/v2"
	"github.com/spf13/cobra"

	"github.com/Work-Fort/WorkFort/internal/infra/fortconfig"
	"github.com/Work-Fort/WorkFort/internal/infra/httpapi"
)

var (
	bind   string
	port   int
	dev    bool
	devURL string
	noOpen bool
)

// New creates the "web" subcommand.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Start the web UI server",
		Long:  "Serve the WorkFort web shell and proxy requests to backend services.",
		RunE:  run,
	}

	cmd.Flags().StringVar(&bind, "bind", "127.0.0.1", "Listen address")
	cmd.Flags().IntVar(&port, "port", 8080, "Listen port")
	cmd.Flags().BoolVar(&dev, "dev", false, "Proxy SPA to Vite dev server")
	cmd.Flags().StringVar(&devURL, "dev-url", "http://localhost:5173", "Vite dev server URL (used with --dev)")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Don't auto-open browser")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	registry := fortconfig.New()
	fort := registry.Active()

	log.Info("starting web server",
		"fort", fort.Name,
		"local", fort.Local,
		"services", len(fort.Services),
	)

	// Find the auth service URL for BFF token conversion.
	var authURL string
	for _, svc := range fort.Services {
		if svc.Name == "auth" && svc.Enabled {
			authURL = svc.URL
			break
		}
	}

	var tc *httpapi.TokenConverter
	if authURL != "" {
		tc = httpapi.NewTokenConverter(authURL)
	} else {
		log.Warn("auth service not configured — BFF token conversion disabled")
	}

	// SPA handler.
	var spaFS fs.FS
	if !dev {
		sub, err := fs.Sub(webFS, "placeholder")
		if err != nil {
			return fmt.Errorf("embedded SPA: %w", err)
		}
		spaFS = sub
	}

	handler := httpapi.NewHandler(fort, tc, spaFS)

	// In dev mode, wrap the handler to proxy non-/api/* to Vite.
	if dev {
		devProxy := httpapi.NewSPADevProxy(devURL)
		topMux := http.NewServeMux()
		topMux.Handle("/api/", handler)
		topMux.Handle("/", devProxy)
		handler = topMux
	}

	addr := fmt.Sprintf("%s:%d", bind, port)
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Info("shutting down web server")
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutCtx)
	}()

	url := fmt.Sprintf("http://%s", addr)
	log.Info("web server listening", "url", url)

	if !noOpen {
		go openBrowser(url)
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("web server: %w", err)
	}

	return nil
}

func openBrowser(url string) {
	// Small delay to let the server start.
	time.Sleep(200 * time.Millisecond)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	_ = cmd.Start()
}
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./cmd/web/`
Expected: Success

- [ ] **Step 4: Commit**

```bash
git add cmd/web/web.go cmd/web/embed.go cmd/web/placeholder/index.html
git commit -m "feat(web): add workfort web command with BFF proxy and SPA serving"
```

### Task 11: CGo Build Gating and Command Registration

**Files:**
- Modify: `cmd/root.go`
- Create: `cmd/root_cgo.go`
- Modify: `cmd/chat/chat.go`

**Context:** Move the chat command registration behind a `//go:build cgo` tag so that `CGO_ENABLED=0` builds exclude it. The web command is registered unconditionally. Rename `NewChatCmd()` to `New()` for consistency.

**Reference:** Read `cmd/root.go` (current state) and `cmd/chat/chat.go`.

- [ ] **Step 1: Rename `NewChatCmd` to `New` in `cmd/chat/chat.go`**

In `cmd/chat/chat.go` line 19, change:
```go
func NewChatCmd() *cobra.Command {
```
to:
```go
func New() *cobra.Command {
```

- [ ] **Step 2: Create `cmd/root_cgo.go` with build tag**

```go
// cmd/root_cgo.go
//go:build cgo

package cmd

import "github.com/Work-Fort/WorkFort/cmd/chat"

func init() {
	rootCmd.AddCommand(chat.New())
}
```

- [ ] **Step 3: Update `cmd/root.go` — remove chat import, add web import**

Remove the chat import and registration. Add the web command:

```go
// cmd/root.go — updated imports
import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Work-Fort/WorkFort/cmd/web"
	"github.com/Work-Fort/WorkFort/pkg/config"
	"github.com/Work-Fort/WorkFort/pkg/ui"
)
```

In `init()`, replace `rootCmd.AddCommand(chat.NewChatCmd())` with:
```go
	rootCmd.AddCommand(web.New())
```

- [ ] **Step 4: Verify CGo build works**

Run: `go build .`
Expected: Success — both chat and web commands registered (CGo enabled by default on Linux)

- [ ] **Step 5: Verify non-CGo build works**

Run: `CGO_ENABLED=0 go build .`
Expected: Success — only web command registered, chat excluded

Note: This may fail if other packages (like `pkg/stt`) are imported unconditionally from somewhere other than `cmd/chat/`. The expected CGo import chain is: `cmd/chat/` → `internal/chat/` → `pkg/stt/` (whisper.cpp CGo). If the `CGO_ENABLED=0` build fails, trace imports with `go list -deps ./cmd/chat/` and ensure the chain doesn't leak into `cmd/root.go`. Other packages (`pkg/audio/`) may also use CGo — check with `go list -f '{{.CgoFiles}}' ./...`.

- [ ] **Step 6: Run all tests**

Run: `go test ./... -race -count=1`
Expected: PASS — all tests across all packages pass

- [ ] **Step 7: Commit**

```bash
git add cmd/root.go cmd/root_cgo.go cmd/chat/chat.go
git commit -m "refactor(cmd): gate chat behind CGo build tag, add web command"
```
