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

// Package logger provides functionality related to logging.
package logger

import (
	"go.uber.org/zap"

	"github.com/percona/percona-everest-backend/cmd/config"
)

// MustInitLogger initializes logger and panics in case of an error.
func MustInitLogger() *zap.Logger {
	var (
		logger *zap.Logger
		err    error
	)

	if config.Debug {
		loggerCfg := zap.NewDevelopmentConfig()
		loggerCfg.DisableCaller = true
		logger, err = loggerCfg.Build()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		panic("cannot initialize logger")
	}

	return logger
}
