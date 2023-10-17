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

// Package model ...
package model

import (
	"context"
	"errors"
	"time"

	"github.com/lib/pq"
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

// SetSecret creates or updates a secret.
// Shall be reworked with gorm's clauses to update on conflict once we upgrade to 1.20+ version
// because this approach is prone to race-conditions.
func (db *Database) SetSecret(ctx context.Context, id, value string) error {
	err := db.CreateSecret(ctx, id, value)
	var pgErr *pq.Error
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		// duplicate key error
		return db.UpdateSecret(ctx, id, value)
	}

	return err
}
