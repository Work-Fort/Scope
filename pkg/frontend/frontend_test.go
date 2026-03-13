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
