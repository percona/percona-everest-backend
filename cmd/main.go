// Package main is the entry point of the service.
package main

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/go-logr/zapr"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
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
	g.Use(middleware.OapiRequestValidatorWithOptions(swagger, &middleware.Options{
		SilenceServersWarning: false, // This is false on purpose due to a bug in oapi-codegen implementation
	}))
	api.RegisterHandlers(g, server)

	err = e.Start(fmt.Sprintf("0.0.0.0:%d", c.HTTPPort))
	if err != nil {
		l.Fatal(err)
	}

	l.Info("Shutting down")
}
