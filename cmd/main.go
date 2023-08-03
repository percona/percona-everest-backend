// Package main is the entry point of the service.
package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/go-logr/zapr"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/percona/percona-everest-backend/api"
	"github.com/percona/percona-everest-backend/cmd/config"
	"github.com/percona/percona-everest-backend/pkg/logger"
	"github.com/percona/percona-everest-backend/public"
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

	// This is how you set up a basic Echo router
	e := echo.New()
	fsys, err := fs.Sub(public.Static, "dist")
	if err != nil {
		l.Fatalf("error reading filesystem\n: %s", err)
	}
	staticFilesHandler := http.FileServer(http.FS(fsys))
	e.GET("/*", echo.WrapHandler(staticFilesHandler))
	// Log all requests
	e.Use(echomiddleware.Logger())
	e.Pre(echomiddleware.RemoveTrailingSlash())

	basePath, err := swagger.Servers.BasePath()
	if err != nil {
		l.Fatalf("Error obtaining base path\n: %s", err)
	}

	// Use our validation middleware to check all requests against the OpenAPI schema.
	g := e.Group(basePath)
	g.Use(middleware.OapiRequestValidator(swagger))
	api.RegisterHandlers(g, server)

	go func() {
		err := e.Start(fmt.Sprintf("0.0.0.0:%d", c.HTTPPort))
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	l.Info("Shutting down http server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		l.Error(errors.Wrap(err, "could not shut down http server"))
	} else {
		l.Info("http server shut down")
	}

	l.Info("Shutting down Everest")
	if err := server.Shutdown(ctx); err != nil {
		l.Error(errors.Wrap(err, "could not shut down Everest"))
	} else {
		l.Info("Everest shut down")
	}

	l.Info("Exiting")
}
