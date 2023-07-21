package config

import "github.com/kelseyhightower/envconfig"

// EverestConfig stores the configuration for the application.
type EverestConfig struct {
	HTTPPort int    `default:"8081" envconfig:"HTTP_PORT"`
	DSN      string `default:"postgres://admin:pwd@127.0.0.1:5432/postgres?sslmode=disable" envconfig:"DSN"`
}

// ParseConfig parses env vars and fills EverestConfig.
func ParseConfig() (*EverestConfig, error) {
	c := &EverestConfig{}
	err := envconfig.Process("", c)
	return c, err
}
