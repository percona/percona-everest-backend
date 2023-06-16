package model

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CreateBackupStorageParams parameters for BackupStorage record creation.
type CreateBackupStorageParams struct {
	Name       string
	BucketName string
	URL        string
	Region     string
}

// UpdateBackupStorageParams parameters for BackupStorage record update.
type UpdateBackupStorageParams struct {
	ID         string
	Name       *string
	BucketName *string
	URL        *string
	Region     *string
}

// BackupStorage represents db model for BackupStorage.
type BackupStorage struct {
	ID         string
	Name       string
	BucketName string
	URL        string
	Region     string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateBackupStorage creates a BackupStorage record.
func (db *Database) CreateBackupStorage(_ context.Context, params CreateBackupStorageParams) (*BackupStorage, error) {
	s := &BackupStorage{ //nolint:exhaustruct
		ID:         uuid.NewString(),
		Name:       params.Name,
		BucketName: params.BucketName,
		URL:        params.URL,
		Region:     params.Region,
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
	var storage BackupStorage
	err := db.gormDB.First(&storage, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &storage, nil
}

// UpdateBackupStorage updates a BackupStorage record.
func (db *Database) UpdateBackupStorage(ctx context.Context, params UpdateBackupStorageParams) (*BackupStorage, error) {
	old, err := db.GetBackupStorage(ctx, params.ID)
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

	// Updates only non-empty fields defined in record
	if err = db.gormDB.Model(&old).Updates(record).Error; err != nil {
		return nil, err
	}

	return old, nil
}

// DeleteBackupStorage returns BackupStorage record by its ID.
func (db *Database) DeleteBackupStorage(_ context.Context, id string) error {
	var storage BackupStorage
	err := db.gormDB.Delete(&storage, "id = ?", id).Error
	if err != nil {
		return err
	}
	return nil
}
