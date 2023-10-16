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
	"crypto/rand"
	"errors"

	physPostgreSQL "github.com/hashicorp/vault/physical/postgresql"
	"github.com/hashicorp/vault/sdk/logical"
	vault "github.com/hashicorp/vault/vault"
)

// SecretsStorage implements methods for interacting with secrets.
type SecretsStorage struct {
	barrier *vault.AESGCMBarrier
}

// NewSecretsStorage returns a new SecretsStorage instance.
func NewSecretsStorage(ctx context.Context, dsn string, key []byte) (*SecretsStorage, error) {
	physical, err := physPostgreSQL.NewPostgreSQLBackend(
		map[string]string{
			"connection_url": dsn,
			"table":          "secrets",
			"ha_enabled":     "false",
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	barrier, err := vault.NewAESGCMBarrier(physical)
	if err != nil {
		return nil, err
	}

	// Initialize the barrier
	err = barrier.Initialize(ctx, key, nil, rand.Reader)
	if err != nil && !errors.Is(err, vault.ErrBarrierAlreadyInit) {
		return nil, err
	}

	// Unseal the barrier
	if err := barrier.Unseal(ctx, key); err != nil {
		return nil, err
	}

	return &SecretsStorage{barrier: barrier}, nil
}

// CreateSecret creates a new Secret record in db.
func (secStor *SecretsStorage) CreateSecret(ctx context.Context, id, value string) error {
	// Put an entry in nested/path
	entry := &logical.StorageEntry{
		Key:   id,
		Value: []byte(value),
	}
	return secStor.barrier.Put(ctx, entry)
}

// GetSecret returns the secret by its id.
func (secStor *SecretsStorage) GetSecret(ctx context.Context, id string) (string, error) {
	storedEntry, err := secStor.barrier.Get(ctx, id)
	if err != nil {
		return "", err
	}
	return string(storedEntry.Value), nil
}

// UpdateSecret updates the secret by its id.
func (secStor *SecretsStorage) UpdateSecret(ctx context.Context, id, value string) error {
	return secStor.CreateSecret(ctx, id, value)
}

// DeleteSecret deletes the secret by its id. Returns the deleted secret.
func (secStor *SecretsStorage) DeleteSecret(ctx context.Context, id string) error {
	return secStor.barrier.Delete(ctx, id)
}
