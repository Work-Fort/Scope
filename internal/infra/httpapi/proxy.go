package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/Work-Fort/Scope/internal/domain"
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
