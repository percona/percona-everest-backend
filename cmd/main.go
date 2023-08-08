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
// Package main is the entry point of the service.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-logr/zapr"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/percona/percona-everest-backend/api"
	"github.com/percona/percona-everest-backend/cmd/config"
	"github.com/percona/percona-everest-backend/pkg/logger"
)

func main() {
	logger := logger.MustInitLogger()
	defer logger.Sync() //nolint:errcheck
	l := logger.Sugar()

	// This is required because controller-runtime requires a logger
	// to be set within 30 seconds of the program initialization.
	log := zapr.NewLogger(logger)
	ctrlruntimelog.SetLogger(log)

	swagger, err := api.GetSwagger()
	if err != nil {
		l.Fatalf("Error loading swagger spec\n: %s", err)
	}

	c, err := config.ParseConfig()
	if err != nil {
		l.Fatalf("Failed parsing config: %+v", err)
	}
	if !c.Verbose {
		logger = logger.WithOptions(zap.IncreaseLevel(zap.InfoLevel))
		l = logger.Sugar()
	}
	l.Debug("Debug logging enabled")

	server, err := api.NewEverestServer(c, l)
	if err != nil {
		l.Fatalf("Error creating Everest Server\n: %s", err)
	}

	e := echo.New()
	if err := server.BootstrapHTTPServer(e, swagger); err != nil {
		l.Fatal(err)
	}

	go func() {
		err := e.Start(fmt.Sprintf("0.0.0.0:%d", c.HTTPPort))
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx, e); err != nil {
		l.Error(errors.Wrap(err, "could not shut down Everest"))
	} else {
		l.Info("Everest shut down")
	}

	l.Info("Exiting")
}
