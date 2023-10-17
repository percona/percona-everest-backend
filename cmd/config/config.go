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

import (
	"os"

	"github.com/kelseyhightower/envconfig"
)

//nolint:gochecknoglobals
var (
	// TelemetryURL Everest telemetry endpoint. The variable is set for the release builds via ldflags
	// to have the correct default telemetry url.
	TelemetryURL string
	// TelemetryInterval Everest telemetry sending frequency. The variable is set for the release builds via ldflags
	// to have the correct default telemetry interval.
	TelemetryInterval string
)

// EverestConfig stores the configuration for the application.
type EverestConfig struct {
	DSN      string `default:"postgres://admin:pwd@127.0.0.1:5432/postgres?sslmode=disable" envconfig:"DSN"`
	HTTPPort int    `default:"8080" envconfig:"HTTP_PORT"`
	Verbose  bool   `default:"false" envconfig:"VERBOSE"`

	// URL stores the url on which Everest is available.
	URL string `default:"http://localhost:8080" required:"true" envconfig:"URL"`

	Auth struct {
		Hostname string `default:"localhost:8081" required:"true" envconfig:"AUTH_HOSTNAME"`
		Insecure bool   `envconfig:"AUTH_INSECURE_CONNECTION"`
		Issuer   string `default:"http://localhost:8081" required:"true" envconfig:"AUTH_ISSUER"`
		// KeyPath stores a path to a .json key file for a service account
		// used to perform the initial Zitadel configuration.
		KeyPath string `required:"true" envconfig:"AUTH_KEY_PATH"`
	}

	// TelemetryURL Everest telemetry endpoint.
	TelemetryURL string `envconfig:"TELEMETRY_URL"`
	// TelemetryInterval Everest telemetry sending frequency.
	TelemetryInterval string `envconfig:"TELEMETRY_INTERVAL"`
}

// ParseConfig parses env vars and fills EverestConfig.
func ParseConfig() (*EverestConfig, error) {
	c := &EverestConfig{}
	if err := envconfig.Process("", c); err != nil {
		return nil, err
	}

	if c.TelemetryURL == "" {
		// checking opt-out - if the env variable does not even exist, set the default URL
		if _, ok := os.LookupEnv("TELEMETRY_URL"); !ok {
			c.TelemetryURL = TelemetryURL
		}
	}
	if c.TelemetryInterval == "" {
		c.TelemetryInterval = TelemetryInterval
	}

	return c, nil
}
