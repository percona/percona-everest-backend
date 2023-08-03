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
