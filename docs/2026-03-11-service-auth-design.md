# WorkFort Service Auth — Design Spec

## Overview

A unified authentication and authorization system for the WorkFort ecosystem, built on better-auth. All identity (users, agents, services) is owned by better-auth as the single source of truth. Services validate callers via a shared Go middleware package (`pkg/auth/`) that verifies JWTs and API keys. The WorkFort CLI proxy acts as a BFF (Backend for Frontend) — the browser never sees a JWT; tokens stay server-side.

## Constraints

- **SOC 2 alignment**: Tokens never reach the browser. The BFF pattern ensures `HttpOnly` session cookies are the only credential in the browser.
- **Apache 2.0 licensing**: The WorkFort CLI (including `pkg/auth/` and `@workfort/ui`) is Apache 2.0 licensed. Service remotes can be any license (proprietary, GPL, MIT).
- **Single identity source**: better-auth owns all identities. Services do not maintain their own user tables.
- **Backward compatibility**: Services can migrate incrementally. The shared middleware handles both JWT and API key validation transparently.

## Identity Model

Three identity types, all managed by better-auth:

| Type | Example | Auth Method | Provisioning |
|------|---------|-------------|-------------|
| **User** | `kazw` logging into the web UI | Session cookie → BFF converts to JWT | Self-registration or admin invite |
| **Agent** | A Hive MCP agent performing tasks | API key (Bearer token) | Manual via admin API now, Substrate later |
| **Service** | Sharkfin calling Nexus's API | Pre-shared API key (Bearer token) | Manual via admin API |

### Identity fields

Every identity carries these fields, regardless of type:

| Field | Purpose | Mutability |
|-------|---------|------------|
| `ID` | UUID, stable primary key | Immutable |
| `Username` | Unique handle (e.g., `kazw`, `agent-deploy-01`) | Rarely changed |
| `Name` | Full name (e.g., `Kaz Walker`, `Deploy Agent 01`) | Mutable |
| `DisplayName` | Preferred short name (e.g., `Kaz`, `Deploy Agent`) | Mutable |
| `Type` | `user`, `agent`, or `service` | Immutable |

**Usage guidance for services:**
- **Chat messages, @mentions, presence** → `DisplayName`
- **Profile pages, audit logs** → `Name`
- **Unique references, URLs, API identifiers** → `Username`
- **Database foreign keys, internal references** → `ID` (UUID)

### User schema extension

better-auth's default user has `id`, `email`, `name`. Extended with:

```typescript
user: {
  additionalFields: {
    username:    { type: "string", unique: true, required: true },
    displayName: { type: "string" },
    type:        { type: "string", defaultValue: "user" },
  },
}
```

## Token Flow

### Human users (BFF pattern)

The WorkFort CLI proxy acts as a BFF. The browser uses `HttpOnly` session cookies only. The proxy converts cookies to JWTs for backend services.

```
Browser                    CLI Proxy (BFF)              Auth Service           Backend Service
  │                            │                            │                      │
  │─── request ───────────────▶│                            │                      │
  │    (cookie auto-attached)  │── GET /api/auth/token ────▶│                      │
  │                            │◀── JWT ────────────────────│                      │
  │                            │── forward request ────────────────────────────────▶│
  │                            │   Authorization: Bearer <JWT>                     │
  │                            │◀─────────────────────────────────────── response ─│
  │◀── response ───────────────│                            │                      │
```

The proxy caches the JWT (15-min lifetime, refreshed before expiry). The browser never sees a JWT.

**Proxy route behavior:**
- `/api/auth/*` — **pass-through**. The proxy forwards requests transparently to the auth service (including cookies). No JWT conversion. This is how login, registration, and session checks work.
- `/api/{service}/*` (sharkfin, nexus, hive) — **BFF conversion**. The proxy reads the session cookie, converts it to a JWT, and forwards with `Authorization: Bearer`.

