# Session Persistence Across Page Reloads — Plan 11

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Detect existing authenticated sessions on page load so users don't have to sign in on every browser refresh.

**Architecture:** The BFF gets a new `GET /api/session` endpoint that validates the session cookie server-side via `TokenConverter`. The shell probes this endpoint on startup and sets `needsAuth(false)` if valid. No JWT touches the browser — validation happens entirely in the BFF.

**Tech Stack:** Go `net/http` (BFF), SolidJS (Shell), vitest/happy-dom (Shell tests if applicable)

**Repo:** `scope/lead`

---

### Task 1: BFF Session Endpoint — Authenticated Case

**Files:**
- Modify: `internal/infra/httpapi/handler.go`
- Modify: `internal/infra/httpapi/handler_test.go`

**Step 1: Write the failing test**

Add to `internal/infra/httpapi/handler_test.go`:

```go
func TestHandler_SessionEndpoint_Authenticated(t *testing.T) {
	// Set up a mock auth server that accepts "valid-session" cookie
	// and returns a JWT — same pattern as TestHandler_BFFProxyRouting.
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ui/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]any{
				"status": "unavailable",
				"name":   "auth",
				"label":  "Auth",
				"route":  "/auth",
			})
			return
		}
		// Token conversion: accept "valid-session", reject everything else.
		cookie, err := r.Cookie("better-auth.session_token")
		if err != nil || cookie.Value != "valid-session" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": "jwt-for-session-check"})
	}))
	defer authServer.Close()

	tracker := httpapi.NewServiceTracker([]string{authServer.URL})
	tracker.InitialProbe(context.Background())

	fort := domain.Fort{
		Name:     "local",
		Local:    true,
		Services: []domain.ConfigService{{URL: authServer.URL}},
	}

	tc := httpapi.NewTokenConverter(authServer.URL)
	handler := httpapi.NewHandler(fort, tracker, tc, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "valid-session"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var resp struct {
		Authenticated bool `json:"authenticated"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Authenticated {
		t.Fatal("expected authenticated: true with valid session cookie")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd scope/lead && mise run test`
Expected: FAIL — `/api/session` returns 404 (route not registered).

**Step 3: Implement the session handler**

In `internal/infra/httpapi/handler.go`, add the route in `NewHandler` next to the other shell endpoints:

```go
mux.HandleFunc("GET /api/session", sessionHandler(tc))
```

Add the handler function:

```go
func sessionHandler(tc *TokenConverter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authenticated := false
		if tc != nil {
			_, err := tc.Token(r)
			authenticated = err == nil
		}

		_ = json.NewEncoder(w).Encode(map[string]bool{
			"authenticated": authenticated,
		})
	}
}
```

**Step 4: Run test to verify it passes**

Run: `cd scope/lead && mise run test`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/infra/httpapi/handler.go internal/infra/httpapi/handler_test.go
git commit -m "feat(bff): add /api/session endpoint — authenticated case"
```

---

### Task 2: BFF Session Endpoint — Unauthenticated Cases

**Files:**
- Modify: `internal/infra/httpapi/handler_test.go`

**Step 1: Write failing tests for edge cases**

Add to `internal/infra/httpapi/handler_test.go`:

```go
func TestHandler_SessionEndpoint_NoCookie(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()
	fort := newTestFort(tracker)

	authSvc, _ := tracker.ServiceByName("auth")
	tc := httpapi.NewTokenConverter(authSvc.URL)
	handler := httpapi.NewHandler(fort, tracker, tc, nil)

	// Request with no session cookie.
	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Authenticated bool `json:"authenticated"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Authenticated {
		t.Fatal("expected authenticated: false with no cookie")
	}
}

func TestHandler_SessionEndpoint_InvalidCookie(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()
	fort := newTestFort(tracker)

	authSvc, _ := tracker.ServiceByName("auth")
	tc := httpapi.NewTokenConverter(authSvc.URL)
	handler := httpapi.NewHandler(fort, tracker, tc, nil)

	// Request with an invalid/expired session cookie.
	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "expired-garbage"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Authenticated bool `json:"authenticated"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Authenticated {
		t.Fatal("expected authenticated: false with invalid cookie")
	}
}

func TestHandler_SessionEndpoint_NoTokenConverter(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()
	fort := newTestFort(tracker)

	// Pass nil token converter — no auth service configured.
	handler := httpapi.NewHandler(fort, tracker, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "valid-session"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Authenticated bool `json:"authenticated"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Authenticated {
		t.Fatal("expected authenticated: false when no token converter")
	}
}

