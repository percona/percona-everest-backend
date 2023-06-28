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
	secret := &Secret{
		ID:    id,
		Value: value,
	}
	return db.gormDB.Create(secret).Error
}

// GetSecret returns the secret by its id.
func (db *Database) GetSecret(_ context.Context, id string) (string, error) {
	secret := &Secret{
		ID: id,
	}
	err := db.gormDB.First(secret).Error
	if err != nil {
		return "", err
	}
	return secret.Value, nil
}

// UpdateSecret updates the secret by its id.
func (db *Database) UpdateSecret(_ context.Context, id, value string) error {
	secret := &Secret{
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
	secret := &Secret{
		ID: id,
	}
	oldValue, err := db.GetSecret(c, id)
	if err != nil {
		return "", err
	}

	err = db.gormDB.Delete(secret).Error
	if err != nil {
		return "", err
	}
	return oldValue, nil
}
