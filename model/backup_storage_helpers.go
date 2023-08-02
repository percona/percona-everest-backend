package model

import (
	"context"
	"time"
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
	Name        string
	Description string
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
	err := db.gormDB.Where("name = ?", name).First(storage).Error
	if err != nil {
		return nil, err
	}
	return storage, nil
}

// UpdateBackupStorage updates a BackupStorage record.
func (db *Database) UpdateBackupStorage(_ context.Context, params UpdateBackupStorageParams) (*BackupStorage, error) {
	old := &BackupStorage{
		Name: params.Name,
	}
	err := db.gormDB.First(old).Error
	if err != nil {
		return nil, err
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
	if err = db.gormDB.Model(old).Updates(record).Error; err != nil {
		return nil, err
	}

	return old, nil
}

// DeleteBackupStorage returns BackupStorage record by its Name.
func (db *Database) DeleteBackupStorage(_ context.Context, name string) error {
	storage := &BackupStorage{
		Name: name,
	}
	err := db.gormDB.Delete(storage).Error
	if err != nil {
		return err
	}
	return nil
}
