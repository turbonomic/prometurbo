package client

import (
	"net/url"
)

const (
	defaultAPIPath = "vmturbo/rest/"
)

type Config struct {
	// ServerAddress is A URL, including <Scheme>://<IP>:<Port>.
	serverAddress *url.URL
	// APIPath is a sub-path that points to an API root.
	apiPath       string

	// For Basic authentication.
	basicAuth     *BasicAuthentication
}

type ConfigBuilder struct {
	serverAddress *url.URL
	apiPath       string
	basicAuth     *BasicAuthentication
}

func NewConfigBuilder(serverAddress *url.URL) *ConfigBuilder {
	return &ConfigBuilder{
		serverAddress: serverAddress,
	}
}

func (cb *ConfigBuilder) APIPath(apiPath string) *ConfigBuilder {
	cb.apiPath = apiPath
	return cb
}

func (cb *ConfigBuilder) BasicAuthentication(usrn, passd string) *ConfigBuilder {
	cb.basicAuth = &BasicAuthentication{
		username: usrn,
		password: passd,
	}
	return cb
}

func (cb *ConfigBuilder) Create() *Config {
	return &Config{
		serverAddress: cb.serverAddress,
		// If API path not specified, use the default API path.
		apiPath: func() string {
			if cb.apiPath == "" {
				return defaultAPIPath
			}
			return cb.apiPath
		}(),
		basicAuth: cb.basicAuth,
	}
}