**Error handling:**
- **Auth service unreachable** (proxy can't convert cookie to JWT): return `502 Bad Gateway` with JSON error body.
- **Session cookie expired or invalid** (auth service returns 401 on token request): return `401 Unauthorized`, clear the session cookie to force re-login.
- **JWT cache expired and refresh fails**: same as auth service unreachable — `502`. The proxy does not serve stale JWTs.

### Agents and services

API keys are sent directly as Bearer tokens. No BFF involvement.

```
Agent/Service                          Backend Service
  │                                        │
  │── request ────────────────────────────▶│
  │   Authorization: Bearer <api-key>      │
  │                                        │── verify via auth service (cached)
  │◀── response ───────────────────────────│
```

### WebSocket upgrade

The BFF converts the session cookie to a JWT and attaches it during the upgrade handshake. The service validates the JWT and extracts the identity. No in-protocol identity step needed.

**JWT expiry on long-lived connections:** The JWT is validated only at upgrade time. Once the WebSocket is established, the identity is trusted for the lifetime of the connection. If the user's session is revoked, the service will not know until the connection drops and the client attempts to reconnect (at which point the upgrade will fail). This is an acceptable trade-off — WebSocket connections are inherently stateful and tied to a single TCP connection. Forcing periodic re-auth over WebSocket frames adds complexity without meaningful security gain, since the connection itself is the trust boundary.

```
Browser                    CLI Proxy (BFF)              Service (e.g., Sharkfin)
  │                            │                            │
  │─── WS upgrade ────────────▶│                            │
  │    (cookie auto-attached)  │── convert cookie → JWT     │
  │                            │── WS upgrade ─────────────▶│
  │                            │   Authorization: Bearer <JWT>
  │                            │◀── 101 Switching ──────────│
  │◀── 101 Switching ─────────│                            │
  │◀════ WebSocket frames ════▶│◀════ WebSocket frames ════▶│
```

### JWT claims

```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "username": "kazw",
  "name": "Kaz Walker",
  "display_name": "Kaz",
  "type": "user",
  "exp": 1741700000
}
```

## CLI Login Flow

The CLI is the primary login interface. Two methods:

### Email + password

CLI prompts directly in the terminal, authenticates against better-auth, receives a session. No browser needed.

### OAuth (GitHub, Google)

Uses the better-auth Device Authorization plugin (RFC 8628). CLI gets a device code, opens the browser to the provider's consent screen, polls for the token.

### Browser handoff (one-time token)

After authentication (either method), the CLI hands off the session to the browser:

1. CLI registers a one-time cryptographically random token with the auth service, tied to the session (short TTL, single-use)
2. CLI displays: `Logged in as kazw. Web UI: http://localhost:8080?token=<onetime>`
3. When the user opens that URL (or `workfort web` opens it automatically), the auth service validates the token, deletes it immediately, and sets a signed `HttpOnly` `SameSite` session cookie
4. The browser is now authenticated — the cookie persists across CLI restarts, browser restarts, and tab closes
5. The cookie lifetime matches the better-auth session expiry
6. If the session expires, user runs `workfort login` again

The CLI persists its own session credential (access token from better-auth) in `$XDG_DATA_HOME/workfort/session.json` for use by CLI commands that call service APIs directly (without the browser).

**Security properties of the one-time token:**
- Cryptographically random, generated per invocation, held in memory only
- Consumed on first use or expires after 60 seconds
- If leaked (browser history, logs), it's already been consumed and is useless
- Not a credential — just a correlation secret between the CLI process and the browser

```
workfort login
  ├── email + password → CLI prompts, authenticates directly
  │
  └── OAuth (--provider github) → device flow in browser
                                    │
                         ┌──────────┘
                         ▼
              CLI has a session
                         │
              register one-time token with auth service
                         │
              open/display http://localhost:8080?token=<onetime>
                         │
              browser validates token → gets persistent cookie
```

## better-auth Plugin Configuration

The auth service runs better-auth with these plugins:

| Plugin | Purpose |
|--------|---------|
| **JWT** | Issues JWTs from sessions, exposes JWKS endpoint for Go services |
| **Bearer** | Enables `Authorization: Bearer` header auth |
| **API Key** | Org-scoped API keys for agents and services |
| **Admin** | Programmatic identity creation |
| **Device Authorization** | RFC 8628 device flow for CLI OAuth login |
| **Organization** | Scopes agents and API keys to teams/orgs |

### JWT claims configuration

```typescript
jwt({
  jwt: {
    definePayload: async ({ user }) => ({
      sub: user.id,
      username: user.username,
      name: user.name,
      display_name: user.displayName,
      type: user.type ?? "user",
    }),
  },
})
```

### API key prefix convention

| Identity type | Prefix | Example |
|---------------|--------|---------|
| User (personal access token) | `wf_` | `wf_a8f3...` |
| Agent | `wf-agent_` | `wf-agent_b2c1...` |
| Service | `wf-svc_` | `wf-svc_d4e5...` |

## Shared Go Middleware (`pkg/auth/`)

Lives in the WorkFort CLI repo at `pkg/auth/`. All services import it. Apache 2.0 licensed.

### Capabilities

1. **HTTP middleware** — wraps any `http.Handler`, extracts and validates `Authorization: Bearer` header, populates request context with an `Identity`
2. **JWKS-based JWT validation** — fetches and caches public keys from better-auth's `/api/auth/jwks`, validates tokens locally (no auth server roundtrip per request)
3. **API key validation** — calls better-auth's `/api/auth/verify-api-key`, caches results briefly
4. **Context helpers** — `IdentityFromContext(ctx)` to retrieve the verified identity
5. **WebSocket support** — validates Bearer token from the upgrade request before accepting the connection

### Package structure

```
pkg/auth/
  identity.go      // Identity struct, Type constants
  middleware.go     // HTTP middleware (Authenticate handler wrapper)
  jwt.go           // JWKS fetching, JWT parsing and validation
  apikey.go        // API key verification via auth service
  context.go       // Context key, IdentityFromContext(), MustIdentity()
  options.go       // Config: auth service URL, JWKS refresh interval, cache TTL
```

### Identity struct

```go
type Identity struct {
    ID          string // UUID, stable primary key
    Username    string // unique handle (e.g., "kazw")
    Name        string // full name (e.g., "Kaz Walker")
    DisplayName string // preferred display name (e.g., "Kaz")
    Type        string // "user", "agent", "service"
}
```

### Usage by a service

```go
import (
    "github.com/Work-Fort/WorkFort/pkg/auth"
    "github.com/Work-Fort/WorkFort/pkg/auth/jwt"
    "github.com/Work-Fort/WorkFort/pkg/auth/apikey"
)

func main() {
    opts := auth.DefaultOptions("http://127.0.0.1:3000")
    jwtV, err := jwt.New(ctx, opts.JWKSURL, opts.JWKSRefreshInterval)
    akV := apikey.New(opts.VerifyAPIKeyURL, opts.APIKeyCacheTTL)
    authMiddleware := auth.NewFromValidators(jwtV, akV)

    mux := http.NewServeMux()
    mux.Handle("GET /v1/health", healthHandler)         // public
    mux.Handle("/v1/", authMiddleware(apiHandler))       // protected
    mux.Handle("POST /mcp", authMiddleware(mcpHandler))  // protected
}

// Inside a handler:
func handleListVMs(w http.ResponseWriter, r *http.Request) {
    id, _ := auth.IdentityFromContext(r.Context())
    // id.ID, id.Username, id.DisplayName, id.Type
}
```

### Caching strategy

| What | Where | TTL | Rationale |
|------|-------|-----|-----------|
| JWKS keys | In-memory | 5 min (configurable) | Keys rotate infrequently |
| API key verification | In-memory | 30 sec (configurable) | Reduce roundtrips to auth service. Trade-off: a revoked key remains valid for up to 30 seconds. Acceptable for the current scale; a push-based invalidation mechanism can be added later if needed. |
| JWT validation | N/A | N/A | Pure local crypto, no caching needed |

## Migration Guidance Per Service

### All services (common steps)

1. Add `github.com/Work-Fort/WorkFort/pkg/auth` as a dependency
2. Wrap HTTP handlers with `auth.New(opts)` middleware
3. Use `auth.IdentityFromContext(r.Context())` to get the caller's identity
4. Key on `Identity.ID` (UUID) for any per-user data storage, not username
5. Keep health check and OpenAPI/docs endpoints outside the middleware (public)

### Sharkfin

| Change | Detail |
|--------|--------|
| Drop `users` table | better-auth owns user identity. Remove user store implementations in both `pkg/infra/sqlite/users.go` and `pkg/infra/postgres/users.go`, and the users migration. |
| Remove `SessionManager` | No more in-memory token tracking. Remove `pkg/daemon/session.go`. |
| Remove `identify`/`register` WS step | Identity comes from the JWT validated during the WebSocket upgrade handshake. |
| Keep `roles` and `role_permissions` | Authorization (who can do what in chat) is still Sharkfin's concern. |
| Full schema migration to UUID keys | Sharkfin currently uses `int64` user IDs across all tables (users, channels, messages, read cursors, mentions, roles). All user-referencing foreign keys must migrate from `int64` to UUID strings (`Identity.ID`). This is a full schema migration, not just roles. Both SQLite (`pkg/infra/sqlite/`) and PostgreSQL (`pkg/infra/postgres/`) store implementations are affected. |
| MCP connections | Extract identity from `Authorization: Bearer` header via the shared middleware. |

### Nexus

| Change | Detail |
|--------|--------|
| Add auth middleware | Currently completely open. Wrap all REST and MCP endpoints. |
| Health check stays public | Register outside the middleware. |
| No user table needed | Read `Identity` from context for audit logging. |

### Hive

| Change | Detail |
|--------|--------|
| Replace `APIKeyAuth` middleware | Use the shared `pkg/auth` middleware instead of the custom `middleware.go`. |
| Replace `X-Agent-Id` extraction | Agent identity comes from `Identity` in context. `Identity.ID` replaces the old agent ID, `Identity.Type == "agent"` distinguishes agents. |
| Keep `AuthzService` | Permission system stays. Update to check permissions against `Identity.ID` instead of the old agent ID format. |
| Agent provisioning | Currently manual via better-auth admin API. Substrate will automate this later. |

## Future: OAuth 2.1 Client Credentials

The current design uses API keys for agents and services. When Substrate arrives and the system scales, the OAuth 2.1 Provider plugin can be added for standards-based machine-to-machine auth via `client_credentials` grant. The shared middleware already handles JWTs, so adding this path requires no service-side changes — only the auth service configuration and provisioning flow change.
