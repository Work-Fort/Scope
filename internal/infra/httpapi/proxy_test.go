package httpapi_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Work-Fort/Scope/internal/domain"
	"github.com/Work-Fort/Scope/internal/infra/httpapi"
)

func TestProxy_PathStripping(t *testing.T) {
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

func TestProxy_GatewayFort(t *testing.T) {
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
