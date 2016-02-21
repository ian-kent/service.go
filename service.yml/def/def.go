package def

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"gopkg.in/yaml.v1"
)

// ServiceWithResolvedDeps ...
type ServiceWithResolvedDeps struct {
	Service
	Deps []ServiceWithResolvedDeps `yaml:"-"`
}

// Service ...
type Service struct {
	Name    string `yaml:"name"`
	Source  string `yaml:"source"`
	Version string `yaml:"version"`

	Targets map[string]string `yaml:"targets"`
	Tags    []string          `yaml:"tags"`

	Config []Config `yaml:"config"`
	Deps   []Dep    `yaml:"deps"`
}

// Config ...
type Config struct {
	Desc string `yaml:"desc"`
	Env  string `yaml:"env"`
	Flag string `yaml:"flag"`
	Type string `yaml:"type"`
}

// Dep ...
type Dep struct {
	Source  string `yaml:"source"`
	Version string `yaml:"version"`
}

// DefaultNames is a list of default service.yml filenames
var DefaultNames = []string{
	"service.yml", "service.json", ".service.yml", ".service.json",
	"svc.yml", "svc.json", ".svc.yml", ".svc.json",
}

// Load loads the service definition file
func Load(filename string) (svc Service, err error) {
	var b []byte
	b, err = ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	svc, err = LoadBytes(filename, b)

	return
}

// LoadBytes loads the service definition bytes
func LoadBytes(filename string, b []byte) (svc Service, err error) {
	switch {
	case strings.HasSuffix(filename, ".json"):
		err = json.Unmarshal(b, &svc)
	case strings.HasSuffix(filename, ".yml"):
		fallthrough
	case strings.HasSuffix(filename, ".yaml"):
		fallthrough
	default:
		err = yaml.Unmarshal(b, &svc)
	}

	return
}

// MustLoad loads the service definition file, but panics on error
func MustLoad(filename string) Service {
	return MustLoadOne(filename)
}

// LoadOne ...
func LoadOne(filenames ...string) (svc Service, err error) {
	for _, f := range filenames {
		svc, err = Load(f)
		if err == nil {
			return
		}
	}
	return
}

// MustLoadOne ...
func MustLoadOne(filenames ...string) Service {
	svc, err := LoadOne(filenames...)
	if err != nil {
		panic(err)
	}
	return svc
}

// Resolve resolves service dependencies
func (s Service) Resolve() (res *ServiceWithResolvedDeps, err error) {
	res = &ServiceWithResolvedDeps{s, nil}

	for _, dep := range s.Deps {
		name, b, err := ResolveSource(dep.Source)
		if err != nil {
			return res, err
		}
		svc, err := LoadBytes(name, b)
		if err != nil {
			return res, err
		}
		r, err := svc.Resolve()
		if err != nil {
			return res, err
		}
		res.Deps = append(res.Deps, *r)
	}

	return
}

// ResolveSource loads a source path
func ResolveSource(source string) (name string, b []byte, err error) {
	switch {
	case strings.HasPrefix(source, "../"):
		fallthrough
	case strings.HasPrefix(source, "./"):
		fallthrough
	case strings.HasPrefix(source, "/"):
		for _, d := range DefaultNames {
			b, err = ioutil.ReadFile(source + "/" + d)
			if err == nil {
				break
			}
		}
	case strings.HasPrefix(source, "github.com/"):
		var res *http.Response
		src := strings.TrimPrefix(source, "github.com/")

		for _, name = range DefaultNames {
			res, err = http.Get("https://raw.githubusercontent.com/" + src + "/master/" + name)
			if err != nil {
				continue
			}
			if res.StatusCode > 399 {
				err = fmt.Errorf("unexpected status: %d", res.StatusCode)
				continue
			}
			b, err = ioutil.ReadAll(res.Body)
			if err != nil {
				err = fmt.Errorf("error reading body: %s", err)
				continue
			}
			break
		}
	default:
		err = fmt.Errorf("unrecognised source type: %s", source)
	}
	return
}
