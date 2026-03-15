# Module Federation + Tauri v2 PoC Results

**Date:** 2026-03-14
**Verdict:** GO

## Environment

- Rust: cargo 1.94.0, rustc 1.94.0
- Node: v24.14.0
- pnpm: 10.32.1
- Tauri CLI: 2.10.1 (via npx @tauri-apps/cli)
- Tauri core: 2.10.3
- @module-federation/vite: 1.13.0
- @module-federation/runtime: 0.8.12
- Vite: 6.4.1
- OS: Arch Linux (kernel 6.18.13), Wayland + WebKitGTK 2.50.6

## Test Results

### Test 1: `document.createElement('script')` — PASS

Dynamically creating a `<script>` element and appending it to the DOM works without any CSP issues. This is the mechanism Module Federation uses internally to load remote chunks.

### Test 2: Remote Module Load via Module Federation — PASS

- Remote app exposes `./greet` module via `@module-federation/vite`
- Host uses `@module-federation/runtime` `init()` + `loadRemote()` to dynamically load the remote
- The remote function `greet()` returned `"Hello from remote!"` successfully
- Works in both browser and Tauri webview

### Test 3: Running in Tauri — CONFIRMED

`window.__TAURI_INTERNALS__` is present, confirming execution inside the Tauri v2 webview (not a regular browser).

## CSP Settings

### With `"csp": null` (no restrictions) — PASS

Everything works out of the box.

### With restrictive CSP — PASS

The following CSP works:

```
default-src 'self';
script-src 'self' 'unsafe-inline' 'unsafe-eval' http://localhost:3001;
connect-src 'self' http://localhost:3001 ws://localhost:1420 ws://127.0.0.1:*;
style-src 'self' 'unsafe-inline'
```

Key CSP directives needed for Module Federation:
- `script-src 'unsafe-inline' 'unsafe-eval'` — required for MF's dynamic script injection and module evaluation
- `script-src http://localhost:3001` (or the remote origin) — required to load remote entry and chunks
- `connect-src` for the remote origin — required for fetching the MF manifest and module chunks

### Production CSP recommendation

```
default-src 'self';
script-src 'self' 'unsafe-inline' 'unsafe-eval' https://<remote-cdn>;
connect-src 'self' https://<remote-cdn>;
style-src 'self' 'unsafe-inline';
```

Note: `'unsafe-eval'` is needed because Module Federation uses `new Function()` or similar dynamic evaluation internally. If this is unacceptable, investigate MF's `scriptType: 'esm'` or `module` type entries, though this was not tested in this PoC.

## What Worked

1. Tauri v2 desktop webview loads and executes Module Federation remotes without issues
2. `document.createElement('script')` works in Tauri's WebKitGTK webview
3. Cross-origin script loading from localhost:3001 (remote) into localhost:1420 (host) works
4. `@module-federation/runtime` `init()` + `loadRemote()` pattern works correctly
5. Tauri IPC (`@tauri-apps/api/core` `invoke()`) works alongside Module Federation
6. First Rust compilation takes ~2 minutes; incremental rebuilds take ~5-12 seconds

## What Didn't Work / Issues

1. **DTS plugin WebSocket errors** (cosmetic only): The `@module-federation/dts-plugin` tries to connect to a WebSocket for type hints and fails. This is harmless and only affects DX type generation, not runtime behavior.
2. **`mf-manifest.json` not served in dev mode**: In Vite dev mode, the manifest is not available at `/mf-manifest.json`. The `@module-federation/vite` plugin uses Vite's dev server module graph instead. The runtime `init()` call with `entry: 'http://localhost:3001/mf-manifest.json'` still works because the plugin intercepts the request.

## Workarounds Required

None. Module Federation works out of the box with Tauri v2 when CSP is either disabled (`null`) or properly configured.

## Architecture Notes

- **Remote**: Standalone Vite dev server on port 3001, exposes modules via `@module-federation/vite` plugin
- **Host**: Vite dev server on port 1420, wrapped by Tauri. Uses `@module-federation/runtime` to dynamically load remotes at runtime via `registerRemotes()` / `init()` + `loadRemote()`
- **Tauri**: Wraps the host's dev server URL. The `devUrl` in `tauri.conf.json` points to `http://localhost:1420`
- In production, `frontendDist` would point to the Vite build output directory

## Console Output (from Tauri webview)

```
[PoC] createElement("script") test: PASS
[PoC] Remote greet(): Hello from remote!
[PoC] Remote module load: PASS
[PoC] Running in Tauri: true
```
