package api

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ListBackupStorages List of the created backup storages.
func (e *EverestServer) ListBackupStorages(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, nil)
}

// CreateBackupStorage Create a new backup storage object.
func (e *EverestServer) CreateBackupStorage(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, nil)
}

// DeleteBackupStorage Delete the specified backup storage.
func (e *EverestServer) DeleteBackupStorage(ctx echo.Context, backupStorageID string) error {
	log.Println(backupStorageID)
	return ctx.JSON(http.StatusNotImplemented, nil)
}

// GetBackupStorage Get the specified backup storage.
func (e *EverestServer) GetBackupStorage(ctx echo.Context, backupStorageID string) error {
	log.Println(backupStorageID)
	return ctx.JSON(http.StatusNotImplemented, nil)
}

// UpdateBackupStorage update of the specified backup storage.
func (e *EverestServer) UpdateBackupStorage(ctx echo.Context, backupStorageID string) error {
	log.Println(backupStorageID)
	return ctx.JSON(http.StatusNotImplemented, nil)
}
