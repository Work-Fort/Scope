package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Work-Fort/Scope/internal/infra/httpapi"
)

func TestAuthProxy_RewritesCookiePath(t *testing.T) {
	// Auth backend sets a cookie with Path=/
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "better-auth.session_token",
			Value: "abc123",
			Path:  "/",
		})
		w.WriteHeader(200)
	}))
	defer backend.Close()

	proxy := httpapi.NewAuthProxy("auth", backend.URL, true, "", "local")

	req := httptest.NewRequest("GET", "/api/auth/session", nil)
	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected Set-Cookie header")
	}
	if cookies[0].Path != "/forts/local/" {
		t.Errorf("got cookie path %q, want %q", cookies[0].Path, "/forts/local/")
	}
}
