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
	"time"

	"github.com/google/uuid"
)

// CreateBackupStorageParams parameters for BackupStorage record creation.
type CreateBackupStorageParams struct {
	Name        string
	Type        string
	BucketName  string
	URL         string
	Region      string
	AccessKeyID string
	SecretKeyID string
}

// UpdateBackupStorageParams parameters for BackupStorage record update.
type UpdateBackupStorageParams struct {
	ID          string
	Name        *string
	BucketName  *string
	URL         *string
	Region      *string
	AccessKeyID *string
	SecretKeyID *string
}

// BackupStorage represents db model for BackupStorage.
type BackupStorage struct {
	ID          string
	Type        string
	Name        string
	BucketName  string
	URL         string
	Region      string
	AccessKeyID string
	SecretKeyID string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateBackupStorage creates a BackupStorage record.
func (db *Database) CreateBackupStorage(_ context.Context, params CreateBackupStorageParams) (*BackupStorage, error) {
	s := &BackupStorage{
		ID:          uuid.NewString(),
		Name:        params.Name,
		Type:        params.Type,
		BucketName:  params.BucketName,
		URL:         params.URL,
		Region:      params.Region,
		AccessKeyID: params.AccessKeyID,
		SecretKeyID: params.SecretKeyID,
	}
	err := db.gormDB.Create(s).Error
	if err != nil {
		return nil, err
	}

	return s, nil
}

// ListBackupStorages returns all available BackupStorages records.
func (db *Database) ListBackupStorages(_ context.Context) ([]BackupStorage, error) {
	var storages []BackupStorage
	err := db.gormDB.Find(&storages).Error
	if err != nil {
		return nil, err
	}
	return storages, nil
}

// GetBackupStorage returns BackupStorage record by its ID.
func (db *Database) GetBackupStorage(_ context.Context, id string) (*BackupStorage, error) {
	storage := &BackupStorage{
		ID: id,
	}
	err := db.gormDB.First(storage).Error
	if err != nil {
		return nil, err
	}
	return storage, nil
}

// UpdateBackupStorage updates a BackupStorage record.
func (db *Database) UpdateBackupStorage(_ context.Context, params UpdateBackupStorageParams) (*BackupStorage, error) {
	old := &BackupStorage{
		ID: params.ID,
	}
	err := db.gormDB.First(old).Error
	if err != nil {
		return nil, err
	}

	record := BackupStorage{}
	if params.Name != nil {
		record.Name = *params.Name
	}
	if params.BucketName != nil {
		record.BucketName = *params.BucketName
	}
	if params.URL != nil {
		record.URL = *params.URL
	}
	if params.Region != nil {
		record.Region = *params.Region
	}
	if params.AccessKeyID != nil {
		record.AccessKeyID = *params.AccessKeyID
	}
	if params.SecretKeyID != nil {
		record.SecretKeyID = *params.SecretKeyID
	}

	// Updates only non-empty fields defined in record
	if err = db.gormDB.Model(old).Updates(record).Error; err != nil {
		return nil, err
	}

	return old, nil
}

// DeleteBackupStorage returns BackupStorage record by its ID.
func (db *Database) DeleteBackupStorage(_ context.Context, id string) error {
	storage := &BackupStorage{
		ID: id,
	}
	err := db.gormDB.Delete(storage).Error
	if err != nil {
		return err
	}
	return nil
}
