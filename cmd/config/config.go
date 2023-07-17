package config

import "github.com/kelseyhightower/envconfig"

// EverestConfig stores the configuration for the application.
type EverestConfig struct {
	DSN string `envconfig:"DSN"`
}

// ParseConfig parses env vars and fills EverestConfig.
func ParseConfig() (*EverestConfig, error) {
	c := &EverestConfig{}
	err := envconfig.Process("", c)
	return c, err
}
