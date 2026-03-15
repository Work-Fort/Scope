# Module Federation + Tauri v2 Compatibility Research Spike

**Date:** 2026-03-14
**Status:** NEEDS-TESTING

---

## Summary

Module Federation can likely work inside a Tauri v2 webview, but it is not a
documented or battle-tested combination. No public project was found shipping
this exact stack. The main risks are CSP restrictions and custom-protocol
origin handling. Recent fixes in the Module Federation runtime (late 2024)
removed the hard dependency on `unsafe-eval`, which was the biggest blocker.
A proof-of-concept is needed before committing to this architecture.

**Verdict: NEEDS-TESTING** -- feasible in principle, but requires a PoC to
validate CSP configuration, script injection via custom protocol, and chunk
loading through the Rust proxy.

---

## Evidence Found

### 1. No documented cases of MF + Tauri

- Searched GitHub issues in `module-federation/universe` and
  `module-federation/vite` for "tauri", "webview", "electron", "desktop".
  **Zero relevant results.** The only hits were dependency-bump PRs.
- Searched GitHub API (`"module federation" tauri webview`): no relevant
  issues or discussions.
- No blog posts, Stack Overflow answers, or community discussions found
  describing this combination.
- No public repositories found combining `@module-federation/vite` with
  Tauri.

### 2. Module Federation CSP issues (fixed in late 2024)

The MF runtime historically required `unsafe-eval` due to usage of
`new Function('return this')()` to resolve the global object. This was the
single largest CSP compatibility problem. Three key PRs resolved it:

