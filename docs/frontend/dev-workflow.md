# Frontend Development Workflow

Running service frontends during development requires two concurrent dev servers plus a fort configuration.

## Start the shell dev servers

Run both commands in separate terminals:

```bash
# Terminal 1: Go backend with SPA proxy to Vite
mise run dev:go

# Terminal 2: Vite dev server for shell SPA
cd web/shell && pnpm dev
```

- **Go backend** (`mise run dev:go`) binds to `http://127.0.0.1:16100` with the `--dev` flag. This tells the shell to proxy all SPA requests to the Vite dev server on port 5173.
- **Vite dev server** (`pnpm dev` in `web/shell`) runs on `http://localhost:5173` and serves the shell SPA with HMR support.

Both must be running before you load the shell in a browser. Loading `http://127.0.0.1:16100` will 404 if Vite is not ready.

## Configure a service for local development

Edit your fort's `workfort.toml` config file. Change the service URL to point to your local dev server:

```toml
[[forts.services]]
name = "nexus"
url = "http://localhost:3001"  # Your service dev server
```

The Go backend proxies `/forts/{fort}/api/{service}/` requests to this URL. If the service is local (same machine), your `workfort.toml` will have `local = true` under `[forts]`.

**Request path rewrite:** The Go proxy strips the `/api/{serviceName}` prefix before forwarding:
- Browser requests `GET /forts/myfort/api/nexus/ui/remoteEntry.js`
- Go proxies to `http://localhost:3001/ui/remoteEntry.js`

## Vite Module Federation build requirements

Your service's Vite dev server must emit the Module Federation `remoteEntry.js` file under the `/ui/` prefix. The shell loads this file when it discovers your service.

```
http://localhost:3001/ui/remoteEntry.js      # MF manifest — must exist
http://localhost:3001/ui/assets/chunk-*.js   # MF chunks
http://localhost:3001/ui/assets/*.css        # Styles
```

Standard Vite + MF plugin setup will place these in a `dist/ui/` directory when built, or serve them at that path in dev mode.

## Service discovery and startup

The shell polls `GET /forts/{fort}/api/services` every 30 seconds. The Go backend fetches `/ui/health` from your service for each configured service:

```bash
curl http://localhost:3001/ui/health
# Expected 200 response:
# {
#   "name": "nexus",
#   "label": "Nexus",
#   "route": "/nexus",
#   "ws_paths": []
# }
```

See [Service Frontend Contract](./service-contract.md) for the full health probe spec.

**Timeline:**
1. Shell polls `GET /forts/myfort/api/services` — up to 30s from now.
2. Go backend probes `http://localhost:3001/ui/health` — expects 200 + manifest.
3. Shell receives services array with your service marked `ui: true`.
4. Shell's JS polling loop calls `registerRemotes` to register the MF remote.
5. Service appears in the shell's sidebar.

If your service is slow to start, the first poll will find it unreachable. It will retry every 30 seconds until the health probe succeeds.

## Hot Module Replacement (HMR)

**Within the service remote:** Vite HMR works normally. Edit a component in your service, save, and the module hot-reloads in the shell's iframe context.

**Full remote reload:** Changing `remoteEntry.js` requires a page refresh because the shell caches the entry point URL (the remote is only registered once). Vite serves `remoteEntry.js` with `Cache-Control: no-cache`, so a hard refresh (Cmd+Shift+R / Ctrl+Shift+R) will pick up the new version.

## Troubleshooting

### Service not appearing in the shell

1. **Check the Go log output** — look for errors when the Go backend polls `/ui/health`.
2. **Manually test the health probe:**
   ```bash
   curl -i http://localhost:3001/ui/health
   ```
   Should return 200. If it's 503, the service is running but `remoteEntry.js` does not exist.

3. **Check the shell's service list** — in a browser console:
   ```js
   fetch('/api/forts')
     .then(r => r.json())
     .then(console.log)
   ```
   Confirm your fort is listed.

4. **Check the services endpoint:**
   ```js
   fetch('/forts/myfort/api/services')
     .then(r => r.json())
     .then(console.log)
   ```
   Confirm your service appears with `ui: true`.

### Module Federation load failure

**Browser console error:** `Shared version mismatch` or similar.

- Ensure your service declares shared dependencies with `import: false` to consume them from the shell, not bundle its own copy.
- See [Service Frontend Contract — MF shared singletons](./service-contract.md#mf-shared-singletons) for the required shared list.

### CORS errors in browser

**Symptom:** Requests from the shell to the service fail with CORS errors.

**Diagnosis:** All requests from the shell to services go through the Go proxy (`/forts/{fort}/api/{service}/...`). There are no direct cross-origin calls. If you see CORS errors, check:

1. Is the Go backend running with `--dev`?
2. Is the service URL in your fort config correct?
3. Is the service actually responding to `GET http://localhost:3001/ui/health`?

The Go proxy adds `Access-Control-Allow-Origin: *` to all responses, so CORS should not be an issue.

### Service shows `ui: false` and won't load

**Diagnosis:** The Go backend's health probe succeeded (service is reachable) but did not find `remoteEntry.js`.

**Fix:** Ensure your Vite build output (or dev server) includes `remoteEntry.js` at `/ui/remoteEntry.js`. If using the MF plugin, check your plugin config:

```js
// vite.config.ts
federation({
  name: 'myservice',
  filename: 'remoteEntry.js',  // Ensure this is set
  exposes: { ... },
  shared: { ... }
})
```

The file must exist at the configured path before the health probe runs, or the Go server will cache the negative result for the lifetime of that process.

---

## Related

- [Frontend Architecture](./architecture.md) — Request path details and the full discovery flow.
- [Service Frontend Contract](./service-contract.md) — Health probe spec and MF runtime requirements.
