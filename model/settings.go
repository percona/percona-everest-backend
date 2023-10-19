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
)

// Settings represents db model for Everest settings.
type Settings struct {
	ID string
}

// SettingsParams represents params for Everest settings.
type SettingsParams struct {
	ID string
}

// CreateSettings creates an Everest settings record.
func (db *Database) CreateSettings(_ context.Context, params SettingsParams) (*Settings, error) {
	s := &Settings{
		ID: params.ID,
	}
	err := db.gormDB.Create(s).Error
	if err != nil {
		return nil, err
	}

	return s, nil
}

// GetSettings returns Everest settings.
func (db *Database) GetSettings(_ context.Context) (*Settings, error) {
	var settings []Settings
	err := db.gormDB.First(&settings).Error
	if err != nil {
		return nil, err
	}
	if len(settings) > 1 {
		return nil, errors.New("more than one set of settings found")
	}
	if len(settings) == 0 {
		return nil, errors.New("no settings found")
	}
	return &settings[0], nil
}
