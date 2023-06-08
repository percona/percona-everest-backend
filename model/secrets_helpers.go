package model

import (
	"time"

	"github.com/labstack/echo/v4"
)

// CreateSecret creates a new Secret record in db.
func (db *Database) CreateSecret(_ echo.Context, id, value string) error {
	return db.gormDB.Create(&Secret{
		ID:        id,
		Value:     value,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}).Error
}

// GetSecret returns the secret by its id.
func (db *Database) GetSecret(_ echo.Context, id string) (string, error) {
	var secret Secret
	err := db.gormDB.First(&secret, "id = ?", id).Error
	if err != nil {
		return "", err
	}
	return secret.Value, nil
}
