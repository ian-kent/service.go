package service

import (
	"net/http"
	"os"

	"github.com/ian-kent/service.go/handlers/requestID"
	"github.com/ian-kent/service.go/handlers/timeout"
	"github.com/ian-kent/service.go/log"

	"github.com/gorilla/pat"
	"github.com/justinas/alice"
)

// DefaultMiddleware is the default middleware used to create a service
var DefaultMiddleware = []alice.Constructor{
	requestID.Handler(20),
	log.Handler,
	timeout.DefaultHandler,
}

// Service represents a service
type Service interface {
	Chain(handler ...alice.Constructor)
	Start()
	Router() *pat.Router
}

type service struct {
	config HTTPConfig
	router *pat.Router
	chain  []alice.Constructor
	alice  *alice.Chain
}

// Web returns a new web service using the provided config
func Web(config WebConfig) Service {
	return HTTP(config)
}

// API returns a new API service using the provided config
func API(config APIConfig) Service {
	return HTTP(config)
}

// HTTP returns a new HTTP service using the provided config
func HTTP(config HTTPConfig) Service {
	log.Event("configuration", "", log.Data{"config": config})

	log.Namespace = config.Namespace()

	return &service{
		config: config,
		router: pat.New(),
	}
}

func (s *service) Start() {
	chain := alice.New(s.middleware()...).Then(s.router)

	bindAddr := s.config.BindAddr()
	certFile, keyFile := s.config.CertFile(), s.config.KeyFile()

	if len(certFile) > 0 && len(keyFile) > 0 {
		log.Debug("listening tls", log.Data{"addr": bindAddr, "cert": certFile, "key": keyFile})
		err := http.ListenAndServeTLS(bindAddr, certFile, keyFile, chain)
		if err != nil {
			log.Error(err, nil)
			os.Exit(1)
		}
		return
	}

	log.Debug("listening", log.Data{"addr": bindAddr})
	err := http.ListenAndServe(bindAddr, chain)
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
}

func (s *service) middleware() []alice.Constructor {
	middleware := append([]alice.Constructor{}, DefaultMiddleware...)
	middleware = append(middleware, s.chain...)
	return middleware
}

func (s *service) Chain(handler ...alice.Constructor) {
	s.chain = append(s.chain, handler...)
}

func (s *service) Router() *pat.Router {
	return s.router
}
