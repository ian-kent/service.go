// Package timeout implements a timeoutHandler

// Mostly borrowed from core net/http.

package timeout

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/ian-kent/service.go/log"
)

// DefaultHandler returns a Handler with a default timeout
func DefaultHandler(h http.Handler) http.Handler {
	return Handler(h, 1*time.Second, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.TraceR(req, "timed out", nil)
		w.WriteHeader(http.StatusRequestTimeout)
	}))
}

// Handler returns a Handler that runs h with the given time limit.
//
// The new Handler calls h.ServeHTTP to handle each request, but if a
// call runs for longer than its time limit, the handler responds with
// a 503 Service Unavailable error and the given message in its body.
// (If msg is empty, a suitable default message will be sent.)
// After such a timeout, writes by h to its ResponseWriter will return
// ErrHandlerTimeout.
func Handler(h http.Handler, dt time.Duration, fh http.Handler) http.Handler {
	f := func() <-chan time.Time {
		return time.After(dt)
	}
	return &handler{h, f, fh}
}

// ErrHandlerTimeout is returned on ResponseWriter Write calls
// in handlers which have timed out.
var ErrHandlerTimeout = errors.New("http: Handler timeout")

type handler struct {
	handler     http.Handler
	timeout     func() <-chan time.Time // returns channel producing a timeout
	failHandler http.Handler
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	done := make(chan bool, 1)
	tw := &writer{w: w}
	go func() {
		h.handler.ServeHTTP(tw, r)
		done <- true
	}()
	select {
	case <-done:
		return
	case <-h.timeout():
		tw.mu.Lock()
		defer tw.mu.Unlock()
		log.TraceR(r, "request timed out", nil)
		if !tw.wroteHeader {
			log.TraceR(r, "headers not written, calling failure handler", nil)
			h.failHandler.ServeHTTP(w, r)
		}
		tw.timedOut = true
	}
}

type writer struct {
	w http.ResponseWriter

	mu          sync.Mutex
	timedOut    bool
	wroteHeader bool
}

func (tw *writer) Header() http.Header {
	return tw.w.Header()
}

func (tw *writer) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.wroteHeader = true // implicitly at least
	if tw.timedOut {
		return 0, ErrHandlerTimeout
	}
	return tw.w.Write(p)
}

func (tw *writer) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut || tw.wroteHeader {
		return
	}
	tw.wroteHeader = true
	tw.w.WriteHeader(code)
}
