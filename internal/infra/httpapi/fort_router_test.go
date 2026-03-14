package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
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

func TestFortRouter_InvalidFortName_404(t *testing.T) {
	reg := &mockRegistry{forts: []domain.Fort{}}
	router := httpapi.NewFortRouter(reg, nil)

	// These names must be valid URL path segments but invalid fort identifiers.
	// "UPPER" contains uppercase letters — rejected by ValidFortName.
	// "has%20space" is URL-encoded space — decodes to "has space", rejected by ValidFortName.
	// "-leading" starts with a hyphen — rejected by ValidFortName.
	for _, rawPath := range []string{"/forts/UPPER/api/services", "/forts/has%20space/api/services", "/forts/-leading/api/services"} {
		req := httptest.NewRequest("GET", rawPath, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 404 {
			t.Errorf("path %q: got status %d, want 404", rawPath, w.Code)
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

func TestFortRouter_DispatchToFort(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()

	fort := newTestFort(tracker)
	reg := &mockRegistry{forts: []domain.Fort{fort}}
	router := httpapi.NewFortRouter(reg, nil)

	req := httptest.NewRequest("GET", "/forts/local/api/services", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", w.Code)
	}
}

func TestFortRouter_ConcurrentInit(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()

	fort := newTestFort(tracker)
	reg := &mockRegistry{forts: []domain.Fort{fort}}
	router := httpapi.NewFortRouter(reg, nil)

	const N = 10
	var wg sync.WaitGroup
	codes := make([]int, N)

	wg.Add(N)
	for i := 0; i < N; i++ {
		i := i
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/forts/local/api/services", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			codes[i] = w.Code
		}()
	}
	wg.Wait()

	for i, code := range codes {
		if code != http.StatusOK {
			t.Errorf("goroutine %d: got status %d, want 200", i, code)
		}
	}
}
