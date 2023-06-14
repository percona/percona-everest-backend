package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CreateBackupStorage creates a BackupStorage record.
func (db *Database) CreateBackupStorage(_ echo.Context, params CreateBackupStorageParams) (*BackupStorage, error) {
	s := &BackupStorage{
		ID:         uuid.NewString(),
		Name:       params.Name,
		BucketName: params.BucketName,
		URL:        params.URL,
		Region:     params.Region,

		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	err := db.gormDB.Create(s).Error
	if err != nil {
		return nil, err
	}

	return s, nil
}

// ListBackupStorages returns all available BackupStorages records.
func (db *Database) ListBackupStorages(_ echo.Context) ([]BackupStorage, error) {
	var storages []BackupStorage
	err := db.gormDB.Find(&storages).Error
	if err != nil {
		return nil, err
	}
	return storages, nil
}

// GetBackupStorage returns BackupStorage record by its ID.
func (db *Database) GetBackupStorage(_ echo.Context, id string) (*BackupStorage, error) {
	var storage BackupStorage
	err := db.gormDB.First(&storage, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &storage, nil
}

// UpdateBackupStorage updates a BackupStorage record.
func (db *Database) UpdateBackupStorage(ctx echo.Context, params UpdateBackupStorageParams) (*BackupStorage, error) {
	record, err := db.GetBackupStorage(ctx, params.ID)
	if err != nil {
		return nil, err
	}

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

	err = db.gormDB.Save(record).Error
	if err != nil {
		return nil, err
	}

	return record, nil
}
