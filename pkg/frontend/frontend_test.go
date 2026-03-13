package frontend_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/Work-Fort/Scope/pkg/frontend"
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

func TestCacheHeaders_Assets(t *testing.T) {
	fsys := fstest.MapFS{
		"remoteEntry.js":      &fstest.MapFile{Data: []byte("// entry")},
		"assets/chunk-abc.js": &fstest.MapFile{Data: []byte("// chunk")},
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

	cc := rec.Header().Get("Cache-Control")
	if cc != "" {
		t.Fatalf("expected no Cache-Control on 404, got %q", cc)
	}
}
