// Package static serves the built SPA frontend (web/dist) from the rest
// server when WEB_DIST points to an existing directory. Routes registered by
// handler.RegisterHandlers keep precedence — only unmatched paths reach the
// static handler, which falls back to index.html for SPA history-mode routing.
package static

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// apiPrefix is the backend API route prefix; requests under it are never
// served by the static handler so API 404s stay JSON 404s.
const apiPrefix = "/api/"

// Handler returns an http.Handler serving dir, or nil when dir is empty or
// does not exist (pure API mode, unchanged behavior).
func Handler(dir string) http.Handler {
	if dir == "" {
		return nil
	}
	fs := os.DirFS(dir)
	if _, err := os.Stat(filepath.Join(dir, "index.html")); err != nil {
		return nil
	}
	fileServer := http.FileServer(http.FS(fs))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, apiPrefix) {
			http.NotFound(w, r)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = "index.html"
		}
		if _, err := fs.Open(name); err != nil {
			// Unknown path: SPA fallback so history-mode deep links work.
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}
