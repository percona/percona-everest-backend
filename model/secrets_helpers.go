package model

import (
	"context"
	"time"
)

// CreateSecret creates a new Secret record in db.
func (db *Database) CreateSecret(_ context.Context, id, value string) error {
	return db.gormDB.Create(&Secret{
		ID:        id,
		Value:     value,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}).Error
}

// GetSecret returns the secret by its id.
func (db *Database) GetSecret(_ context.Context, id string) (string, error) {
	var secret Secret
	err := db.gormDB.First(&secret, "id = ?", id).Error
	if err != nil {
		return "", err
	}
	return secret.Value, nil
}

// UpdateSecret updates the secret by its id.
func (db *Database) UpdateSecret(_ context.Context, id, value string) error {
	secret := &Secret{ //nolint:exhaustruct
		ID:    id,
		Value: value,
	}
	err := db.gormDB.Update(&secret).Error
	if err != nil {
		return err
	}
	return nil
}
