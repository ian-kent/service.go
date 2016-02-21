package main

import (
	"net/http"

	"github.com/ian-kent/service.go"
	"github.com/ian-kent/service.go/handlers/healthcheck"
	"github.com/ian-kent/service.go/log"
)

func main() {
	svc := service.API(configure())

	svc.Chain(exampleMiddleware)

	healthcheck.Register(svc.Router(), "/healthcheck", func() bool {
		return true
	})

	svc.Router().Path("/").Methods("GET").HandlerFunc(exampleHandler)

	svc.Start()
}

func exampleHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(`{}`))
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
