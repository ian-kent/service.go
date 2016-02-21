package healthcheck

import (
	"net/http"

	"github.com/gorilla/pat"
)

// Register registers a healthcheck route with the router
func Register(r *pat.Router, path string, healthy func() bool) {
	r.Path(path).Methods("GET").HandlerFunc(Handler(healthy))
}

// Handler is a healthcheck HTTP handler
func Handler(healthy func() bool) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if healthy != nil && !healthy() {
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(200)
	}
}
