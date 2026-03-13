package httpapi

import (
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewSPAHandler serves an embedded SPA filesystem. Requests for real files
// are served directly. All other paths fall back to index.html for
// client-side routing.
func NewSPAHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(fsys))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the file exists in the embedded FS.
		path := r.URL.Path
		if path == "/" {
			path = "index.html"
		} else if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}

		if _, err := fs.Stat(fsys, path); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback: serve index.html for SPA routing.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

// NewSPADevProxy returns a handler that proxies all requests to a Vite
// dev server. Used with the --dev flag.
func NewSPADevProxy(devURL string) http.Handler {
	target, _ := url.Parse(devURL)
	return httputil.NewSingleHostReverseProxy(target)
}
