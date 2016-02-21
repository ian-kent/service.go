package session

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ian-kent/service.go/log"
)

// Config is the session config
type Config interface {
	Name() string
	Secret() []byte
}

type store struct {
	*sessions.CookieStore
	name string
}

func (s store) get(r *http.Request) (se *sessions.Session, e error) {
	log.TraceR(r, "getting session", nil)
	se, e = s.CookieStore.Get(r, s.name)
	if e != nil {
		log.ErrorR(r, e, nil)
		return
	}
	log.TraceR(r, "got session", log.Data{"session": se})
	return se, e
}

var s *store

// Init initialises the session store
func Init(cfg Config) {
	log.Debug("initialising session storage", log.Data{"name": cfg.Name(), "secret": cfg.Secret()})
	s = &store{sessions.NewCookieStore(cfg.Secret()), cfg.Name()}
}

// Get returns a session from the session store
func Get(r *http.Request) (*sessions.Session, error) {
	if s == nil {
		panic("session: s is nil")
	}
	return s.get(r)
}
