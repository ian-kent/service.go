package producer

import "strings"

// FIXME rationalise with consumer/config.go

// Config represents the configuration required for a consumer service
type Config interface {
	KafkaBrokers() []string
}

type defaultConfig struct {
	KafkaBrokers string `env:"KAFKA_BROKERS" flag:"kafka-brokers" flagDesc:"Kafka brokers"`
}

// DefaultConfig is a default Config implementation
type DefaultConfig struct {
	defaultConfig
}

// KafkaBrokers implements Config.KafkaBrokers
func (c DefaultConfig) KafkaBrokers() []string {
	return strings.Split(c.defaultConfig.KafkaBrokers, ",")
}
