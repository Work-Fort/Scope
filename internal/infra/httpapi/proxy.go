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

// NewAuthProxy creates a reverse proxy for auth that rewrites Set-Cookie paths
// to scope cookies to the fort's URL prefix.
func NewAuthProxy(serviceName, targetURL string, local bool, gatewayURL string, fortName string) http.Handler {
	prefix := "/api/" + serviceName
	cookiePath := "/forts/" + fortName + "/"

	if local {
		target, _ := url.Parse(targetURL)
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ModifyResponse = func(resp *http.Response) error {
			rewriteCookiePaths(resp, cookiePath)
			return nil
		}
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
	proxy.ModifyResponse = func(resp *http.Response) error {
		rewriteCookiePaths(resp, cookiePath)
		return nil
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
}

func rewriteCookiePaths(resp *http.Response, path string) {
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		return
	}
	resp.Header.Del("Set-Cookie")
	for _, c := range cookies {
		c.Path = path
		if v := c.String(); v != "" {
			resp.Header.Add("Set-Cookie", v)
		}
	}
}