| PR | Description | Merged |
|----|-------------|--------|
| [universe#3163](https://github.com/module-federation/universe/pull/3163) | Use `document.defaultView` instead of `new Function()` for global object | 2024-11-05 |
| [universe#3054](https://github.com/module-federation/universe/pull/3054) | Replace `new Function()` with `import()` for ESM entry loading | 2024-10 |
| [universe#3067](https://github.com/module-federation/universe/pull/3067) | Make Next.js plugin work with strict CSP | 2024-10 |

Related issues: [#3103](https://github.com/module-federation/universe/issues/3103),
[#3053](https://github.com/module-federation/universe/issues/3053),
[#2759](https://github.com/module-federation/universe/issues/2759),
[#2015](https://github.com/module-federation/universe/issues/2015).

**Implication:** With MF runtime >= late 2024 versions, `unsafe-eval` is no
longer required. This removes the most severe CSP blocker for Tauri.

### 3. MF script loading mechanism

From the MF SDK source (`packages/sdk/src/dom.ts`), the runtime loads
remotes via:

```js
const script = document.createElement('script');
script.src = remoteEntryUrl;
document.head.appendChild(script);
```

This is standard DOM script injection. It does **not** use `eval()` or
`new Function()` for the script loading itself (only for global resolution,
which is now fixed). The script tag approach works under CSP as long as
`script-src` allows the origin serving the remote entry.

For chunk loading after the remote entry is resolved, MF uses dynamic
`import()` calls, which are natively supported by modern webview engines.

---

## CSP Requirements

### Tauri v2 CSP behavior

- Tauri v2 **does not set a CSP by default** (`"csp": null` in the default
  `SecurityConfig`). When no CSP is configured, the webview applies its
  platform-default policy (generally permissive).
- When a CSP **is** configured in `tauri.conf.json`, Tauri automatically
  injects nonces for its own initialization scripts and modifies the policy
  (unless `dangerousDisableAssetCspModification: true`).
- In **isolation mode**, Tauri enforces a strict CSP that blocks inline
  scripts without matching nonces/hashes. This caused
  [issue #10956](https://github.com/tauri-apps/tauri/issues/10956).

### Required CSP for MF in Tauri

If you set a CSP (recommended for production), you need at minimum:

```json
{
  "security": {
    "csp": "default-src 'self'; script-src 'self' https://your-mf-server.com; connect-src 'self' https://your-mf-server.com; style-src 'self' 'unsafe-inline'"
  }
}
```

Key directives:

| Directive | Value | Reason |
|-----------|-------|--------|
| `script-src` | `'self'` + remote origin(s) | MF injects `<script>` tags pointing to remote entry URLs |
| `connect-src` | `'self'` + remote origin(s) | MF chunks loaded via `fetch()` / `import()` |
| `style-src` | `'self' 'unsafe-inline'` | Many UI frameworks inject inline styles |

**If remotes are proxied through the Rust backend** (same origin), then
`'self'` alone may suffice for `script-src` and `connect-src`, which is the
most secure option.

### What is NOT needed (post-2024 fixes)

- `'unsafe-eval'` -- no longer required by MF runtime
- `'unsafe-inline'` for scripts -- MF uses `src`-based script tags, not
  inline scripts

---

## Known Issues and Workarounds

### Issue 1: Custom protocol origin matching

Tauri serves content via `tauri://localhost` (Linux/macOS) or
`https://tauri.localhost` (Windows). The `'self'` CSP directive should match
these origins, but this needs verification. If `'self'` does not resolve
correctly under the custom protocol, you may need to explicitly allowlist
the protocol:

```
script-src 'self' tauri: https://tauri.localhost
```

**Workaround:** Use `devUrl` pointing to a standard `http://localhost:PORT`
dev server during development, which avoids custom protocol issues entirely.
For production, the Rust proxy serves remotes on the same origin.

### Issue 2: CORS with proxied remotes

When the Rust proxy fetches remote entries from upstream servers and serves
them to the webview, standard CORS rules apply. The proxy must set
appropriate `Access-Control-Allow-Origin` headers on responses, or the
webview will block the requests.

**Workaround:** Since the Rust proxy controls the response headers, it can
inject correct CORS headers. If all remotes are served from the same origin
as the shell, CORS is not an issue.

### Issue 3: Tauri isolation mode

Isolation mode adds an iframe-based security boundary with a strict CSP.
This may interfere with MF's script injection into `document.head`.

**Workaround:** Avoid isolation mode if using MF, or test thoroughly. The
brownfield pattern (default) is more compatible.

### Issue 4: Script loading timing

Tauri injects its own initialization scripts before the app loads. MF's
`registerRemotes()` must run after Tauri's IPC bridge is ready. This is
normally handled by the framework (Svelte/React) mounting after DOMContentLoaded.

**Workaround:** Ensure `registerRemotes()` is called from application code
(e.g., in `+layout.ts` or `main.ts`), not from a script that races with
Tauri initialization.

### Issue 5: Mobile webview differences

On Android (WebView) and iOS (WKWebView), dynamic `import()` support and
CSP enforcement may differ from desktop. Android WebView historically had
limitations with ES modules.

**Workaround:** Test on actual mobile devices. Consider using the `var`
library type for remote entries instead of `module` to maximize
compatibility.

---

## Recommendation

### Approach: Proxied remotes through Rust backend (same-origin)

The safest architecture for MF in Tauri is:

1. **All remote entry URLs are relative paths** (e.g., `/remotes/app1/remoteEntry.js`)
2. **The Rust backend proxies** these requests to the actual remote servers,
   adding auth headers
3. **CSP uses `'self'` only** for `script-src` and `connect-src`
4. **No isolation mode** (use brownfield pattern)

This avoids all CORS and cross-origin CSP issues because everything appears
same-origin to the webview.

### Required PoC validation

Before committing to this architecture, build a minimal PoC that validates:

- [ ] `document.createElement('script')` works with `src` pointing to a
      proxied remote entry URL under the Tauri custom protocol
- [ ] Dynamic `import()` of federated chunks works after remote entry loads
- [ ] `registerRemotes()` from `@module-federation/vite` successfully
      initializes in the Tauri webview
- [ ] CSP with `script-src 'self'` does not block MF script injection
- [ ] Works on all target platforms (Linux, macOS, Windows)
- [ ] Works on mobile webviews if mobile support is planned

### Risk level: Medium

The individual pieces (MF script injection, Tauri webview, dynamic imports)
are all well-understood. The risk is in their intersection, which is
undocumented. The same-origin proxy pattern mitigates the most dangerous
failure modes (CSP blocks, CORS errors). The PoC should take 1-2 days.

### Version requirements

- `@module-federation/vite` >= 0.8.x (post-CSP fixes)
- `@module-federation/runtime` >= 0.8.x (post-CSP fixes)
- Tauri v2.x (v1 used different protocol handling)
