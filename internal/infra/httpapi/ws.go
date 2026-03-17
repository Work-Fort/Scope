package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // same-origin or non-browser client
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		return u.Host == r.Host
	},
}

// ConnectionCallbacks notifies the caller of WebSocket connection lifecycle events.
type ConnectionCallbacks struct {
	OnConnect    func(service string)
	OnDisconnect func(service string)
}

// NewWSProxy creates an http.Handler that proxies WebSocket connections
// to a backend service. Only paths in wsPaths are allowed — others return 400.
//
// The serviceName is used to strip the /api/{service} prefix before matching.
// The backendURL is the base WebSocket URL of the service (e.g., "ws://127.0.0.1:16000").
func NewWSProxy(backendURL string, wsPaths []string, serviceName string, cb *ConnectionCallbacks) http.Handler {
	pathSet := make(map[string]bool, len(wsPaths))
	for _, p := range wsPaths {
		pathSet[p] = true
	}
	prefix := "/api/" + serviceName

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strip the service prefix and check against the whitelist.
		stripped := strings.TrimPrefix(r.URL.Path, prefix)
		if stripped == "" {
			stripped = "/"
		}

		if !pathSet[stripped] {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "WebSocket upgrade not allowed on this path",
			})
			return
		}

		// Dial the backend.
		backendWSURL := backendURL + stripped
		header := http.Header{}
		// Forward Authorization header if present (JWT from BFF).
		if auth := r.Header.Get("Authorization"); auth != "" {
			header.Set("Authorization", auth)
		}

		backendConn, _, err := websocket.DefaultDialer.Dial(backendWSURL, header)
		if err != nil {
			http.Error(w, "backend unavailable", http.StatusBadGateway)
			return
		}
		defer backendConn.Close()

		// Upgrade the client connection.
		clientConn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer clientConn.Close()

		// Notify connection established.
		if cb != nil && cb.OnConnect != nil {
			cb.OnConnect(serviceName)
		}
		if cb != nil && cb.OnDisconnect != nil {
			defer cb.OnDisconnect(serviceName)
		}

		// Bidirectional proxying.
		done := make(chan struct{})

		// Client → Backend
		go func() {
			defer close(done)
			pumpMessages(clientConn, backendConn)
		}()

		// Backend → Client
		pumpMessages(backendConn, clientConn)
		<-done
	})
}

func pumpMessages(src, dst *websocket.Conn) {
	for {
		mt, r, err := src.NextReader()
		if err != nil {
			_ = dst.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
		w, err := dst.NextWriter(mt)
		if err != nil {
			return
		}
		if _, err := io.Copy(w, r); err != nil {
			return
		}
		if err := w.Close(); err != nil {
			return
		}
	}
}
