package config

import "github.com/kelseyhightower/envconfig"

type EverestConfig struct {
	DSN string `envconfig:"DSN"`
}

func ParseConfig() (*EverestConfig, error) {
	c := &EverestConfig{}
	err := envconfig.Process("", c)
	return c, err
}
