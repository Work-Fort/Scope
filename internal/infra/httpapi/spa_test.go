package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/Work-Fort/Scope/internal/infra/httpapi"
)

func TestSPA_StaticFile(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html":        &fstest.MapFile{Data: []byte("<html>shell</html>")},
		"assets/app-abc.js": &fstest.MapFile{Data: []byte("// app")},
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
