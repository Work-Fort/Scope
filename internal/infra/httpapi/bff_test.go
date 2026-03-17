package httpapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Work-Fort/Scope/internal/infra/httpapi"
)

func TestTokenConverter_Success(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("better-auth.session_token")
		if err != nil || cookie.Value != "valid-session" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "jwt-token-123"})
	}))
	defer authServer.Close()

	tc := httpapi.NewTokenConverter(authServer.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/sharkfin/v1/channels", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "valid-session"})

	token, err := tc.Token(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "jwt-token-123" {
		t.Fatalf("expected jwt-token-123, got %q", token)
	}
}

func TestTokenConverter_NoCookie(t *testing.T) {
	tc := httpapi.NewTokenConverter("http://localhost:0")

	req := httptest.NewRequest(http.MethodGet, "/api/sharkfin/v1/channels", nil)
	_, err := tc.Token(req)
	if err == nil {
		t.Fatal("expected error for missing cookie")
	}
}

func TestTokenConverter_ExpiredSession_EvictsCache(t *testing.T) {
	calls := 0
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch calls {
		case 1:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "jwt-first"})
		case 2:
			w.WriteHeader(http.StatusUnauthorized)
		case 3:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "jwt-refreshed"})
		}
	}))
	defer authServer.Close()

	tc := httpapi.NewTokenConverterForTest(authServer.URL, 0, 0)

	req := httptest.NewRequest(http.MethodGet, "/api/sharkfin/v1/channels", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "evict-test"})

	tok, err := tc.Token(req)
	if err != nil {
		t.Fatalf("call 1: %v", err)
	}
	if tok != "jwt-first" {
		t.Fatalf("call 1: expected jwt-first, got %q", tok)
	}

	_, err = tc.Token(req)
	if err == nil {
		t.Fatal("call 2: expected error for expired session")
	}

	tok, err = tc.Token(req)
	if err != nil {
		t.Fatalf("call 3: %v", err)
	}
	if tok != "jwt-refreshed" {
		t.Fatalf("call 3: expected jwt-refreshed, got %q", tok)
	}
	if calls != 3 {
		t.Fatalf("expected 3 auth calls, got %d", calls)
	}
}

func TestTokenConverter_CacheHit(t *testing.T) {
	calls := 0
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "jwt-cached"})
	}))
	defer authServer.Close()

	tc := httpapi.NewTokenConverter(authServer.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/sharkfin/v1/channels", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "cache-test"})

	token1, err := tc.Token(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token2, err := tc.Token(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token1 != token2 {
		t.Fatalf("expected same token from cache")
	}
	if calls != 1 {
		t.Fatalf("expected 1 auth call (cached), got %d", calls)
	}
}

func TestTokenConverter_ExpiredEntry_Evicted(t *testing.T) {
	calls := 0
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"token": fmt.Sprintf("jwt-call-%d", calls)})
	}))
	defer authServer.Close()

	// Very short TTL so entries expire quickly; refreshBefore=0 so we'd serve from cache if not expired.
	tc := httpapi.NewTokenConverterForTest(authServer.URL, 50*time.Millisecond, 0)

	req := httptest.NewRequest(http.MethodGet, "/api/sharkfin/v1/channels", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "evict-session"})

	// First call — fetches from auth.
	tok1, err := tc.Token(req)
	if err != nil {
		t.Fatalf("call 1: %v", err)
	}
	if tok1 != "jwt-call-1" {
		t.Fatalf("call 1: expected jwt-call-1, got %q", tok1)
	}

	// Wait for expiry.
	time.Sleep(100 * time.Millisecond)

	// Verify CacheLen still has 1 entry (stale) before the eviction call.
	if n := tc.CacheLen(); n != 1 {
		t.Fatalf("expected 1 cached entry before eviction, got %d", n)
	}

	// Second call — expired entry should be evicted and re-fetched.
	tok2, err := tc.Token(req)
	if err != nil {
		t.Fatalf("call 2: %v", err)
	}
	if tok2 != "jwt-call-2" {
		t.Fatalf("call 2: expected jwt-call-2 (re-fetch), got %q", tok2)
	}
	if calls != 2 {
		t.Fatalf("expected 2 auth calls (eviction + re-fetch), got %d", calls)
	}
}

func TestTokenConverter_AuthUnreachable(t *testing.T) {
	tc := httpapi.NewTokenConverter("http://127.0.0.1:1")

	req := httptest.NewRequest(http.MethodGet, "/api/sharkfin/v1/channels", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "some-session"})

	_, err := tc.Token(req)
	if err == nil {
		t.Fatal("expected error for unreachable auth service")
	}
}
