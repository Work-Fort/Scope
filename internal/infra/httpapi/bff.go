package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	sessionCookieName = "better-auth.session_token"
	tokenLifetime     = 15 * time.Minute
	refreshBefore     = 5 * time.Minute
)

var (
	errNoSession      = errors.New("bff: no session cookie")
	errSessionExpired = errors.New("bff: session expired")
	errAuthDown       = errors.New("bff: auth service unavailable")
)

type cachedToken struct {
	jwt    string
	expiry time.Time
}

// TokenConverter converts session cookies to JWTs by calling the auth service.
type TokenConverter struct {
	authURL       string
	client        *http.Client
	tokenLifetime time.Duration
	refreshBefore time.Duration

	mu     sync.RWMutex
	tokens map[string]cachedToken
}

// NewTokenConverter creates a token converter that calls authServiceURL
// to exchange session cookies for JWTs.
func NewTokenConverter(authServiceURL string) *TokenConverter {
	return &TokenConverter{
		authURL:       authServiceURL + "/v1/token",
		client:        &http.Client{Timeout: 5 * time.Second},
		tokenLifetime: tokenLifetime,
		refreshBefore: refreshBefore,
		tokens:        make(map[string]cachedToken),
	}
}

// NewTokenConverterForTest creates a converter with custom timing parameters.
func NewTokenConverterForTest(authServiceURL string, ttl, refresh time.Duration) *TokenConverter {
	return &TokenConverter{
		authURL:       authServiceURL + "/v1/token",
		client:        &http.Client{Timeout: 5 * time.Second},
		tokenLifetime: ttl,
		refreshBefore: refresh,
		tokens:        make(map[string]cachedToken),
	}
}

// Token extracts the session cookie from the request and returns a JWT.
func (tc *TokenConverter) Token(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", errNoSession
	}
	sessionVal := cookie.Value

	tc.mu.RLock()
	cached, ok := tc.tokens[sessionVal]
	tc.mu.RUnlock()
	if ok && time.Until(cached.expiry) > tc.refreshBefore {
		return cached.jwt, nil
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, tc.authURL, nil)
	if err != nil {
		return "", fmt.Errorf("bff: create request: %w", err)
	}
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionVal})

	resp, err := tc.client.Do(req)
	if err != nil {
		return "", errAuthDown
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		tc.mu.Lock()
		delete(tc.tokens, sessionVal)
		tc.mu.Unlock()
		return "", errSessionExpired
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bff: auth returned status %d", resp.StatusCode)
	}

	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("bff: decode response: %w", err)
	}

	if body.Token == "" {
		return "", fmt.Errorf("bff: auth returned empty token")
	}

	tc.mu.Lock()
	tc.tokens[sessionVal] = cachedToken{
		jwt:    body.Token,
		expiry: time.Now().Add(tc.tokenLifetime),
	}
	tc.mu.Unlock()

	return body.Token, nil
}
