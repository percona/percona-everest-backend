package model

import (
	"context"
	"time"
)

// Secret represents a key-value secret. TODO: move secrets out of pg //nolint:godox.
type Secret struct {
	ID    string
	Value string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateSecret creates a new Secret record in db.
func (db *Database) CreateSecret(_ context.Context, id, value string) error {
	return db.gormDB.Create(&Secret{ //nolint:exhaustruct
		ID:    id,
		Value: value,
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
	err := db.gormDB.Save(secret).Error
	if err != nil {
		return err
	}
	return nil
}

// DeleteSecret deletes the secret by its id.
func (db *Database) DeleteSecret(_ context.Context, id string) error {
	secret := Secret{ //nolint:exhaustruct
		ID: id,
	}
	err := db.gormDB.Delete(&secret).Error
	if err != nil {
		return err
	}
	return nil
}
