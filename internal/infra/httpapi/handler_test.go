package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/Work-Fort/Scope/internal/domain"
	"github.com/Work-Fort/Scope/internal/infra/httpapi"
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
