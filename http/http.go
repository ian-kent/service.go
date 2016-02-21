package http

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

// DefaultClient holds a default HTTP client
//
// The client can be customised per-request by passing one in
// as the argument to Call
var DefaultClient = http.DefaultClient

// Service represents a HTTP service
type Service interface {
	URL() string
}

// BasicService is a Service which has only a URL
type BasicService string

// URL returns the service URL
func (bs BasicService) URL() string {
	return string(bs)
}

var _ Caller = BasicService("")

// Caller ...
type Caller interface {
	Call(...interface{}) Requester
}

// Requester ...
type Requester interface {
	ForwardHeaders(...string) Requester

	Delete(string) Resulter
	Get(string) Resulter
	Head(string) Resulter
	Options(string) Resulter
	Patch(string) Resulter
	Post(string) Resulter
	Put(string) Resulter
	Trace(string) Resulter
}

// Resulter ...
type Resulter interface {
	Do() (*http.Response, error)
	Result(dest interface{}) (*http.Response, error)
	Stream(delim byte, c chan StreamResulter) (*http.Response, error)
}

// StreamResulter ...
type StreamResulter interface {
	Bytes() []byte
	Error() error
	Result(dest interface{}) error
}

// AuthHeader ...
type AuthHeader interface {
	String() string
}

// Headers ...
type Headers map[string]string

// Token is a type alias for a string containing an OAuth2 token
type Token string

func (t Token) String() string {
	return "Bearer " + string(t)
}

// Key is a type alias for a string containing an API key
type Key string

