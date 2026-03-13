package httpapi

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"sort"

	"github.com/Work-Fort/Scope/internal/domain"
)

// Service metadata for nav tabs — presentation concern, not domain.
var serviceMetadata = map[string]struct{ Label, Route string }{
	"auth":     {"Auth", "/auth"},
	"sharkfin": {"Chat", "/chat"},
	"nexus":    {"Nexus", "/nexus"},
	"hive":     {"Hive", "/hive"},
}

// NewHandler creates the top-level HTTP handler for the web shell.
//
// Parameters:
//   - fort: the active fort configuration
//   - tc: token converter for BFF auth (nil disables BFF — only shell endpoints and SPA work)
//   - spaFS: embedded SPA filesystem (nil disables SPA serving — use NewSPADevProxy for dev mode)
func NewHandler(fort domain.Fort, tc *TokenConverter, spaFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	// Shell endpoints.
	mux.HandleFunc("GET /api/services", servicesHandler(fort))
	mux.HandleFunc("GET /api/config", configHandler(fort))

	// Service proxies.
	for _, svc := range fort.Services {
		prefix := "/api/" + svc.Name + "/"

		if svc.Name == "auth" {
			// Auth routes are pass-through — no BFF conversion.
			proxy := NewServiceProxy(svc, fort.Local, fort.Gateway)
			mux.Handle(prefix, proxy)
			continue
		}

		// Non-auth services get BFF conversion.
		proxy := NewServiceProxy(svc, fort.Local, fort.Gateway)

		// WebSocket handler for services with WS paths.
		var wsHandler http.Handler
		if len(svc.WSPaths) > 0 && svc.Enabled {
			wsURL := "ws" + svc.URL[4:]
			if !fort.Local {
				wsURL = "ws" + fort.Gateway[4:]
			}
			wsHandler = NewWSProxy(wsURL, svc.WSPaths, svc.Name)
		}

		mux.Handle(prefix, bffMiddleware(tc, svc, proxy, wsHandler))
	}

	// SPA fallback.
	if spaFS != nil {
		mux.Handle("/", NewSPAHandler(spaFS))
	}

	return mux
}

// bffMiddleware wraps a service proxy with BFF token conversion.
// WebSocket upgrade requests are routed to the wsHandler if available.
func bffMiddleware(tc *TokenConverter, svc domain.Service, proxy http.Handler, wsHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for WebSocket upgrade.
		if wsHandler != nil && isWebSocketUpgrade(r) {
			// BFF: convert cookie to JWT before WS upgrade.
			if tc != nil {
				token, err := tc.Token(r)
				if err != nil {
					writeAuthError(w, err)
					return
				}
				r.Header.Set("Authorization", "Bearer "+token)
			}
			wsHandler.ServeHTTP(w, r)
			return
		}

		// Regular HTTP — BFF conversion.
		if tc == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "auth: no token converter configured"})
			return
		}

		token, err := tc.Token(r)
		if err != nil {
			writeAuthError(w, err)
			return
		}

		// Replace cookie auth with Bearer token for downstream.
		r.Header.Set("Authorization", "Bearer "+token)
		proxy.ServeHTTP(w, r)
	})
}

func writeAuthError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	switch err {
	case errNoSession, errSessionExpired:
		w.WriteHeader(http.StatusUnauthorized)
		// Clear session cookie on expiry.
		if err == errSessionExpired {
			http.SetCookie(w, &http.Cookie{
				Name:   sessionCookieName,
				Value:  "",
				MaxAge: -1,
			})
		}
	case errAuthDown:
		w.WriteHeader(http.StatusBadGateway)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func isWebSocketUpgrade(r *http.Request) bool {
	for _, v := range r.Header.Values("Connection") {
		if v == "Upgrade" || v == "upgrade" {
			return true
		}
	}
	return false
}

func servicesHandler(fort domain.Fort) http.HandlerFunc {
	type serviceInfo struct {
		Name    string `json:"name"`
		Label   string `json:"label"`
		Route   string `json:"route"`
		Enabled bool   `json:"enabled"`
		UI      bool   `json:"ui"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		svcs := make([]serviceInfo, 0, len(fort.Services))
		for _, svc := range fort.Services {
			meta, ok := serviceMetadata[svc.Name]
			if !ok {
				meta = struct{ Label, Route string }{svc.Name, "/" + svc.Name}
			}
			svcs = append(svcs, serviceInfo{
				Name:    svc.Name,
				Label:   meta.Label,
				Route:   meta.Route,
				Enabled: svc.Enabled,
				UI:      false, // Set by probing /ui/health at startup
			})
		}
		sort.Slice(svcs, func(i, j int) bool { return svcs[i].Name < svcs[j].Name })

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"fort":     fort.Name,
			"services": svcs,
		})
	}
}

func configHandler(fort domain.Fort) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"fort": fort.Name,
		})
	}
}
