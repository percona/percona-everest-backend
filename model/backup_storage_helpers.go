package model

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CreateBackupStorageParams parameters for BackupStorage record creation.
type CreateBackupStorageParams struct {
	Name        string
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
	s := &BackupStorage{ //nolint:exhaustruct
		ID:          uuid.NewString(),
		Name:        params.Name,
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
	storage := &BackupStorage{ //nolint:exhaustruct
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
	old := &BackupStorage{ //nolint:exhaustruct
		ID: params.ID,
	}
	err := db.gormDB.First(old).Error
	if err != nil {
		return nil, err
	}

	record := BackupStorage{} //nolint:exhaustruct
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
	storage := &BackupStorage{ //nolint:exhaustruct
		ID: id,
	}
	err := db.gormDB.Delete(storage).Error
	if err != nil {
		return err
	}
	return nil
}
