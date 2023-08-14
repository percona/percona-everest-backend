// percona-everest-backend
// Copyright (C) 2023 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package config ...
package config

import "github.com/kelseyhightower/envconfig"

// EverestConfig stores the configuration for the application.
type EverestConfig struct {
	DSN      string `default:"postgres://admin:pwd@127.0.0.1:5432/postgres?sslmode=disable" envconfig:"DSN"`
	HTTPPort int    `default:"8080" envconfig:"HTTP_PORT"`
	Verbose  bool   `default:"false" envconfig:"VERBOSE"`
}

// ParseConfig parses env vars and fills EverestConfig.
func ParseConfig() (*EverestConfig, error) {
	c := &EverestConfig{}
	err := envconfig.Process("", c)
	return c, err
}
