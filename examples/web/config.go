package main

import (
	"os"

	"github.com/ian-kent/gofigure"
	"github.com/ian-kent/service.go"
	"github.com/ian-kent/service.go/examples/web/assets"
	"github.com/ian-kent/service.go/log"
)

var cfg *config

type config struct {
	service.DefaultWebConfig

	Example string `env:"EXAMPLE" flag:"example" flagDesc:"An example configuration item"`
}

func (c config) Namespace() string { return "service-namespace" }

func (c config) Asset() func(string) ([]byte, error) {
	return assets.Asset
}

func (c config) AssetNames() func() []string {
	return assets.AssetNames
}

func (c config) Config() interface{} {
	return c
}

func (c config) Name() string {
	return c.DefaultWebConfig.SessionName
}

func (c config) Secret() []byte {
	return []byte(c.DefaultWebConfig.SessionSecret)
}

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
