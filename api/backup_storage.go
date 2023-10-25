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

// Package api ...
package api

import (
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
)

// ListBackupStorages lists backup storages.
func (e *EverestServer) ListBackupStorages(ctx echo.Context) error {
	// TODO: Fix it

	return ctx.JSON(http.StatusOK, nil)
}

// CreateBackupStorage creates a new backup storage object.
// Rollbacks are implemented without transactions bc the secrets storage is going to be moved out of pg.
func (e *EverestServer) CreateBackupStorage(ctx echo.Context) error {
	params, err := validateCreateBackupStorageRequest(ctx, e.l)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	c := ctx.Request().Context()
	// TODO: Create backup storage in k8s
	_ = c
	_ = params

	return ctx.JSON(http.StatusOK, nil)
}

// DeleteBackupStorage deletes the specified backup storage.
func (e *EverestServer) DeleteBackupStorage(ctx echo.Context, backupStorageName string) error {
	// TODO: Fix it

	return ctx.NoContent(http.StatusNoContent)
}

// GetBackupStorage retrieves the specified backup storage.
func (e *EverestServer) GetBackupStorage(ctx echo.Context, backupStorageID string) error {
	// TODO: Implement the logic
	return ctx.JSON(http.StatusOK, nil)
}

// UpdateBackupStorage updates of the specified backup storage.
func (e *EverestServer) UpdateBackupStorage(ctx echo.Context, backupStorageName string) error {
	// TODO: FIX IT
	params, err := validateUpdateBackupStorageRequest(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	_ = params
	//err = validateStorageAccessByUpdate(ctx, oldData, params, e.l)
	//if err != nil {
	//	return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	//}

	return nil
}
