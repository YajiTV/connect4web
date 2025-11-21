package httphandler

import (
	"io/fs"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

// NewStaticHandler serves embedded static files with simple caching and content types
func NewStaticHandler(fsys fs.FS) http.Handler {
	fsHandler := http.FileServer(http.FS(fsys))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := strings.ToLower(filepath.Ext(r.URL.Path))

		// sets cache headers by asset type
		switch ext {
		case ".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".ico":
			w.Header().Set("Cache-Control", "public, max-age=604800")
		default:
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}

		// sets a content type if missing
		if w.Header().Get("Content-Type") == "" {
			if c := mime.TypeByExtension(ext); c != "" {
				w.Header().Set("Content-Type", c)
			} else {
				// guards against some servers not recognizing .css in unusual paths
				if strings.HasSuffix(strings.ToLower(path.Base(r.URL.Path)), ".css") {
					w.Header().Set("Content-Type", "text/css; charset=utf-8")
				}
			}
		}

		fsHandler.ServeHTTP(w, r)
	})
}
