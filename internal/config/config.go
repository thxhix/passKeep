package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
)

// Config holds the application configuration loaded from environment variables
// and optionally overridden by command-line flags.
//
// It includes REST server settings, database connection info, JWT configuration,
// and encryption settings.
type Config struct {
	// RESTAddress is the address where the REST server will listen.
	// Loaded from the environment variable REST_ADDRESS (default: "localhost:8080").
	RESTAddress string `env:"REST_ADDRESS" envDefault:"localhost:8080"`

	// PostgresQL is the PostgreSQL connection URI.
	// Loaded from DATABASE_URI environment variable.
	PostgresQL string `env:"DATABASE_URI"`

	// DatabaseInitTimeoutSeconds specifies the timeout for initializing database connections.
	DatabaseInitTimeoutSeconds int `envDefault:"15"`

	// MigrationsPath is the path to SQL migration files.
	MigrationsPath string `envDefault:"file://migrations"`

	// Embedded JWT configuration.
	JWTConfig

	// Embedded cryptography configuration.
	CryptConfig
}

// NewConfig parses environment variables and command-line flags to create a Config.
//
// Returns a pointer to Config or an error if parsing fails.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	cfg.parseFlags()

	return cfg, nil
}

// parseFlags overrides configuration values with command-line flags if provided.
//
// - "-a" overrides RESTAddress
// - "-d" overrides PostgresQL
func (c *Config) parseFlags() {
	restAddress := flag.String("a", c.RESTAddress, "Адрес запуска REST сервера")
	postgres := flag.String("d", c.PostgresQL, "Креды подключения к PostgreSQL")

	flag.Parse()

	c.RESTAddress = *restAddress
	c.PostgresQL = *postgres
}
