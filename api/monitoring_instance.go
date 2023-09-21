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
	"errors"
	"fmt"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"

	"github.com/percona/percona-everest-backend/model"
	"github.com/percona/percona-everest-backend/pkg/kubernetes"
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

	apiKeyID, err := e.createAndStorePMMApiKey(
		ctx.Request().Context(), params.Name,
		params.Url, params.Pmm.ApiKey, params.Pmm.User, params.Pmm.Password,
	)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString(err.Error()),
		})
	}

	i, err = e.storage.CreateMonitoringInstance(&model.MonitoringInstance{
		Type:           model.MonitoringInstanceType(params.Type),
		Name:           params.Name,
		URL:            params.Url,
		APIKeySecretID: apiKeyID,
	})
	if err != nil {
		e.l.Error(err)

		_, err := e.secretsStorage.DeleteSecret(ctx.Request().Context(), apiKeyID)
		if err != nil {
			e.l.Warnf("Could not delete secret %s from secret storage due to error: %s", apiKeyID, err)
		}

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
		keyID, err := e.createAndStorePMMApiKey(
			ctx.Request().Context(), i.Name,
			params.Url, params.Pmm.ApiKey, params.Pmm.User, params.Pmm.Password,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, Error{
				Message: pointer.ToString(err.Error()),
			})
		}

		apiKeyID = &keyID
	}

	return e.performMonitoringInstanceUpdate(ctx, name, apiKeyID, i.APIKeySecretID, params)
}

// DeleteMonitoringInstance deletes a monitoring instance.
func (e *EverestServer) DeleteMonitoringInstance(ctx echo.Context, name string) error { //nolint:cyclop
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

	ks, err := e.storage.ListKubernetesClusters(ctx.Request().Context())
	if err != nil {
		return errors.Join(err, errors.New("could not list Kubernetes clusters"))
	}
	if len(ks) == 0 {
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("No registered kubernetes clusters available")})
	}
	// FIXME: Revisit it once multi k8s support will be enabled
	_, kubeClient, _, err := e.initKubeClient(ctx.Request().Context(), ks[0].ID)
	if err != nil {
		e.l.Error(errors.Join(err, errors.New("could not init kube client")))
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not make connection to the kubernetes cluster")})
	}

	err = kubeClient.DeleteConfig(ctx.Request().Context(), i, func(ctx context.Context, name string) (bool, error) {
		return kubernetes.IsMonitoringConfigInUse(ctx, name, kubeClient)
	})
	if err != nil && !errors.Is(err, kubernetes.ErrConfigInUse) {
		e.l.Error(errors.Join(err, errors.New("could not delete monitoring config from kubernetes cluster")))
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not delete monitoring config from the Kubernetes cluster")})
	}

	err = e.storage.Transaction(func(tx *gorm.DB) error {
		if err := e.storage.DeleteMonitoringInstance(i.Name, tx); err != nil {
			e.l.Error(err)
			return errors.New("could not delete monitoring instance")
		}

		_, err = e.secretsStorage.DeleteSecret(context.Background(), i.APIKeySecretID)
		if err != nil {
			return errors.Join(err, fmt.Errorf("could not delete monitoring instance API key secret %s", i.APIKeySecretID))
		}

		return nil
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString(err.Error()),
		})
	}

	return ctx.NoContent(http.StatusNoContent)
}

// monitoringInstanceToAPIJson converts monitoring instance model to API JSON response.
func (e *EverestServer) monitoringInstanceToAPIJson(i *model.MonitoringInstance) *MonitoringInstance {
	return &MonitoringInstance{
		Type: MonitoringInstanceBaseWithNameType(i.Type),
		Name: i.Name,
		Url:  i.URL,
	}
}

func (e *EverestServer) createAndStorePMMApiKey(ctx context.Context, name, url, apiKey, user, password string) (string, error) {
	var err error
	if apiKey == "" {
		e.l.Debug("Getting PMM API key by username and password")
		apiKey, err = pmm.CreatePMMApiKey(
			ctx, url, fmt.Sprintf("everest-%s-%s", name, uuid.NewString()),
			user, password,
		)
		if err != nil {
			e.l.Error(err)
			return "", errors.New("could not create an API key in PMM")
		}
	}

	apiKeyID := uuid.NewString()
	if err := e.secretsStorage.CreateSecret(ctx, apiKeyID, apiKey); err != nil {
		e.l.Error(err)
		return "", errors.New("could not save API key to secrets storage")
	}

	return apiKeyID, nil
}

func (e *EverestServer) performMonitoringInstanceUpdate( //nolint:cyclop
	ctx echo.Context, name string, apiKeyID *string, previousAPIKeyID string,
	params *UpdateMonitoringInstanceJSONRequestBody,
) error {
	var monitoringInstance *model.MonitoringInstance
	err := e.storage.Transaction(func(tx *gorm.DB) error {
		ks, err := e.storage.ListKubernetesClusters(ctx.Request().Context())
		if err != nil {
			return errors.Join(err, errors.New("could not list Kubernetes clusters"))
		}
		if len(ks) == 0 {
			return errors.New("no registered Kubernetes clusters available")
		}
		err = e.storage.UpdateMonitoringInstance(name, model.UpdateMonitoringInstanceParams{
			Type:           (*model.MonitoringInstanceType)(&params.Type),
			URL:            &params.Url,
			APIKeySecretID: apiKeyID,
		})
		if err != nil {
			if _, err := e.secretsStorage.DeleteSecret(ctx.Request().Context(), *apiKeyID); err != nil {
				return errors.Join(err, fmt.Errorf("could not delete secret %s from secret storage", *apiKeyID))
			}

			e.l.Error(err)
			return errors.New("could not update monitoring instance")
		}

		monitoringInstance, err = e.storage.GetMonitoringInstance(name)
		if err != nil {
			e.l.Error(err)
			return errors.New("could not find updated monitoring instance")
		}
		// FIXME: Revisit it once multi k8s support will be enabled
		// FIXME: This is not recommended to do network calls in a database transaction
		// This will be removed during the implementation of multi k8s support
		// However, right now it guarantees data consistency
		_, kubeClient, _, err := e.initKubeClient(ctx.Request().Context(), ks[0].ID)
		if err != nil {
			return errors.Join(err, errors.New("could not init kube client to update config"))
		}

		if err := kubeClient.UpdateConfig(ctx.Request().Context(), monitoringInstance, e.secretsStorage.GetSecret); err != nil {
			return errors.Join(err, errors.New("could not update config"))
		}

		if apiKeyID != nil {
			if _, err := e.secretsStorage.DeleteSecret(context.Background(), previousAPIKeyID); err != nil {
				return errors.Join(err, fmt.Errorf("could not delete monitoring instance api key secret %s", previousAPIKeyID))
			}
		}

		return nil
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString(err.Error()),
		})
	}

	return ctx.JSON(http.StatusOK, e.monitoringInstanceToAPIJson(monitoringInstance))
}
