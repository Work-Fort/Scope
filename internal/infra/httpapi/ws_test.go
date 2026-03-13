package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/Work-Fort/Scope/internal/infra/httpapi"
)

func TestWSProxy_WhitelistedPath(t *testing.T) {
	// Backend WS server that accepts upgrades.
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("backend upgrade error: %v", err)
			return
		}
		defer conn.Close()
		// Echo one message back.
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		_ = conn.WriteMessage(mt, msg)
	}))
	defer backend.Close()

	backendURL := "ws" + strings.TrimPrefix(backend.URL, "http")
	wsHandler := httpapi.NewWSProxy(backendURL, []string{"/ws", "/presence"}, "nexus")

	// Wrap in a test server so we can dial it.
	proxy := httptest.NewServer(wsHandler)
	defer proxy.Close()

	proxyURL := "ws" + strings.TrimPrefix(proxy.URL, "http") + "/api/nexus/ws"
	conn, resp, err := websocket.DefaultDialer.Dial(proxyURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v (resp: %v)", err, resp)
	}
	defer conn.Close()

	// Send and receive a message.
	if err := conn.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		t.Fatalf("write error: %v", err)
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if string(msg) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(msg))
	}
}

func TestWSProxy_NonWhitelistedPath(t *testing.T) {
	wsHandler := httpapi.NewWSProxy("ws://localhost:0", []string{"/ws"}, "nexus")
	proxy := httptest.NewServer(wsHandler)
	defer proxy.Close()

	proxyURL := "ws" + strings.TrimPrefix(proxy.URL, "http") + "/api/nexus/not-allowed"
	_, resp, err := websocket.DefaultDialer.Dial(proxyURL, nil)
	if err == nil {
		t.Fatal("expected error for non-whitelisted path")
	}
	if resp != nil && resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