func TestHandler_SessionEndpoint_AlwaysReturns200(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()
	fort := newTestFort(tracker)
	handler := httpapi.NewHandler(fort, tracker, nil, nil)

	// Verify the endpoint always returns 200 OK, never 401.
	// The authenticated field in the body carries the auth state.
	// This prevents the browser's fetch from throwing on non-OK status.
	for _, cookie := range []string{"", "valid-session", "garbage"} {
		req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
		if cookie != "" {
			req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: cookie})
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("cookie=%q: expected 200, got %d", cookie, rec.Code)
		}
	}
}
```

**Step 2: Run tests to verify they pass**

These should already pass with the implementation from Task 1 — they test the same handler with different inputs.

Run: `cd scope/lead && mise run test`
Expected: PASS (all 4 new tests + all existing tests).

**Step 3: Commit**

```bash
git add internal/infra/httpapi/handler_test.go
git commit -m "test(bff): add edge case tests for /api/session endpoint"
```

---

### Task 3: Shell API — `checkSession` Function

**Files:**
- Modify: `web/shell/src/lib/api.ts`

**Step 1: Add the checkSession function**

```typescript
export async function checkSession(fort: string): Promise<boolean> {
  try {
    const res = await fetch(`/forts/${fort}/api/session`);
    if (!res.ok) return false;
    const body = await res.json();
    return body.authenticated === true;
  } catch {
    return false;
  }
}
```

Key design decisions:
- Returns `false` on any error (network failure, non-200, invalid JSON) — fail closed.
- Uses the fort-scoped path `/forts/{fort}/api/session` which routes through the fort dispatcher to the fort handler.
- No auth headers needed — the browser automatically sends the HttpOnly session cookie.

**Step 2: Commit**

```bash
git add web/shell/src/lib/api.ts
git commit -m "feat(shell): add checkSession API function"
```

---

### Task 4: Shell Store — Probe Session on Startup

**Files:**
- Modify: `web/shell/src/stores/services.ts`

**Step 1: Read the current file**

Read `web/shell/src/stores/services.ts` to understand the current `startPolling`, `needsAuth`, and `clearAuthRequired` implementation.

**Step 2: Update `startPolling` to probe session**

Import `checkSession`:

```typescript
import { fetchServices, checkSession, type ServiceInfo, type Conflict, type ServicesResponse } from '../lib/api';
```

In `startPolling`, after the first `fetchServices` resolves and `handlePollResult` runs, probe the session if not in setup mode:

```typescript
export function startPolling(fort: string): void {
  if (activeFort !== fort) {
    stopPolling();
    prevConnected = new Map();
    setServiceList([]);
    setConflictList([]);
    setNeedsAuth(true); // Reset on fort change.
  }
  activeFort = fort;

  fetchServices(fort).then((res) => {
    handlePollResult(res);

    // Probe for existing session if not in setup mode.
    if (!setupMode()) {
      checkSession(fort).then((authenticated) => {
        if (authenticated) setNeedsAuth(false);
      });
    }
  }).catch(console.error);

  intervalId = setInterval(() => {
    fetchServices(fort).then(handlePollResult).catch(console.error);
  }, POLL_INTERVAL);
}
```

Note: The session probe only runs on the first poll — not every 30 seconds. This avoids unnecessary Passport calls. If the session expires mid-use, the next BFF-proxied request will return 401 and the app can handle it then.

**Step 3: Remove the old `hasSessionCookie` function if still present**

Check if `hasSessionCookie` is still in the file. If so, delete it — it's been replaced by the server-side probe.

**Step 4: Verify build**

Run: `cd web/shell && pnpm build`
Expected: Build succeeds.

**Step 5: Commit**

```bash
git add web/shell/src/stores/services.ts
git commit -m "feat(shell): probe /api/session on startup to persist auth across reloads"
```

---

### Task 5: Integration Test with Playwright

**Step 1: Ensure all services are running**

- Passport: `passport.nexus:3000`
- Sharkfin: `systemctl --user status sharkfin` (should be active)
- Shell Vite: `localhost:5173`
- BFF: `127.0.0.1:16100`

**Step 2: Sign in**

Navigate to `http://127.0.0.1:16100`, sign in with `admin@workfort.dev` / `adminpass123!`.

**Step 3: Verify signed-in state**

Take a snapshot — should show the shell with service tabs, not the sign-in form.

**Step 4: Hard reload the page**

```javascript
await page.reload({ waitUntil: 'domcontentloaded' });
```

**Step 5: Verify session persists**

Take a snapshot after reload. The shell should show the service tabs, NOT the sign-in form. The session probe detected the existing cookie.

Wait a few seconds for the session probe to complete if needed.

**Step 6: Verify Chat still works after reload**

Click Chat. Take a snapshot. The Sharkfin UI should load with full admin permissions (channels visible, input bar present, no "no permission" messages).

**Step 7: Check for console errors**

Run `browser_console_messages` with level `error`. Should be zero real errors (favicon 404 is acceptable).

**Step 8: Test expired session (optional)**

Clear cookies via Playwright and reload. The sign-in form should appear since the session probe returns `authenticated: false`.

```javascript
await page.context().clearCookies();
await page.reload();
```

Snapshot — should show sign-in form.

**Step 9: Document results**

If all steps pass, the session persistence feature is verified. If any step fails, document the failure with screenshots and error messages.
