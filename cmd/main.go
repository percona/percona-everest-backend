// Package main is the entry point of the service.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/percona/percona-everest-backend/api"
	"github.com/percona/percona-everest-backend/cmd/config"
	"github.com/percona/percona-everest-backend/model"
	"github.com/percona/percona-everest-backend/public"
)

func main() { //nolint:funlen
	const httpPort = 8081
	port := flag.Int("port", httpPort, "Port for test HTTP server")
	flag.Parse()

	l := zap.L().Sugar()

	swagger, err := api.GetSwagger()
	if err != nil {
		l.Fatalf("Error loading swagger spec\n: %s", err)
	}

	pgStorageName := "postgres"
	pgDSNF := "postgres://admin:pwd@127.0.0.1:5432/postgres?sslmode=disable"
	pgMigrationsF := "migrations"

	c, err := config.ParseConfig()
	if err != nil {
		l.Fatalf("Failed parsing config: %+v", err)
	}
	if c.DSN != "" {
		pgDSNF = c.DSN
	}
	db, err := model.NewDatabase(pgStorageName, pgDSNF, pgMigrationsF)
	if err != nil {
		l.Fatalf("Failed to init storage: %+v", err)
	}
	defer func() {
		err = db.Close()
		if err != nil {
			l.Error("can't close db connection", zap.Error(err))
		}
	}()

	if _, err = db.Migrate(); err != nil {
		l.Fatalf("Failed to migrate database: %+v", err)
	}

	server := &api.EverestServer{
		Storage:        db,
		SecretsStorage: db, // so far the db implements both interfaces - the regular storage and the secrets storage
	}
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
	g := e.Group(fmt.Sprintf("%s/*", basePath), middleware.OapiRequestValidator(swagger))
	api.RegisterHandlersWithBaseURL(g, server, basePath)

	// And we serve HTTP until the world ends.
	address := e.Start(fmt.Sprintf("0.0.0.0:%d", *port))
	l.Infof("Everest server is available on %s", address)
}
