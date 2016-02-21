package consumer

import (
	"log"
	"os"
	"testing"

	"github.com/Shopify/sarama"
)

func TestConsumer(t *testing.T) {
	sarama.Logger = log.New(os.Stdout, "", log.LstdFlags)

	cfg := defaultConfig{
		ZookeeperURL:    "localhost",
		Topics:          []string{"ikent-test"},
		Timeout:         60,
		ZookeeperChroot: "kafka",
		ConsumerGroup:   "ikent-test",
	}
	consumer := New(DefaultConfig{cfg})

	for event := range consumer.Start() {
		log.Printf("%+v", event)
		log.Printf("%+v", event.Value)
	}
}
