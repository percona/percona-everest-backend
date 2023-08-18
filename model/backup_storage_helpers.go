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
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// CreateBackupStorageParams parameters for BackupStorage record creation.
type CreateBackupStorageParams struct {
	Name        string
	Description string
	Type        string
	BucketName  string
	URL         string
	Region      string
	AccessKeyID string
	SecretKeyID string
}

// UpdateBackupStorageParams parameters for BackupStorage record update.
type UpdateBackupStorageParams struct {
	Name        string
	Description *string
	BucketName  *string
	URL         *string
	Region      *string
	AccessKeyID *string
	SecretKeyID *string
}

// BackupStorage represents db model for BackupStorage.
type BackupStorage struct {
	Type        string
	Name        string `gorm:"primaryKey"`
	Description string
	BucketName  string
	URL         string
	Region      string
	AccessKeyID string
	SecretKeyID string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// SecretName returns the name of the k8s secret as referenced by the k8s MonitoringConfig resource.
func (b *BackupStorage) SecretName() string {
	return fmt.Sprintf("%s-secret", b.Name)
}

// Secrets returns all monitoring instance secrets from secrets storage.
func (b *BackupStorage) Secrets(ctx context.Context, getSecret func(ctx context.Context, id string) (string, error)) (map[string]string, error) {
	secretKey, err := getSecret(ctx, b.SecretKeyID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get secretKey")
	}
	accessKey, err := getSecret(ctx, b.AccessKeyID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get accessKey")
	}
	return map[string]string{
		b.SecretKeyID: secretKey,
		b.AccessKeyID: accessKey,
	}, nil
}

// K8sResource returns a resource which shall be created when storing this struct in Kubernetes.
func (b *BackupStorage) K8sResource(namespace string) (runtime.Object, error) { //nolint:unparam,ireturn
	bs := &everestv1alpha1.BackupStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name,
			Namespace: namespace,
		},
		Spec: everestv1alpha1.BackupStorageSpec{
			Type:                  everestv1alpha1.BackupStorageType(b.Type),
			Bucket:                b.BucketName,
			Region:                b.Region,
			EndpointURL:           b.URL,
			CredentialsSecretName: b.SecretName(),
		},
	}

	return bs, nil
}

// CreateBackupStorage creates a BackupStorage record.
func (db *Database) CreateBackupStorage(_ context.Context, params CreateBackupStorageParams) (*BackupStorage, error) {
	s := &BackupStorage{
		Name:        params.Name,
		Description: params.Description,
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

// GetBackupStorage returns BackupStorage record by its Name.
func (db *Database) GetBackupStorage(_ context.Context, name string) (*BackupStorage, error) {
	storage := &BackupStorage{}
	// fixme: for some reason, gorm doesn't understand the Name field as a PrimaryKey,
	// so "Where" is added as a quickfix
	err := db.gormDB.First(storage, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return storage, nil
}

// UpdateBackupStorage updates a BackupStorage record.
func (db *Database) UpdateBackupStorage(_ context.Context, tx *gorm.DB, params UpdateBackupStorageParams) error {
	target := db.gormDB
	if tx != nil {
		target = tx
	}
	old := &BackupStorage{}
	err := target.First(old, "name = ?", params.Name).Error
	if err != nil {
		return err
	}

	record := BackupStorage{}
	if params.Description != nil {
		record.Description = *params.Description
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
	if err = target.Model(old).Where("name = ?", params.Name).Updates(record).Error; err != nil {
		return err
	}

	return nil
}

// DeleteBackupStorage returns BackupStorage record by its Name.
func (db *Database) DeleteBackupStorage(_ context.Context, name string, tx *gorm.DB) error {
	gormDB := db.gormDB
	if tx != nil {
		gormDB = tx
	}
	return gormDB.Delete(&BackupStorage{}, "name = ?", name).Error
}
