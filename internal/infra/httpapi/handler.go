package httpapi

import (
	"encoding/json"
	"io/fs"
	"net/http"

	"github.com/Work-Fort/Scope/internal/domain"
)

// NewHandler creates the top-level HTTP handler for the web shell.
//
// Parameters:
//   - fort: the active fort configuration
//   - tracker: live service tracker (health probing + WS connection tracking)
//   - tc: token converter for BFF auth (nil disables BFF — only shell endpoints and SPA work)
//   - spaFS: embedded SPA filesystem (nil disables SPA serving — use NewSPADevProxy for dev mode)
func NewHandler(fort domain.Fort, tracker *ServiceTracker, tc *TokenConverter, spaFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	// Shell endpoints.
	mux.HandleFunc("GET /api/services", servicesHandler(fort.Name, tracker))
	mux.HandleFunc("GET /api/config", configHandler(fort.Name))

	// Register routes for all currently discovered services.
	registerServiceRoutes(mux, tracker, tc, fort)

	// Register routes for services discovered after initial probe (via polling).
	tracker.OnServiceDiscovered = func(svc TrackedService) {
		registerOneServiceRoute(mux, svc, tracker, tc, fort)
	}

	// SPA fallback.
	if spaFS != nil {
		mux.Handle("/", NewSPAHandler(spaFS))
	}

	return mux
}

func registerServiceRoutes(mux *http.ServeMux, tracker *ServiceTracker, tc *TokenConverter, fort domain.Fort) {
	for _, svc := range tracker.Services() {
		registerOneServiceRoute(mux, svc, tracker, tc, fort)
	}
}

func registerOneServiceRoute(mux *http.ServeMux, svc TrackedService, tracker *ServiceTracker, tc *TokenConverter, fort domain.Fort) {
	prefix := "/api/" + svc.Name + "/"

	if svc.Name == "auth" {
		proxy := NewAuthProxy(svc.Name, svc.URL, fort.Local, fort.Gateway, fort.Name)
		mux.Handle("/api/auth/", proxy)
		return
	}

	// Non-auth services get BFF conversion.
	proxy := NewServiceProxy(svc.Name, svc.URL, fort.Local, fort.Gateway)

	// WebSocket handler for services with WS paths.
	var wsHandler http.Handler
	if svc.hasWS {
		wsURL := "ws" + svc.URL[4:]
		if !fort.Local {
			wsURL = "ws" + fort.Gateway[4:]
		}
		wsHandler = NewWSProxy(wsURL, svc.WSPaths, svc.Name, &ConnectionCallbacks{
			OnConnect:    tracker.OnConnect,
			OnDisconnect: tracker.OnDisconnect,
		})
	}

	mux.Handle(prefix, bffMiddleware(fort.Name, tc, proxy, wsHandler))
}

// bffMiddleware wraps a service proxy with BFF token conversion.
// WebSocket upgrade requests are routed to the wsHandler if available.
func bffMiddleware(fortName string, tc *TokenConverter, proxy http.Handler, wsHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for WebSocket upgrade.
		if wsHandler != nil && isWebSocketUpgrade(r) {
			// BFF: convert cookie to JWT before WS upgrade.
			if tc != nil {
				token, err := tc.Token(r)
				if err != nil {
					writeAuthError(w, err, fortName)
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
			writeAuthError(w, err, fortName)
			return
		}

		// Replace cookie auth with Bearer token for downstream.
		r.Header.Set("Authorization", "Bearer "+token)
		proxy.ServeHTTP(w, r)
	})
}

func writeAuthError(w http.ResponseWriter, err error, fortName string) {
	w.Header().Set("Content-Type", "application/json")
	switch err {
	case errNoSession, errSessionExpired:
		w.WriteHeader(http.StatusUnauthorized)
		if err == errSessionExpired {
			http.SetCookie(w, &http.Cookie{
				Name:   sessionCookieName,
				Value:  "",
				Path:   "/forts/" + fortName + "/",
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

func servicesHandler(fortName string, tracker *ServiceTracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"fort":      fortName,
			"services":  tracker.Services(),
			"conflicts": tracker.Conflicts(),
		})
	}
}

func configHandler(fortName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"fort": fortName,
		})
	}
}
