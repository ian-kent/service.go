package static

import (
	"errors"
	"mime"
	"net/http"
	"strings"

	"github.com/gorilla/pat"
	"github.com/ian-kent/service.go/log"
	"github.com/ian-kent/service.go/web/render"
)

// Register creates routes for each static resource
func Register(config render.Config, r *pat.Router) {
	log.Debug("registering not found handler for static package", nil)
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		render.HTML(w, http.StatusNotFound, "error", render.DefaultVars(req, map[string]interface{}{"error": "Page not found"}))
	})

	log.Debug("registering static content handlers for static package", nil)
	for _, file := range config.AssetNames()() {
		if strings.HasPrefix(file, "static/") {
			path := strings.TrimPrefix(file, "static")
			log.Trace("registering handler for static asset", log.Data{"path": path})

			var mimeType string
			switch {
			case strings.HasSuffix(path, ".css"):
				mimeType = "text/css"
			case strings.HasSuffix(path, ".js"):
				mimeType = "application/javascript"
			default:
				mimeType = mime.TypeByExtension(path)
			}

			log.Trace("using mime type", log.Data{"type": mimeType})

			r.Path(path).Methods("GET").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if b, err := config.Asset()("static" + path); err == nil {
					w.Header().Set("Content-Type", mimeType)
					w.Header().Set("Cache-control", "public, max-age=259200")
					w.WriteHeader(200)
					w.Write(b)
					return
				}
				// This should never happen!
				log.ErrorR(req, errors.New("it happened ¯\\_(ツ)_/¯"), nil)
				r.NotFoundHandler.ServeHTTP(w, req)
			})
		}
	}
}
