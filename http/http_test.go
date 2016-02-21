package http

import (
	"io"
	"log"
	"net/http"
	"os"
	"testing"
)

var exampleAPI = BasicService("https://some-api")

func TestResult(t *testing.T) {
	req := &http.Request{
		Header: make(http.Header),
	}
	req.Header.Set("X-Request-Id", "test1234")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")

	key := Key(os.Getenv("API_KEY"))

	var m struct {
		Example string `json:"example"`
	}

	res, err := exampleAPI.Call(req, key).Get("/some-url").Result(&m)
	if err != nil {
		t.Error(err)
		return
	}

	log.Printf("%+v", res)
	log.Printf("%+v", m)
}

func TestStream(t *testing.T) {
	req := &http.Request{
		Header: make(http.Header),
	}
	req.Header.Set("X-Request-Id", "test1234")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")

	key := Key(os.Getenv("API_KEY"))

	c := make(chan StreamResulter)

	res, err := exampleAPI.Call(req, key).Get("/some-url").Stream('\n', c)
	if err != nil {
		t.Error(err)
		return
	}

	log.Printf("%+v", res)

	var m struct {
		Example string `json:"example"`
	}

	for {
		r := <-c

		if err := r.Result(&m); err != nil {
			t.Error(err)
			return
		}

		log.Printf("%+v", m)

		if err := r.Error(); err != nil {
			if err != io.EOF {
				t.Error(err)
				return
			}
			break
		}
	}
}
