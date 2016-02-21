package main

import (
	"os"

	"github.com/ian-kent/gofigure"
	"github.com/ian-kent/service.go"
	"github.com/ian-kent/service.go/log"
)

var cfg *config

type config struct {
	service.DefaultAPIConfig

	Example string `env:"EXAMPLE" flag:"example" flagDesc:"An example configuration item"`
}

func (c config) Namespace() string { return "service-namespace" }

func configure() config {
	if cfg != nil {
		return *cfg
	}

	cfg = new(config)
	if err := gofigure.Gofigure(cfg); err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	log.Namespace = cfg.Namespace()

	return *cfg
}
