// Package model provides structures and methods to operate databases and another storages.
// Migration methods, queries, DB connection initialization and another related stuff should be placed here.
package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // driver for loading migrations files
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq" // postgresql driver
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// CreateKubernetesClusterParams parameters for KubernetesCluster record creation.
type CreateKubernetesClusterParams struct {
	Name string
}

// KubernetesCluster represents db model for KubernetesCluster.
type KubernetesCluster struct {
	ID   string `gorm:"id,primary_key"`
	Name string `gorm:"name"`

	CreatedAt time.Time `gorm:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at"`
}

// Secret represents a key-value secret. TODO: move secrets out of pg //nolint:godox.
type Secret struct {
	ID    string `gorm:"id,pk"`
	Value string `gorm:"value"`

	CreatedAt time.Time `gorm:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at"`
}

// CreateBackupStorageParams parameters for BackupStorage record creation.
type CreateBackupStorageParams struct {
	Name       string
	BucketName string
	URL        string
	Region     string
}

// UpdateBackupStorageParams parameters for BackupStorage record update.
type UpdateBackupStorageParams struct {
	ID         string
	Name       *string
	BucketName *string
	URL        *string
	Region     *string
}

// BackupStorage represents db model for BackupStorage.
type BackupStorage struct {
	ID         string `gorm:"id,primary_key"`
	Name       string `gorm:"name"`
	BucketName string `gorm:"bucket_name"`
	URL        string `gorm:"url"`
	Region     string `gorm:"region"`

	CreatedAt time.Time `gorm:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at"`
}

// Database implements methods for interacting with database.
type Database struct {
	gormDB *gorm.DB
	dir    string
	l      *zap.Logger
}

// OpenDB opens a connection to a postgres database instance.
func OpenDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a connection pool to PostgreSQL")
	}
	return db, nil
}

// NewDatabase returns new Database instance.
func NewDatabase(name, dsn, migrationsDir string) (*Database, error) {
	l := zap.L().Named(fmt.Sprintf("DB.%s", name))

	db, err := OpenDB(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	return &Database{
		gormDB: db,
		dir:    migrationsDir,
		l:      l,
	}, nil
}

// Close closes underlying database connections.
func (db *Database) Close() error {
	return db.gormDB.Close()
}

// Exec executes the given query on the database.
func (db *Database) Exec(query string) (sql.Result, error) {
	return db.gormDB.DB().Exec(query)
}

// Migrate migrates database schema up and returns actual schema version number.
func (db *Database) Migrate() (uint, error) {
	pgInstace, err := postgres.WithInstance(db.gormDB.DB(), &postgres.Config{}) //nolint:exhaustruct
	if err != nil {
		return 0, errors.Wrap(err, "failed to setup migrator driver")
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+db.dir, "", pgInstace)
	if err != nil {
		return 0, errors.Wrap(err, "failed to setup migrator")
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return 0, errors.Wrap(err, "failed to apply")
	}

	v, dirty, err := m.Version()
	if err != nil {
		return 0, errors.Wrap(err, "failed to check version")
	}
	if dirty {
		return 0, errors.New("database is dirty; manual fix is required")
	}

	return v, nil
}
