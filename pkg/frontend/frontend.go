package frontend

import (
	"encoding/json"
	"io/fs"
	"net/http"
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

	// File serving will be added in Task 2.
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
