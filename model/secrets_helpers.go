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

// DeleteSecret deletes the secret by its id. Returns the deleted secret.
func (db *Database) DeleteSecret(c context.Context, id string) (string, error) {
	secret := Secret{ //nolint:exhaustruct
		ID: id,
	}
	oldValue, err := db.GetSecret(c, id)
	if err != nil {
		return "", err
	}

	err = db.gormDB.Delete(&secret).Error
	if err != nil {
		return "", err
	}
	return oldValue, nil
}

// ReplaceSecret deletes the secret with the oldKey and creates a new secret with the given value and newKey.
// Returns the old secret.
func (db *Database) ReplaceSecret(ctx context.Context, oldKey, newKey, value string) (*string, error) {
	secret := Secret{ //nolint:exhaustruct
		ID: oldKey,
	}
	tx := db.gormDB.Begin()

	oldValue, err := db.GetSecret(ctx, oldKey)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = db.gormDB.Delete(&secret).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	newSecret := Secret{ //nolint:exhaustruct
		ID:    newKey,
		Value: value,
	}
	err = db.gormDB.Create(&newSecret).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return &oldValue, nil
}
