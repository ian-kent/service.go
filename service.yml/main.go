package main

import (
	"log"

	"github.com/ian-kent/service.go/service.yml/def"
)

func main() {
	svc := def.MustLoad("service.yml")

	r, err := svc.Resolve()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", r)
}
