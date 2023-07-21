// Package main is the entry point of the service.
package main

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/percona/percona-everest-backend/api"
	"github.com/percona/percona-everest-backend/cmd/config"
	"github.com/percona/percona-everest-backend/public"
)

func main() {
	logger, _ := zap.NewDevelopment()
	l := logger.Sugar()

	swagger, err := api.GetSwagger()
	if err != nil {
		l.Fatalf("Error loading swagger spec\n: %s", err)
	}

	c, err := config.ParseConfig()
	if err != nil {
		l.Fatalf("Failed parsing config: %+v", err)
	}

	server, err := api.NewEverestServer(c)
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

	// We now register our petStore above as the handler for the interface
	basePath, err := swagger.Servers.BasePath()
	if err != nil {
		l.Fatalf("Error obtaining base path\n: %s", err)
	}
	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	g := e.Group(basePath)
	g.Use(middleware.OapiRequestValidator(swagger))
	api.RegisterHandlers(g, server)

	// And we serve HTTP until the world ends.
	address := e.Start(fmt.Sprintf("0.0.0.0:%d", c.HTTPPort))
	l.Infof("Everest server is available on %s", address)
}
