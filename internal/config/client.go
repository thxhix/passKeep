package config

import (
	"github.com/caarlos0/env/v11"
)

// ClientConfig holds the configuration for connecting to the REST server.
//
// The configuration can be loaded from environment variables. Currently, it
// supports only the server address.
type ClientConfig struct {
	ServerAddress string `envDefault:"http://localhost:8080"`
}

// NewClientConfig parses environment variables and returns a ClientConfig.
//
// Returns a pointer to ClientConfig or an error if parsing fails.
func NewClientConfig() (*ClientConfig, error) {
	cfg := &ClientConfig{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
