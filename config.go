package service

// Config represents the configuration required for a service
type Config interface {
	Namespace() string
}

// HTTPConfig represents the configuration required for a HTTP service
type HTTPConfig interface {
	Config
	BindAddr() string
	CertFile() string
	KeyFile() string
}

// APIConfig represents the configuration required for an API service
type APIConfig interface {
	HTTPConfig
}

// WebConfig represents the configuration required for a web service
type WebConfig interface {
	HTTPConfig
}

type defaultHTTPConfig struct {
	BindAddr string `env:"BIND_ADDR" flag:"bind-addr" flagDesc:"Bind address"`
	CertFile string `env:"CERT_FILE" flag:"cert-file" flagDesc:"Certificate file"`
	KeyFile  string `env:"KEY_FILE" flag:"key-file" flagDesc:"Key file"`
}

// DefaultAPIConfig is a default APIConfig implementation
type DefaultAPIConfig struct{ defaultHTTPConfig }

// BindAddr implements HTTPConfig.BindAddr
func (c DefaultAPIConfig) BindAddr() string { return c.defaultHTTPConfig.BindAddr }

// CertFile implements HTTPConfig.CertFile
func (c DefaultAPIConfig) CertFile() string { return c.defaultHTTPConfig.CertFile }

// KeyFile implements HTTPConfig.KeyFile
func (c DefaultAPIConfig) KeyFile() string { return c.defaultHTTPConfig.KeyFile }

// DefaultWebConfig is a default WebConfig implementation
type DefaultWebConfig struct {
	defaultHTTPConfig
	SessionSecret string `env:"SESSION_SECRET" flag:"session-secret" flagDesc:"Session secret"`
	SessionName   string `env:"SESSION_NAME" flag:"session-name" flagDesc:"Session name"`
}

// BindAddr implements HTTPConfig.BindAddr
func (c DefaultWebConfig) BindAddr() string { return c.defaultHTTPConfig.BindAddr }

// CertFile implements HTTPConfig.CertFile
func (c DefaultWebConfig) CertFile() string { return c.defaultHTTPConfig.CertFile }

// KeyFile implements HTTPConfig.KeyFile
func (c DefaultWebConfig) KeyFile() string { return c.defaultHTTPConfig.KeyFile }
