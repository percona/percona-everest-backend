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
