# Better-Auth Service — Setup Requirements

Handoff doc for the team setting up the better-auth service. This covers what the rest of the WorkFort ecosystem expects from the auth service. How you deploy it (Nexus VM, container, bare metal) is your call — this doc only covers the configuration contract.

**Reference:** `docs/2026-03-11-service-auth-design.md` for the full design rationale.

---

## 1. Plugins

Install and enable these better-auth plugins:

| Plugin | Purpose | Required by |
|--------|---------|-------------|
| **JWT** | Issues JWTs from sessions, exposes `/api/auth/jwks` | `pkg/auth/jwt` middleware (Go services) |
| **Bearer** | Enables `Authorization: Bearer` header auth | BFF proxy, CLI |
| **API Key** | API keys for agents and service-to-service auth | `pkg/auth/apikey` middleware |
| **Admin** | Programmatic identity creation (agents, services) | Hive agent provisioning, manual setup |
| **Device Authorization** | RFC 8628 device flow for CLI OAuth | `workfort login --provider github` |
| **Organization** | Scopes API keys to teams/orgs | Multi-tenant key management |

---

## 2. User Schema Extension

better-auth's default user has `id`, `email`, `name`. Extend with three additional fields:

```typescript
user: {
  additionalFields: {
    username:    { type: "string", unique: true, required: true },
    displayName: { type: "string" },
    type:        { type: "string", defaultValue: "user" },
  },
},
```

- **`username`** — unique handle (e.g., `kazw`, `agent-deploy-01`). Used for @mentions, URLs, API identifiers.
- **`displayName`** — preferred short name (e.g., `Kaz`). Used in chat messages, presence.
- **`type`** — one of `"user"`, `"agent"`, or `"service"`. Immutable after creation. Defaults to `"user"`.

All three identity types (users, agents, services) are stored in the same better-auth user table. The `type` field distinguishes them.

---

## 3. JWT Claims

Configure the JWT plugin to include our custom fields in the token payload:

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

The Go middleware (`pkg/auth/jwt`) parses these exact claim names. If you change them, every Go service breaks.

**JWT lifetime:** 15 minutes (the BFF proxy caches and refreshes tokens on this cadence).

---

## 4. Endpoints the Go Middleware Expects

The `pkg/auth/` middleware calls two endpoints. These are provided by better-auth's plugins — you shouldn't need to implement anything, just confirm they're exposed.

| Endpoint | Method | Plugin | Consumer |
|----------|--------|--------|----------|
| `/api/auth/jwks` | GET | JWT | `pkg/auth/jwt` fetches public keys for local JWT validation |
| `/api/auth/verify-api-key` | POST | API Key | `pkg/auth/apikey` verifies API keys |
| `/api/auth/token` | GET | JWT + Bearer | BFF proxy converts session cookie → JWT |
| `/api/auth/session` | GET | Core | Shell checks session on boot, CLI checks auth state |

### `/api/auth/verify-api-key` request/response format

Request:
```json
{ "key": "wf_a8f3..." }
```

Success response:
```json
{
  "valid": true,
  "key": {
    "userId": "550e8400-e29b-41d4-a716-446655440000",
    "metadata": {
      "username": "kazw",
      "name": "Kaz Walker",
      "display_name": "Kaz",
      "type": "user"
    }
  }
}
```

Failure response:
```json
{
  "valid": false,
  "error": "invalid api key"
}
```

**Important:** The `metadata` object on the API key must include `username`, `name`, `display_name`, and `type`. The Go middleware reads these to build the `Identity` struct. If metadata is missing, the identity will have empty fields.

---

## 5. API Key Prefixes

When creating API keys via the admin API, use these prefix conventions:

| Identity type | Prefix | Example |
|---------------|--------|---------|
| User (personal access token) | `wf_` | `wf_a8f3...` |
| Agent | `wf-agent_` | `wf-agent_b2c1...` |
| Service | `wf-svc_` | `wf-svc_d4e5...` |

Configure the API Key plugin's prefix if it supports it, or enforce this convention when provisioning keys through the admin API.

---

## 6. Session and Cookie Configuration

- **Cookie:** `HttpOnly`, `SameSite=Lax`, `Secure` in production
- **Session lifetime:** Your call, but something reasonable (7-30 days). The CLI re-authenticates via `workfort login` when it expires.
- **Cookie name:** Use better-auth's default (typically `better-auth.session_token`)

The BFF proxy reads the session cookie and forwards it to `/api/auth/token` to get a JWT. The browser only ever sees the cookie, never a JWT.

---

## 7. OAuth Providers

Configure at least:
- **GitHub** — primary OAuth provider
- **Google** — optional, nice to have

These are used via the Device Authorization plugin (RFC 8628). The CLI gets a device code, opens the browser to the provider's consent screen, and polls for the token.

---

## 8. Storage

SQLite is fine for the current scale. The auth service is the single source of truth for identity — no other service stores user records.

---

## 9. Initial Seed Data

On first setup, create:

1. **Admin user** — `type: "user"`, used for initial setup and provisioning
2. **Service identities** (one per service that does service-to-service calls):
   - Sharkfin: `type: "service"`, `username: "svc-sharkfin"`
   - Nexus: `type: "service"`, `username: "svc-nexus"`
   - Hive: `type: "service"`, `username: "svc-hive"`
3. **API keys** for each service identity (using the `wf-svc_` prefix)

Agent identities are created later via Hive (manual for now, automated by Substrate later).

---

## 10. Verification Checklist

Once running, verify these work:

- [ ] `GET /api/auth/jwks` returns a JSON Web Key Set
- [ ] Create a user with `username`, `displayName`, and `type` fields
- [ ] Log in, get a session cookie
- [ ] `GET /api/auth/token` (with session cookie) returns a JWT containing `sub`, `username`, `name`, `display_name`, `type` claims
- [ ] Create an API key with metadata containing `username`, `name`, `display_name`, `type`
- [ ] `POST /api/auth/verify-api-key` with the key returns the identity with metadata
- [ ] Device authorization flow works (if OAuth providers are configured)

The Go middleware test suite (`go test ./pkg/auth/...`) validates parsing of these formats. If you want to integration-test against a live auth service, point the JWT validator at your `/api/auth/jwks` and sign a real token.

---

## What You Don't Need to Worry About

- **BFF proxy logic** — that's in the CLI, not the auth service
- **Authorization (who can do what)** — each service handles its own permissions. The auth service only does authentication (who are you).
- **The `@workfort/ui` auth module** — that's a frontend wrapper around your session endpoint. It just calls `/api/auth/session`.
