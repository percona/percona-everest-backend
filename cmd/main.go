// Package main is the entry point of the service.
package main

import (
	"flag"
	"fmt"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/percona/percona-everest-backend/api"
	"github.com/percona/percona-everest-backend/model"
)

func main() {
	const httpPort = 8081
	port := flag.Int("port", httpPort, "Port for test HTTP server")
	flag.Parse()

	l := zap.L().Sugar()

	swagger, err := api.GetSwagger()
	if err != nil {
		l.Fatalf("Error loading swagger spec\n: %s", err)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	pgStorageName := "postgres"
	pgDSNF := "postgres://admin:pwd@127.0.0.1:5432/postgres?sslmode=disable"
	pgMigrationsF := "migrations"

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
	// Log all requests
	e.Use(echomiddleware.Logger())
	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	e.Use(middleware.OapiRequestValidator(swagger))

	// We now register our petStore above as the handler for the interface
	api.RegisterHandlers(e, server)

	// And we serve HTTP until the world ends.
	l.Fatal("http server failed", zap.Error(e.Start(fmt.Sprintf("0.0.0.0:%d", *port))))
}