func (k Key) String() string {
	auth := string(k) + ":"
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// DefaultForwardHeaders is a slice of default header names which
// are forwarded from the incoming request.
//
// Forwarded headers can be customised per-request using
// Requester.ForwardHeaders
var DefaultForwardHeaders = []string{
	"X-Forwarded-For", "X-Forwarded-Proto",
	"X-Forwarded-Host", "X-Request-Id",
}

type requester struct {
	serviceCall
	method string
	path   string
}

type streamResult struct {
	requester requester
	response  *http.Response
	bytes     []byte
	err       error
}

func (sr streamResult) Bytes() []byte {
	return sr.bytes
}

func (sr streamResult) Error() error {
	return sr.err
}

func (sr streamResult) Result(dest interface{}) (err error) {
	return sr.requester.unmarshal(bytes.NewReader(sr.bytes), sr.response.Header.Get("Content-Type"), dest)
}

type serviceCall struct {
	service        Service
	request        *http.Request
	body           io.Reader
	headers        Headers
	auth           AuthHeader
	client         *http.Client
	forwardHeaders []string
	unmarshaler    func([]byte, interface{}) error
}

func (r requester) unmarshal(rdr io.Reader, contentType string, dest interface{}) (err error) {
	if dest != nil {
		u := r.unmarshaler

		if u == nil {
			path := strings.ToLower(r.path)
			switch {
			case contentType == "application/json",
				strings.HasSuffix(contentType, "+json"),
				strings.HasSuffix(path, ".json"):
				u = json.Unmarshal
			case contentType == "application/xml",
				strings.HasSuffix(contentType, "+xml"),
				strings.HasSuffix(path, ".xml"):
				u = xml.Unmarshal
			case contentType == "text/x-yaml",
				contentType == "application/yaml",
				strings.HasSuffix(contentType, "+yaml"),
				strings.HasSuffix(path, ".yml"),
				strings.HasSuffix(path, ".yaml"):
				u = yaml.Unmarshal
			}
		}

		if u != nil {
			b, err := ioutil.ReadAll(rdr)
			if err != nil {
				err = fmt.Errorf("http: error reading from reader: %s", err)
				return err
			}

			if rc, ok := rdr.(io.Closer); ok {
				err = rc.Close()
				if err != nil {
					err = fmt.Errorf("http: error closing reader: %s", err)
				}
			}

			err = u(b, dest)
			if err != nil {
				err = fmt.Errorf("http: error unmarshaling response body: %s", err)
			}
		}
	}

	return
}

// Do ...
func (r requester) Do() (*http.Response, error) {
	req, err := http.NewRequest(r.method, r.serviceCall.service.URL()+r.path, nil)
	if err != nil {
		return nil, err
	}

	if r.serviceCall.request != nil {
		for _, hdr := range r.forwardHeaders {
			if h := r.serviceCall.request.Header.Get(hdr); len(h) > 0 {
				req.Header.Set(hdr, h)
			}
		}
	}

	if r.serviceCall.auth != nil {
		req.Header.Set("Authorization", r.serviceCall.auth.String())
	}

	cli := DefaultClient
	if r.client != nil {
		cli = r.client
	}

	return cli.Do(req)
}

// Stream ...
func (r requester) Stream(delim byte, c chan StreamResulter) (*http.Response, error) {
	res, err := r.Do()
	if err != nil {
		return res, err
	}

	go func(res *http.Response) {
		rdr := bufio.NewReader(res.Body)
		for {
			b, err := rdr.ReadBytes(delim)
			c <- streamResult{r, res, b, err}
			if err != nil {
				if err == io.EOF {
					err = res.Body.Close()
				}
				break
			}
		}
	}(res)

	return res, err
}

// Result ...
func (r requester) Result(dest interface{}) (*http.Response, error) {
	res, err := r.Do()
	if err != nil {
		return res, err
	}

	err = r.unmarshal(res.Body, res.Header.Get("Content-Type"), dest)

	return res, err
}

func (r *requester) ForwardHeaders(hdrs ...string) Requester {
	r.forwardHeaders = hdrs
	return r
}

// Call ...
func (bs BasicService) Call(args ...interface{}) Requester {
	var body io.Reader
	var headers Headers
	var auth AuthHeader
	var r *http.Request
	var cli *http.Client
	var unmarshaler func([]byte, interface{}) error

	for _, arg := range args {
		switch arg.(type) {
		case io.Reader:
			if body != nil {
				panic("cannot provide body multiple times")
			}
			body = arg.(io.Reader)
		case Headers:
			a := arg.(Headers)
			if headers != nil {
				for k, v := range a {
					headers[k] = v
				}
			} else {
				headers = a
			}
		case AuthHeader:
			if auth != nil {
				panic("cannot provide auth header multiple times")
			}
			auth = arg.(AuthHeader)
		case *http.Request:
			if r != nil {
				panic("cannot provide multiple requests")
			}
			r = arg.(*http.Request)
		case *http.Client:
			if cli != nil {
				panic("cannot provide multiple clients")
			}
			cli = arg.(*http.Client)
		case func([]byte, interface{}) error:
			if unmarshaler != nil {
				panic("cannot provide multiple unmarshalers")
			}
			unmarshaler = arg.(func([]byte, interface{}) error)
		default:
			panic(fmt.Sprintf("invalid parameter type: %s", reflect.TypeOf(arg).Name()))
		}
	}

	fwHdrs := append([]string{}, DefaultForwardHeaders...)
	return &requester{serviceCall: serviceCall{bs, r, body, headers, auth, cli, fwHdrs, unmarshaler}}
}

func (r *requester) Delete(path string) Resulter {
	r.method = "DELETE"
	r.path = path
	return r
}

func (r *requester) Get(path string) Resulter {
	r.method = "GET"
	r.path = path
	return r
}

func (r *requester) Head(path string) Resulter {
	r.method = "HEAD"
	r.path = path
	return r
}

func (r *requester) Options(path string) Resulter {
	r.method = "OPTIONS"
	r.path = path
	return r
}

func (r *requester) Patch(path string) Resulter {
	r.method = "PATCH"
	r.path = path
	return r
}

func (r *requester) Post(path string) Resulter {
	r.method = "POST"
	r.path = path
	return r
}

func (r *requester) Put(path string) Resulter {
	r.method = "PUT"
	r.path = path
	return r
}

func (r *requester) Trace(path string) Resulter {
	r.method = "TRACE"
	r.path = path
	return r
}
