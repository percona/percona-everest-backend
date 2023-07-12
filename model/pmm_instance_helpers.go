package model

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// UpdatePMMInstanceParams stores fields to be updated in PMM instance.
type UpdatePMMInstanceParams struct {
	URL            *string
	APIKeySecretID *string
}

// CreatePMMInstance creates a new PMM instance.
func (db *Database) CreatePMMInstance(pmm *PMMInstance) (*PMMInstance, error) {
	if pmm == nil {
		return nil, errors.New("pmm parameter cannot be empty")
	}

	if pmm.ID == "" {
		pmm.ID = uuid.NewString()
	}

	if err := db.gormDB.Create(pmm).Error; err != nil {
		return nil, err
	}

	return pmm, nil
}

// ListPMMInstances lists all PMM instances.
func (db *Database) ListPMMInstances() ([]PMMInstance, error) {
	var pmm []PMMInstance
	if err := db.gormDB.Find(&pmm).Error; err != nil {
		return nil, err
	}
	return pmm, nil
}

// GetPMMInstance retrieves a PMM instance.
func (db *Database) GetPMMInstance(id string) (*PMMInstance, error) {
	pmm := &PMMInstance{ID: id}
	if err := db.gormDB.First(pmm).Error; err != nil {
		return nil, err
	}
	return pmm, nil
}

// DeletePMMInstance deletes a PMM instance.
func (db *Database) DeletePMMInstance(id string) error {
	return db.gormDB.Delete(&PMMInstance{ID: id}).Error
}

// UpdatePMMInstance updates fields of a PMM instance based on the provided fields.
func (db *Database) UpdatePMMInstance(id string, params UpdatePMMInstanceParams) error {
	pmm := &PMMInstance{ID: id}
	if params.URL != nil {
		pmm.URL = *params.URL
	}
	if params.APIKeySecretID != nil {
		pmm.APIKeySecretID = *params.APIKeySecretID
	}

	return db.gormDB.Model(&PMMInstance{}).Updates(pmm).Error
}
