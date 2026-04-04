package frontend_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/Work-Fort/Scope/go/frontend"
)

func TestHealthProbe_OK(t *testing.T) {
	fsys := fstest.MapFS{
		"remoteEntry.js": &fstest.MapFile{Data: []byte("// entry")},
	}

	handler := frontend.Handler(fsys, frontend.Manifest{})
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
	if body != `{"status":"ok","name":"","label":"","route":""}`+"\n" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestHealthProbe_Unavailable(t *testing.T) {
	fsys := fstest.MapFS{
		"other.js": &fstest.MapFile{Data: []byte("// other")},
	}

	handler := frontend.Handler(fsys, frontend.Manifest{})
	req := httptest.NewRequest(http.MethodGet, "/ui/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	body := rec.Body.String()
	if body != `{"status":"unavailable","name":"","label":"","route":""}`+"\n" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestCacheHeaders_Assets(t *testing.T) {
	fsys := fstest.MapFS{
		"remoteEntry.js":      &fstest.MapFile{Data: []byte("// entry")},
		"assets/chunk-abc.js": &fstest.MapFile{Data: []byte("// chunk")},
	}

	handler := frontend.Handler(fsys, frontend.Manifest{})
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

func TestCacheHeaders_RemoteEntry(t *testing.T) {
	fsys := fstest.MapFS{
		"remoteEntry.js": &fstest.MapFile{Data: []byte("// entry")},
	}

	handler := frontend.Handler(fsys, frontend.Manifest{})
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

func TestCacheHeaders_OtherFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"remoteEntry.js": &fstest.MapFile{Data: []byte("// entry")},
		"manifest.json":  &fstest.MapFile{Data: []byte(`{}`)},
	}

	handler := frontend.Handler(fsys, frontend.Manifest{})
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

func TestFileNotFound(t *testing.T) {
	fsys := fstest.MapFS{
		"remoteEntry.js": &fstest.MapFile{Data: []byte("// entry")},
	}

	handler := frontend.Handler(fsys, frontend.Manifest{})
	req := httptest.NewRequest(http.MethodGet, "/ui/nonexistent.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	cc := rec.Header().Get("Cache-Control")
	if cc != "" {
		t.Fatalf("expected no Cache-Control on 404, got %q", cc)
	}
}

func TestHealthProbe_WithManifest(t *testing.T) {
	fsys := fstest.MapFS{
		"remoteEntry.js": &fstest.MapFile{Data: []byte("// entry")},
	}

	m := frontend.Manifest{
		Name:    "chat",
		Label:   "Chat",
		Route:   "/chat",
		WSPaths: []string{"/ws/chat"},
	}

	handler := frontend.Handler(fsys, m)
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
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Fatalf("expected status ok, got %q", resp.Status)
	}
	if resp.Name != "chat" {
		t.Fatalf("expected name chat, got %q", resp.Name)
	}
	if resp.Label != "Chat" {
		t.Fatalf("expected label Chat, got %q", resp.Label)
	}
	if resp.Route != "/chat" {
		t.Fatalf("expected route /chat, got %q", resp.Route)
	}
	if len(resp.WSPaths) != 1 || resp.WSPaths[0] != "/ws/chat" {
		t.Fatalf("expected ws_paths [/ws/chat], got %v", resp.WSPaths)
	}
}

func TestHealthProbe_Unavailable_IncludesManifest(t *testing.T) {
	fsys := fstest.MapFS{
		"other.js": &fstest.MapFile{Data: []byte("// other")},
	}

	m := frontend.Manifest{
		Name:  "chat",
		Label: "Chat",
		Route: "/chat",
	}

	handler := frontend.Handler(fsys, m)
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
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "unavailable" {
		t.Fatalf("expected status unavailable, got %q", resp.Status)
	}
	if resp.Name != "chat" {
		t.Fatalf("expected name chat, got %q", resp.Name)
	}
}
