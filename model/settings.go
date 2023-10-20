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
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// EverestIDSettingName name of the Everest ID setting.
const everestIDSettingName = "everest_id"

// Setting represents db model for Everest settings.
type Setting struct {
	ID    string
	Key   string
	Value string
}

// SettingParams represents params for Everest settings.
type SettingParams struct {
	Key   string
	Value string
}

// InitSettings creates an Everest settings record.
func (db *Database) InitSettings(ctx context.Context) error {
	everestID, err := db.GetEverestID(ctx)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// settings are already initialized
	if everestID != "" {
		return nil
	}

	setting := &Setting{Key: everestIDSettingName, Value: uuid.NewString()}
	return db.gormDB.Create(setting).Error
}

// GetSettingByKey returns Everest settings.
func (db *Database) GetSettingByKey(_ context.Context, key string) (string, error) {
	setting := &Setting{}

	err := db.gormDB.First(&setting, "key = ?", key).Error
	if err != nil {
		return "", err
	}

	return setting.Value, nil
}

// GetEverestID returns Everest settings.
func (db *Database) GetEverestID(ctx context.Context) (string, error) {
	return db.GetSettingByKey(ctx, everestIDSettingName)
}
