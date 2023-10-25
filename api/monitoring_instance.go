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

// CreateMonitoringInstance creates a new monitoring instance.
func (e *EverestServer) CreateMonitoringInstance(ctx echo.Context) error {
	params, err := validateCreateMonitoringInstanceRequest(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	_ = params

	// TODO: Fix it
	return ctx.JSON(http.StatusOK, nil)
}

// ListMonitoringInstances lists all monitoring instances.
func (e *EverestServer) ListMonitoringInstances(ctx echo.Context) error {
	// TODO: Fix it
	return ctx.JSON(http.StatusOK, nil)
}

// GetMonitoringInstance retrieves a monitoring instance.
func (e *EverestServer) GetMonitoringInstance(ctx echo.Context, name string) error {
	// TODO: Fix it
	return ctx.JSON(http.StatusOK, nil)
}

// UpdateMonitoringInstance updates a monitoring instance based on the provided fields.
func (e *EverestServer) UpdateMonitoringInstance(ctx echo.Context, name string) error {
	// TODO: Fix it
	params, err := validateUpdateMonitoringInstanceRequest(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	_ = params
	return nil
}

// DeleteMonitoringInstance deletes a monitoring instance.
func (e *EverestServer) DeleteMonitoringInstance(ctx echo.Context, name string) error {
	// TODO: Fix it
	return ctx.NoContent(http.StatusNoContent)
}
