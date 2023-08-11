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

// Package model ..
package model

import (
	"github.com/pkg/errors"
)

// UpdateMonitoringInstanceParams stores fields to be updated in monitoring instance.
type UpdateMonitoringInstanceParams struct {
	Type           *MonitoringInstanceType
	URL            *string
	APIKeySecretID *string
}

// CreateMonitoringInstance creates a new monitoring instance.
func (db *Database) CreateMonitoringInstance(i *MonitoringInstance) (*MonitoringInstance, error) {
	if i == nil {
		return nil, errors.New("i parameter cannot be empty")
	}

	if err := db.gormDB.Create(i).Error; err != nil {
		return nil, err
	}

	return i, nil
}

// ListMonitoringInstances lists all monitoring instances.
func (db *Database) ListMonitoringInstances() ([]MonitoringInstance, error) {
	var i []MonitoringInstance
	if err := db.gormDB.Find(&i).Error; err != nil {
		return nil, err
	}
	return i, nil
}

// GetMonitoringInstance retrieves a monitoring instance.
func (db *Database) GetMonitoringInstance(name string) (*MonitoringInstance, error) {
	i := &MonitoringInstance{Name: name}
	if err := db.gormDB.First(i).Error; err != nil {
		return nil, err
	}
	return i, nil
}

// DeleteMonitoringInstance deletes a monitoring instance.
func (db *Database) DeleteMonitoringInstance(name string) error {
	return db.gormDB.Delete(&MonitoringInstance{Name: name}).Error
}

// UpdateMonitoringInstance updates fields of a monitoring instance based on the provided fields.
func (db *Database) UpdateMonitoringInstance(name string, params UpdateMonitoringInstanceParams) error {
	i := &MonitoringInstance{Name: name}
	if params.Type != nil {
		i.Type = *params.Type
	}
	if params.URL != nil {
		i.URL = *params.URL
	}
	if params.APIKeySecretID != nil {
		i.APIKeySecretID = *params.APIKeySecretID
	}

	return db.gormDB.Model(&MonitoringInstance{}).Updates(i).Error
}
