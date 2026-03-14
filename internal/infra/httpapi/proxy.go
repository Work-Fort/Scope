package httpapi

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// NewServiceProxy creates an http.Handler that proxies requests to a service.
//
// For local forts (local=true): strips the /api/{serviceName} prefix and forwards
// to the targetURL (e.g., /api/nexus/v1/vms -> http://target/v1/vms).
//
// For gateway forts (local=false): preserves the /api/{serviceName} prefix and
// forwards to the gatewayURL.
func NewServiceProxy(serviceName, targetURL string, local bool, gatewayURL string) http.Handler {
	prefix := "/api/" + serviceName

	if local {
		target, _ := url.Parse(targetURL)
		proxy := httputil.NewSingleHostReverseProxy(target)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
			proxy.ServeHTTP(w, r)
		})
	}

	target, _ := url.Parse(gatewayURL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
}
