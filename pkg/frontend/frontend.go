package frontend

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"
)

// Handler returns an http.Handler that serves an embedded Vite build
// as a Module Federation remote.
//
// Routes registered under /ui/:
//
//	/ui/health           — 200 if remoteEntry.js exists, 503 if not
//	/ui/assets/*         — immutable content-hashed chunks (1yr cache)
//	/ui/remoteEntry.js   — federation entry point (no-cache)
//	/ui/*                — everything else (no-cache)
//
// The fsys must be rooted at the Vite build output directory
// (e.g., the result of fs.Sub(embedFS, "web/dist")).
func Handler(fsys fs.FS) http.Handler {
	hasRemoteEntry := fileExists(fsys, "remoteEntry.js")
	fileServer := http.StripPrefix("/ui/", http.FileServer(http.FS(fsys)))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ui/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if hasRemoteEntry {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
		}
	})

	mux.HandleFunc("/ui/", func(w http.ResponseWriter, r *http.Request) {
		// Set cache headers based on path pattern BEFORE calling the file server.
		// http.FileServer does not set or overwrite Cache-Control on success.
		// On error (404, 416), Go 1.23+ strips Cache-Control, which is correct —
		// we don't want to cache error responses.
		path := strings.TrimPrefix(r.URL.Path, "/ui/")
		if strings.HasPrefix(path, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}

		fileServer.ServeHTTP(w, r)
	})

	return mux
}

func fileExists(fsys fs.FS, name string) bool {
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	f.Close()
	return true
}
