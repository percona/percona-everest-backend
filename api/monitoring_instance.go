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
	"context"
	"fmt"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/percona/percona-everest-backend/model"
	"github.com/percona/percona-everest-backend/pkg/pmm"
)

// CreateMonitoringInstance creates a new monitoring instance.
func (e *EverestServer) CreateMonitoringInstance(ctx echo.Context) error {
	params, err := validateCreateMonitoringInstanceRequest(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	i, err := e.storage.GetMonitoringInstance(params.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get monitoring instances"),
		})
	}
	if i != nil {
		return ctx.JSON(http.StatusConflict, Error{
			Message: pointer.ToString("Monitoring instance with the same name already exists"),
		})
	}

	apiKey := params.Pmm.ApiKey
	if apiKey == "" {
		e.l.Debug("Getting PMM API key by username and password")
		apiKey, err = pmm.CreatePMMApiKey(
			ctx.Request().Context(), params.Url, fmt.Sprintf("everest-%s-%s", params.Name, uuid.NewString()),
			params.Pmm.User, params.Pmm.Password,
		)
		if err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusBadRequest, Error{
				Message: pointer.ToString("Could not create an API key in PMM"),
			})
		}
	}

	apiKeyID := uuid.NewString()
	if err := e.secretsStorage.CreateSecret(ctx.Request().Context(), apiKeyID, apiKey); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusConflict, Error{
			Message: pointer.ToString("Could not save API key to secrets storage"),
		})
	}

	i, err = e.storage.CreateMonitoringInstance(&model.MonitoringInstance{
		Type:           model.MonitoringInstanceType(params.Type),
		Name:           params.Name,
		URL:            params.Url,
		APIKeySecretID: apiKeyID,
	})
	if err != nil {
		go func() {
			_, err := e.secretsStorage.DeleteSecret(ctx.Request().Context(), apiKeyID)
			if err != nil {
				e.l.Warnf("Could not delete secret %s from secret storage due to error: %s", apiKeyID, err)
			}
		}()

		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not save monitoring instance")})
	}

	return ctx.JSON(http.StatusOK, e.monitoringInstanceToAPIJson(i))
}

// ListMonitoringInstances lists all monitoring instances.
func (e *EverestServer) ListMonitoringInstances(ctx echo.Context) error {
	list, err := e.storage.ListMonitoringInstances()
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not get a list of monitoring instances")})
	}

	result := make([]*MonitoringInstance, 0, len(list))
	for _, i := range list {
		i := i
		result = append(result, e.monitoringInstanceToAPIJson(&i))
	}

	return ctx.JSON(http.StatusOK, result)
}

// GetMonitoringInstance retrieves a monitoring instance.
func (e *EverestServer) GetMonitoringInstance(ctx echo.Context, name string) error {
	i, err := e.storage.GetMonitoringInstance(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString("Monitoring instance not found")})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not find monitoring instance")})
	}

	return ctx.JSON(http.StatusOK, e.monitoringInstanceToAPIJson(i))
}

// UpdateMonitoringInstance updates a monitoring instance based on the provided fields.
func (e *EverestServer) UpdateMonitoringInstance(ctx echo.Context, name string) error {
	params, err := validateUpdateMonitoringInstanceRequest(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	i, err := e.storage.GetMonitoringInstance(name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusNotFound, Error{
			Message: pointer.ToString("Could not find monitoring instance"),
		})
	}

	var apiKeyID *string
	if params.Pmm != nil {
		apiKey := params.Pmm.ApiKey
		if apiKey == "" {
			e.l.Debug("Getting PMM API key by username and password")
			apiKey, err = pmm.CreatePMMApiKey(
				ctx.Request().Context(), params.Url, fmt.Sprintf("everest-%s-%s", i.Name, uuid.NewString()),
				params.Pmm.User, params.Pmm.Password,
			)
			if err != nil {
				e.l.Error(err)
				return ctx.JSON(http.StatusBadRequest, Error{
					Message: pointer.ToString("Could not create an API key in PMM"),
				})
			}
		}

		s := uuid.NewString()
		apiKeyID = &s
		if err := e.secretsStorage.CreateSecret(ctx.Request().Context(), *apiKeyID, apiKey); err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusConflict, Error{
				Message: pointer.ToString("Could not save API key to secrets storage"),
			})
		}
	}

	err = e.storage.UpdateMonitoringInstance(name, model.UpdateMonitoringInstanceParams{
		Type:           (*model.MonitoringInstanceType)(&params.Type),
		URL:            &params.Url,
		APIKeySecretID: apiKeyID,
	})
	if err != nil {
		go func() {
			_, err := e.secretsStorage.DeleteSecret(ctx.Request().Context(), *apiKeyID)
			if err != nil {
				e.l.Warnf("Could not delete secret %s from secret storage due to error: %s", apiKeyID, err)
			}
		}()

		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("Could not update monitoring instance")})
	}

	if apiKeyID != nil {
		instance := i
		go func() {
			_, err := e.secretsStorage.DeleteSecret(context.Background(), instance.APIKeySecretID)
			if err != nil {
				e.l.Warn(errors.Wrapf(err, "could not delete monitoring instance api key secret %s", instance.APIKeySecretID))
			}
		}()
	}

	i, err = e.storage.GetMonitoringInstance(name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString("Could not find monitoring instance")})
	}

	return ctx.JSON(http.StatusOK, e.monitoringInstanceToAPIJson(i))
}

// DeleteMonitoringInstance deletes a monitoring instance.
func (e *EverestServer) DeleteMonitoringInstance(ctx echo.Context, name string) error {
	i, err := e.storage.GetMonitoringInstance(name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get monitoring instance"),
		})
	}
	if i == nil {
		return ctx.JSON(http.StatusNotFound, Error{
			Message: pointer.ToString("Could not find monitoring instance"),
		})
	}

	if err := e.storage.DeleteMonitoringInstance(name); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not delete monitoring instance"),
		})
	}

	go func() {
		_, err := e.secretsStorage.DeleteSecret(context.Background(), i.APIKeySecretID)
		if err != nil {
			e.l.Warn(errors.Wrapf(err, "could not delete monitoring instance API key secret %s", i.APIKeySecretID))
		}
	}()

	return ctx.NoContent(http.StatusNoContent)
}

// monitoringInstanceToAPIJson converts monitoring instance model to API JSON response.
func (e *EverestServer) monitoringInstanceToAPIJson(i *model.MonitoringInstance) *MonitoringInstance {
	return &MonitoringInstance{
		Type:           MonitoringInstanceType(i.Type),
		Name:           i.Name,
		Url:            i.URL,
		ApiKeySecretId: &i.APIKeySecretID,
	}
}
