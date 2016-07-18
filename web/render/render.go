package render

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/schema"
	"github.com/ian-kent/htmlform"
	"github.com/ian-kent/service.go/log"
	"github.com/ian-kent/service.go/web/session"
	"github.com/justinas/nosurf"
	"github.com/unrolled/render"
	"gopkg.in/bluesuncorp/validator.v5"
)

// Render is a global instance of github.com/unrolled/render.Render
var Render Renderer

// Renderer encapsulates unrolled/render and the renderer config
type Renderer struct {
	*render.Render
	config Config
}

// Vars is a map of variables
type Vars map[string]interface{}

// Config is the renderer configuration
type Config interface {
	Asset() func(string) ([]byte, error)
	AssetNames() func() []string
	// Config is any config struct to make available as `Config` in the template
	Config() interface{}
}

// New creates a new instance of github.com/unrolled/render.Render
func New(config Config) Renderer {
	log.Debug("creating renderer", nil)

	return Renderer{render.New(render.Options{
		Asset:      config.Asset(),
		AssetNames: config.AssetNames(),
		//Delims:     render.Delims{Left: "[:", Right: ":]"},
		Layout: "layout",
		Funcs: []template.FuncMap{template.FuncMap{
			"map":      htmlform.Map,
			"ext":      htmlform.Extend,
			"fnn":      htmlform.FirstNotNil,
			"arr":      htmlform.Arr,
			"lc":       strings.ToLower,
			"uc":       strings.ToUpper,
			"join":     strings.Join,
			"safehtml": func(s string) template.HTML { return template.HTML(s) },
			"date": func(t *time.Time, layout string) string {
				return t.Format(layout)
			},
		}},
	}), config}
}

// NewFromRender returns a new Renderer using the supplied unrolled.Render
func NewFromRender(r *render.Render, config Config) Renderer {
	log.Debug("creating renderer", nil)

	return Renderer{r, config}
}

// HTML is an alias to github.com/unrolled/render.Render.HTML
func HTML(w http.ResponseWriter, status int, name string, binding interface{}, htmlOpt ...render.HTMLOptions) {
	Render.HTML(w, status, name, binding, htmlOpt...)
}

// DefaultVars adds the default vars (User and Session) to the data map
// using the global renderer instance
func DefaultVars(req *http.Request, m Vars) map[string]interface{} {
	return Render.DefaultVars(req, m)
}

// DefaultVars adds the default vars (User and Session) to the data map
func (r Renderer) DefaultVars(req *http.Request, m Vars) map[string]interface{} {
	if m == nil {
		log.TraceR(req, "creating template data map", nil)
		m = make(map[string]interface{})
	}

	s, _ := session.Get(req)
	if s == nil {
		log.TraceR(req, "session not found", nil)
		return m
	}

	m["Config"] = r.config.Config()

	log.TraceR(req, "adding session to template data map", nil)
	m["Session"] = s

	if u, ok := s.Values["User"]; ok {
		log.TraceR(req, "adding user to template data map", nil)
		m["User"] = u
	}

	return m
}

// Decoder is a global gorilla schema decoder
var decoder = func() *schema.Decoder {
	var decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	decoder.ZeroEmpty(true) // allows HTML fields to be "cleared"
	return decoder
}()

// Vtom converts validator.v5 errors to a map[string]interface{}
func Vtom(req *http.Request, errs *validator.StructErrors) func(field string) map[string]interface{} {
	log.TraceR(req, "flattening errors", log.Data{"errors": errs})
	ef := errs.Flatten()
	return func(field string) map[string]interface{} {
		if e, ok := ef[field]; ok {
			log.TraceR(req, "got form error", log.Data{"field": field, "tag": e.Tag})
			return map[string]interface{}{e.Tag: e}
		}
		log.TraceR(req, "error not found in map", log.Data{"field": field})
		return nil
	}
}

// Ftos converts form data into a model
func Ftos(req *http.Request, model interface{}) error {
	log.TraceR(req, "parsing request form", nil)
	err := req.ParseForm()

	if err != nil {
		log.ErrorR(req, err, nil)
		return err
	}

	err = decoder.Decode(model, req.PostForm)
	if err != nil {
		log.ErrorR(req, err, nil)
	}

	return err
}

// Form creates a htmlform.Form from a model and http.Request
func Form(req *http.Request, model interface{}, errs *validator.StructErrors) htmlform.Form {
	log.TraceR(req, "creating form from model", log.Data{"model": model, "errors": errs})
	return htmlform.Create(model, Vtom(req, errs), []string{}, []string{}).WithCSRF(nosurf.FormFieldName, nosurf.Token(req))
}

// WithCsrfHandler is a middleware wrapper providing CSRF validation
func WithCsrfHandler(h http.Handler) http.Handler {
	csrfHandler := nosurf.New(h)
	csrfHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rsn := nosurf.Reason(req).Error()
		log.DebugR(req, "failed csrf validation", log.Data{"reason": rsn})
		HTML(w, http.StatusBadRequest, "error", map[string]interface{}{"error": rsn})
	}))
	return csrfHandler
}
