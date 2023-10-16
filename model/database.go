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

// Package model provides structures and methods to operate databases and another storages.
// Migration methods, queries, DB connection initialization and another related stuff should be placed here.
package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // driver for loading migrations files
	_ "github.com/lib/pq"                                // postgresql driver
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Database implements methods for interacting with database.
type Database struct {
	gormDB *gorm.DB
	dir    string
	l      *zap.Logger
}

// OpenDB opens a connection to a postgres database instance.
func OpenDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to create a connection pool to PostgreSQL"))
	}
	return db, nil
}

// NewDatabase returns new Database instance.
func NewDatabase(name, dsn, migrationsDir string) (*Database, error) {
	l := zap.L().Named(fmt.Sprintf("DB.%s", name))

	db, err := OpenDB(dsn)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to connect to database"))
	}

	return &Database{
		gormDB: db,
		dir:    migrationsDir,
		l:      l,
	}, nil
}

// Close closes underlying database connections.
func (db *Database) Close() error {
	gormDB, err := db.gormDB.DB()
	if err != nil {
		return err
	}
	return gormDB.Close()
}

// Begin begins a transaction and returns the object to work with it.
func (db *Database) Begin(ctx context.Context) *gorm.DB {
	return db.gormDB.Begin()
}

// Exec executes the given query on the database.
func (db *Database) Exec(query string) *gorm.DB {
	return db.gormDB.Exec(query)
}

// Transaction start a transaction as a block,
// return error will rollback, otherwise to commit.
func (db *Database) Transaction(fn func(tx *gorm.DB) error) error {
	return db.gormDB.Transaction(fn)
}

// Migrate migrates database schema up and returns actual schema version number.
func (db *Database) Migrate() (uint, error) {
	gormDB, err := db.gormDB.DB()
	if err != nil {
		return 0, err
	}
	pgInstace, err := migratePostgres.WithInstance(gormDB, &migratePostgres.Config{})
	if err != nil {
		return 0, errors.Join(err, errors.New("failed to setup migrator driver"))
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+db.dir, "", pgInstace)
	if err != nil {
		return 0, errors.Join(err, errors.New("failed to setup migrator"))
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return 0, errors.Join(err, errors.New("failed to apply"))
	}

	v, dirty, err := m.Version()
	if err != nil {
		return 0, errors.Join(err, errors.New("failed to check version"))
	}
	if dirty {
		return 0, errors.New("database is dirty; manual fix is required")
	}

	return v, nil
}
