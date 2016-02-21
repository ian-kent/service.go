package main

import (
	"github.com/ian-kent/service.go/consumer"
	"github.com/ian-kent/service.go/log"
)

func main() {
	consumer := consumer.New(configure())

	for event := range consumer.Start() {
		log.Debug("event", log.Data{"event": event})

		err := consumer.Commit(event)
		if err != nil {
			log.Error(err, nil)
		}
	}
}
