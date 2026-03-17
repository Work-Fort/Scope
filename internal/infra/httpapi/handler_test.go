package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/Work-Fort/Scope/internal/domain"
	"github.com/Work-Fort/Scope/internal/infra/httpapi"
)

// newTestTracker spins up mock HTTP servers that respond to /ui/health,
// then creates a ServiceTracker and runs InitialProbe. Returns the tracker
// and a cleanup function that closes the mock servers.
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

func TestHandler_ConfigEndpoint(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()
	fort := newTestFort(tracker)

	handler := httpapi.NewHandler(fort, tracker, nil, nil)

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

func TestHandler_BFFProxyRouting(t *testing.T) {
	// Mock auth service: health + BFF token conversion.
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		// Token conversion endpoint.
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
		if r.URL.Path == "/ui/health" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"status": "ok",
				"name":   "nexus",
				"label":  "Nexus",
				"route":  "/nexus",
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Header.Get("Authorization") + "|" + r.URL.Path))
	}))
	defer nexusBackend.Close()

	tracker := httpapi.NewServiceTracker([]string{authServer.URL, nexusBackend.URL})
	tracker.InitialProbe(context.Background())

	fort := domain.Fort{
		Name:  "local",
		Local: true,
		Services: []domain.ConfigService{
			{URL: authServer.URL},
			{URL: nexusBackend.URL},
		},
	}

	tc := httpapi.NewTokenConverter(authServer.URL)
	handler := httpapi.NewHandler(fort, tracker, tc, nil)

	// Request with valid session cookie.
	req := httptest.NewRequest(http.MethodGet, "/api/nexus/v1/vms", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "valid-session"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if body != "Bearer jwt-for-nexus|/v1/vms" {
		t.Fatalf("expected 'Bearer jwt-for-nexus|/v1/vms', got %q", body)
	}
}

func TestHandler_BFFProxyRouting_NoCookie(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()

	fort := newTestFort(tracker)

	// Need a real auth URL for token converter.
	authSvc, _ := tracker.ServiceByName("auth")
	tc := httpapi.NewTokenConverter(authSvc.URL)
	handler := httpapi.NewHandler(fort, tracker, tc, nil)

	// Request without session cookie should get 401.
	req := httptest.NewRequest(http.MethodGet, "/api/nexus/v1/vms", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 (no cookie), got %d", rec.Code)
	}
}

func TestHandler_UIAssetsServedWithoutAuth(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()
	fort := newTestFort(tracker)

	// Pass nil token converter — simulates no auth service configured.
	handler := httpapi.NewHandler(fort, tracker, nil, nil)

	// These paths SHOULD bypass auth.
	allowedPaths := []string{
		"/api/sharkfin/ui/remoteEntry.js",
		"/api/sharkfin/ui/health",
		"/api/sharkfin/ui/assets/index-abc123.js",
		"/api/sharkfin/ui/assets/style.css",
	}
	for _, p := range allowedPaths {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code == http.StatusUnauthorized {
			t.Errorf("path %q should bypass auth, got 401", p)
		}
	}
}

func TestHandler_UIArbitraryPathsRequireAuth(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()
	fort := newTestFort(tracker)

	// Pass nil token converter — without auth, non-asset paths should get 401.
	handler := httpapi.NewHandler(fort, tracker, nil, nil)

	// These paths should NOT bypass auth.
	blockedPaths := []string{
		"/api/sharkfin/ui/some-api",
		"/api/sharkfin/ui/admin/secret",
	}
	for _, p := range blockedPaths {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("path %q should require auth, got %d", p, rec.Code)
		}
	}
}

func TestHandler_ServicesIncludesSetupMode(t *testing.T) {
	authWithSetup := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]any{
			"status":     "ok",
			"name":       "auth",
			"label":      "Auth",
			"route":      "",
			"setup_mode": true,
		})
	}))
	defer authWithSetup.Close()

	tracker := httpapi.NewServiceTracker([]string{authWithSetup.URL})
	tracker.InitialProbe(context.Background())

	fort := domain.Fort{Name: "local", Local: true}
	handler := httpapi.NewHandler(fort, tracker, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var resp struct {
		Services []struct {
			Name      string `json:"name"`
			SetupMode bool   `json:"setup_mode,omitempty"`
		} `json:"services"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp.Services) == 0 {
		t.Fatal("expected at least one service")
	}
	if !resp.Services[0].SetupMode {
		t.Fatal("expected setup_mode to be true for auth service")
	}
}

func TestHandler_SPAFallback(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>shell</html>")},
	}

	tracker, cleanup := newTestTracker(t)
	defer cleanup()
	fort := newTestFort(tracker)

	handler := httpapi.NewHandler(fort, tracker, nil, fsys)

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
