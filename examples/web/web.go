package main

import (
	"net/http"

	"github.com/ian-kent/service.go"
	"github.com/ian-kent/service.go/handlers/healthcheck"
	"github.com/ian-kent/service.go/log"
	"github.com/ian-kent/service.go/web/handlers/static"
	"github.com/ian-kent/service.go/web/render"
	"github.com/ian-kent/service.go/web/session"
)

func main() {
	cfg := configure()
	svc := service.Web(cfg)

	render.Render = render.New(cfg)
	session.Init(cfg)
	static.Register(cfg, svc.Router())

	svc.Chain(render.WithCsrfHandler)
	svc.Chain(exampleMiddleware)

	healthcheck.Register(svc.Router(), "/healthcheck", func() bool {
		return true
	})

	svc.Router().Path("/").Methods("GET").HandlerFunc(exampleHandler)

	svc.Start()
}

func exampleHandler(w http.ResponseWriter, req *http.Request) {
	render.HTML(w, http.StatusOK, "example", render.DefaultVars(req, nil))
}

func exampleMiddleware(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.DebugR(req, "example", log.Data{
			"accept": req.Header.Get("Accept"),
			"flag":   configure().Example,
		})
		f.ServeHTTP(w, req)
	})
}
