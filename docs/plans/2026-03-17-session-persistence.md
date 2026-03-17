# Session Persistence Across Page Reloads — Plan 11

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Detect existing authenticated sessions on page load so users don't have to sign in on every browser refresh.

**Architecture:** The BFF gets a new `GET /api/session` endpoint that checks the session cookie via `TokenConverter`. The shell probes this on startup and sets `needsAuth(false)` if the session is valid. No JWT touches the browser — the BFF validates server-side.

**Tech Stack:** Go `net/http` (BFF), SolidJS (Shell)

**Repo:** `scope/lead`

---

### Task 1: BFF Session Probe Endpoint

**Files:**
- Modify: `internal/infra/httpapi/handler.go`
- Modify: `internal/infra/httpapi/handler_test.go`

**Step 1: Write failing test**

Add to `internal/infra/httpapi/handler_test.go`:

```go
func TestHandler_SessionEndpoint_Authenticated(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()
	fort := newTestFort(tracker)

	// Need a real auth URL for token converter.
	authSvc, ok := tracker.ServiceByName("auth")
	if !ok {
		t.Fatal("auth service not found in tracker")
	}
	tc := httpapi.NewTokenConverterForTest(authSvc.URL, 5*time.Minute, 1*time.Minute)

	handler := httpapi.NewHandler(fort, tracker, tc, nil)

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
	if !resp.Authenticated {
		t.Fatal("expected authenticated: true")
	}
}

func TestHandler_SessionEndpoint_NoSession(t *testing.T) {
	tracker, cleanup := newTestTracker(t)
	defer cleanup()
	fort := newTestFort(tracker)

	handler := httpapi.NewHandler(fort, tracker, nil, nil)

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
		t.Fatal("expected authenticated: false")
	}
}
```

Note: The existing test infrastructure has a mock auth server that accepts `valid-session` as a cookie and returns a JWT. The `newTestTracker` sets this up. Read the existing BFF tests (around `TestHandler_BFFProxiesWithJWT`) to see the pattern.

**Step 2: Run test to verify it fails**

Run: `cd scope/lead && mise run test`
Expected: FAIL — `/api/session` returns 404.

**Step 3: Implement**

In `internal/infra/httpapi/handler.go`, add the session endpoint in `NewHandler` alongside the other shell endpoints:

```go
// Shell endpoints.
mux.HandleFunc("GET /api/services", servicesHandler(fort.Name, tracker))
mux.HandleFunc("GET /api/config", configHandler(fort.Name))
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

This calls `tc.Token(r)` which:
- Reads the session cookie
- Returns a cached JWT if available (no Passport call)
- Or calls Passport to exchange the cookie for a JWT (cache miss)
- Returns an error if no cookie, expired session, or Passport is down

If `tc` is nil (no auth service configured), `authenticated` is always false.

**Step 4: Run test to verify it passes**

Run: `cd scope/lead && mise run test`
Expected: All tests pass.

**Step 5: Commit**

```bash
git add internal/infra/httpapi/handler.go internal/infra/httpapi/handler_test.go
git commit -m "feat(bff): add /api/session endpoint for session detection"
```

---

### Task 2: Shell Session Probe on Startup

**Files:**
- Modify: `web/shell/src/lib/api.ts`
- Modify: `web/shell/src/stores/services.ts`
- Modify: `web/shell/src/app.tsx`

**Step 1: Add session probe to API layer**

In `web/shell/src/lib/api.ts`, add:

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

Note: The session endpoint is at `/forts/{fort}/api/session` because it goes through the fort dispatcher, which strips the prefix and forwards to the fort handler. The handler registers it at `/api/session`.

Wait — actually, `/api/session` is registered on the fort-level handler, not the root router. So the full path is `/forts/{fort}/api/session`. But this path goes through `fortDispatch` which strips `/forts/{fort}` and forwards to the fort handler. The SPA fallback (Issue 8 fix) catches non-`/api/` paths — `/api/session` starts with `/api/` so it goes to the fort handler. This should work.

**Step 2: Update services store**

In `web/shell/src/stores/services.ts`:

Replace the `needsAuth` initialization and update logic:

```typescript
import { checkSession } from '../lib/api';

// Auth state: starts true (assume unauthenticated).
// Probed on first poll via /api/session endpoint.
const [needsAuth, setNeedsAuth] = createSignal(true);

/** Called by sign-in/setup forms after successful authentication. */
export function clearAuthRequired(): void {
  setNeedsAuth(false);
}
```

In `startPolling`, after the first `fetchServices` call, probe the session:

```typescript
export function startPolling(fort: string): void {
  if (activeFort !== fort) {
    stopPolling();
    prevConnected = new Map();
    setServiceList([]);
    setConflictList([]);
  }
  activeFort = fort;

  fetchServices(fort).then((res) => {
    handlePollResult(res);

    // After first poll, probe for existing session if not in setup mode.
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

The session probe only runs on the first poll (not every 30s). If the session is valid, `needsAuth` flips to false and the sign-in form disappears. If not, it stays true and the user sees the sign-in form.

**Step 3: Verify build**

Run: `cd web/shell && pnpm build`
Expected: Build succeeds.

**Step 4: Commit**

```bash
git add web/shell/src/lib/api.ts web/shell/src/stores/services.ts
git commit -m "feat(shell): probe /api/session on startup to detect existing auth"
```

---

### Task 3: Integration Verification

1. Start all services (Passport, Sharkfin, Vite, BFF)
2. Sign in via the browser
3. Reload the page
4. The shell should NOT show the sign-in form — the session probe detects the existing cookie
5. Navigate to Chat — it should load with full permissions

Verify with Playwright:
- Navigate to BFF URL
- If sign-in shows, sign in
- Hard reload (Ctrl+R)
- Snapshot — should show the shell, not sign-in form
