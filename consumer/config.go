package consumer

// FIXME rationalise with producer/config.go

import (
	"strings"
	"time"

	"github.com/Shopify/sarama"
)

// Config represents the configuration required for a consumer service
type Config interface {
	ConsumerGroup() string
	InitialOffset() int64
	ProcessingTimeout() time.Duration
	ZookeeperURL() string
	ZookeeperChroot() string
	Topics() []string
}

type defaultConfig struct {
	ZookeeperURL    string `env:"ZOOKEEPER_URL" flag:"zookeeper-url" flagDesc:"Zookeeper URL"`
	Topics          string `env:"TOPICS" flag:"topics" flagDesc:"Topics"`
	Timeout         int    `env:"TIMEOUT" flag:"timeout" flagDesc:"Timeout"`
	ZookeeperChroot string `env:"ZOOKEEPER_CHROOT" flag:"zookeeper-chroot" flagDesc:"Zookeeper Chroot"`
	ConsumerGroup   string `env:"CONSUMER_GROUP" flag:"consumer-group" flagDesc:"Consumer group"`
}

// DefaultConsumerConfig is a default Config implementation
type DefaultConsumerConfig struct {
	defaultConfig
}

// ConsumerGroup implements Config.ConsumerGroup
func (c DefaultConsumerConfig) ConsumerGroup() string { return c.defaultConfig.ConsumerGroup }

// ProcessingTimeout implements Config.ProcessingTimeout
func (c DefaultConsumerConfig) ProcessingTimeout() time.Duration {
	return time.Duration(c.defaultConfig.Timeout) * time.Second
}

// ZookeeperURL implements Config.ZookeeperURL
func (c DefaultConsumerConfig) ZookeeperURL() string { return c.defaultConfig.ZookeeperURL }

// Topics implements Config.Topics
func (c DefaultConsumerConfig) Topics() []string { return strings.Split(c.defaultConfig.Topics, ",") }

// ZookeeperChroot implements Config.ZookeeperChroot
func (c DefaultConsumerConfig) ZookeeperChroot() string {
	return c.defaultConfig.ZookeeperChroot
}

// InitialOffset implements Config.InitialOffset
func (c DefaultConsumerConfig) InitialOffset() int64 { return sarama.OffsetOldest }
