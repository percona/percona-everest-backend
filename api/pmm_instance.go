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
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/percona/percona-everest-backend/model"
)

// CreatePMMInstance creates a new PMM instance.
func (e *EverestServer) CreatePMMInstance(ctx echo.Context) error {
	params, err := validateCreatePMMInstanceRequest(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	apiKeyID := uuid.NewString()
	if err := e.secretsStorage.CreateSecret(ctx.Request().Context(), apiKeyID, params.ApiKey); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not save API key to secrets storage")})
	}

	pmm, err := e.storage.CreatePMMInstance(&model.PMMInstance{
		URL:            params.Url,
		APIKeySecretID: apiKeyID,
	})
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not save PMM instance")})
	}

	return ctx.JSON(http.StatusOK, e.pmmInstanceToAPIJson(pmm))
}

// ListPMMInstances lists all PMM instances.
func (e *EverestServer) ListPMMInstances(ctx echo.Context) error {
	list, err := e.storage.ListPMMInstances()
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not get a list of PMM instances")})
	}

	result := make([]*PMMInstance, 0, len(list))
	for _, pmm := range list {
		pmm := pmm
		result = append(result, e.pmmInstanceToAPIJson(&pmm))
	}

	return ctx.JSON(http.StatusOK, result)
}

// GetPMMInstance retrieves a PMM instance.
func (e *EverestServer) GetPMMInstance(ctx echo.Context, pmmInstanceID string) error {
	pmm, err := e.storage.GetPMMInstance(pmmInstanceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString("PMM instance not found")})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not find PMM instance")})
	}

	return ctx.JSON(http.StatusOK, e.pmmInstanceToAPIJson(pmm))
}

// UpdatePMMInstance updates a PMM instance based on the provided fields.
func (e *EverestServer) UpdatePMMInstance(ctx echo.Context, pmmInstanceID string) error {
	params, err := validateUpdatePMMInstanceRequest(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	pmm, err := e.storage.GetPMMInstance(pmmInstanceID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString("Could not find PMM instance")})
	}

	var apiKeyID *string
	if params.ApiKey != nil {
		id := uuid.NewString()
		apiKeyID = &id
		if err := e.secretsStorage.CreateSecret(ctx.Request().Context(), *apiKeyID, *params.ApiKey); err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not save API key to secrets storage")})
		}
	}

	err = e.storage.UpdatePMMInstance(pmmInstanceID, model.UpdatePMMInstanceParams{
		URL:            params.Url,
		APIKeySecretID: apiKeyID,
	})
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("Could not update PMM instance")})
	}

	if apiKeyID != nil {
		_, err := e.secretsStorage.DeleteSecret(context.Background(), pmm.APIKeySecretID)
		if err != nil {
			e.l.Error(errors.Wrapf(err, "could not delete PMM instance api key secret %s", pmm.APIKeySecretID))
		}
	}

	pmm, err = e.storage.GetPMMInstance(pmmInstanceID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString("Could not find PMM instance")})
	}

	return ctx.JSON(http.StatusOK, e.pmmInstanceToAPIJson(pmm))
}

// DeletePMMInstance deletes a PMM instance.
func (e *EverestServer) DeletePMMInstance(ctx echo.Context, pmmInstanceID string) error {
	pmm, err := e.storage.GetPMMInstance(pmmInstanceID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString("Could not find PMM instance")})
	}

	if err := e.storage.DeletePMMInstance(pmmInstanceID); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not delete PMM instance")})
	}

	go func() {
		_, err := e.secretsStorage.DeleteSecret(context.Background(), pmm.APIKeySecretID)
		if err != nil {
			e.l.Error(errors.Wrapf(err, "could not delete PMM instance api key secret %s", pmm.APIKeySecretID))
		}
	}()

	return ctx.NoContent(http.StatusNoContent)
}

// pmmInstanceToAPIJson converts PMM instance model to API JSON response.
func (e *EverestServer) pmmInstanceToAPIJson(pmm *model.PMMInstance) *PMMInstance {
	return &PMMInstance{
		Id:             &pmm.ID,
		Url:            pmm.URL,
		ApiKeySecretId: pmm.APIKeySecretID,
	}
}
