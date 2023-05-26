// Package storage provides structures and methods to operate databases and another storages.
// Migration methods, queries, DB connection initialization and another related stuff should be placed here.
package storage

import (
	"go.uber.org/zap"
)

// Storage implements methods for interacting with database.
type Storage struct {
	l *zap.Logger //nolint:unused
}
