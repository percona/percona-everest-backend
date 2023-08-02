// Package api contains the API server implementation.
package api

//go:generate ../bin/oapi-codegen --config=server.cfg.yml  ../docs/spec/openapi.yml

import (
	"go.uber.org/zap"

	"github.com/percona/percona-everest-backend/cmd/config"
	"github.com/percona/percona-everest-backend/model"
)

const (
	pgStorageName   = "postgres"
	pgMigrationsDir = "migrations"
)

// EverestServer represents the server struct.
type EverestServer struct {
	config         *config.EverestConfig
	l              *zap.SugaredLogger
	storage        storage
	secretsStorage secretsStorage
	everestK8s     everestK8s
}

// List represents a general object with the list of items.
type List struct {
	Items string `json:"items"`
}

// NewEverestServer creates and configures everest API.
func NewEverestServer(c *config.EverestConfig, l *zap.SugaredLogger) (*EverestServer, error) {
	e := &EverestServer{
		config: c,
		l:      l,
	}
	err := e.initEverest()

	return e, err
}

func (e *EverestServer) initEverest() error {
	db, err := model.NewDatabase(pgStorageName, e.config.DSN, pgMigrationsDir)
	if err != nil {
		return err
	}
	e.storage = db
	e.secretsStorage = db // so far the db implements both interfaces - the regular storage and the secrets storage
	_, err = db.Migrate()
	e.everestK8s = newEverestK8s(e.storage, e.secretsStorage, e.l)
	return err
}
