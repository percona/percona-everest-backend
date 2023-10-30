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
	"fmt"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/percona/percona-everest-backend/pkg/pmm"
)

// CreateMonitoringInstance creates a new monitoring instance.
func (e *EverestServer) CreateMonitoringInstance(ctx echo.Context) error {
	params, err := validateCreateMonitoringInstanceRequest(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	c := ctx.Request().Context()
	m, err := e.kubeClient.GetMonitoringConfig(c, params.Name)
	if err != nil && !k8serrors.IsNotFound(err) {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
		})
	}
	if m != nil {
		err = fmt.Errorf("monitoring config %s already exists", params.Name)
		e.l.Error(err)
		return ctx.JSON(http.StatusConflict, Error{Message: pointer.ToString(err.Error())})
	}
	e.l.Debug("Getting PMM API key by username and password")
	apiKey, err := pmm.CreatePMMApiKey(
		c, params.Url, fmt.Sprintf("everest-%s-%s", params.Name, uuid.NewString()),
		params.Pmm.User, params.Pmm.Password,
	)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not create an API key in PMM"),
		})
	}
	_, err = e.kubeClient.CreateSecret(c, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-secret", params.Name),
			Namespace: e.kubeClient.Namespace(),
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"apiKey": apiKey,
		},
	})
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
		})
	}
	err = e.kubeClient.CreateMonitoringConfig(c, &everestv1alpha1.MonitoringConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: e.kubeClient.Namespace(),
		},
		Spec: everestv1alpha1.MonitoringConfigSpec{
			Type: everestv1alpha1.MonitoringType(params.Type),
			PMM: everestv1alpha1.PMMConfig{
				URL: params.Url,
			},
			CredentialsSecretName: fmt.Sprintf("%s-secret", params.Name),
		},
	})
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
		})
	}
	result := MonitoringInstance{
		Type: MonitoringInstanceBaseWithNameType(params.Type),
		Name: params.Name,
		Url:  params.Url,
	}

	return ctx.JSON(http.StatusOK, result)
}

// ListMonitoringInstances lists all monitoring instances.
func (e *EverestServer) ListMonitoringInstances(ctx echo.Context) error {
	mcList, err := e.kubeClient.ListMonitoringConfigs(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not get a list of monitoring instances")})
	}

	result := make([]*MonitoringInstance, 0, len(mcList.Items))
	for _, mc := range mcList.Items {
		mc := mc
		result = append(result, &MonitoringInstance{
			Type: MonitoringInstanceBaseWithNameType(mc.Spec.Type),
			Name: mc.Name,
			Url:  mc.Spec.PMM.URL,
		})
	}
	return ctx.JSON(http.StatusOK, result)
}

// GetMonitoringInstance retrieves a monitoring instance.
func (e *EverestServer) GetMonitoringInstance(ctx echo.Context, name string) error {
	m, err := e.kubeClient.GetMonitoringConfig(ctx.Request().Context(), name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not get a list of monitoring instances")})
	}

	return ctx.JSON(http.StatusOK, &MonitoringInstance{
		Type: MonitoringInstanceBaseWithNameType(m.Spec.Type),
		Name: m.Name,
		Url:  m.Spec.PMM.URL,
	})
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
	if err := e.kubeClient.DeleteMonitoringConfig(ctx.Request().Context(), name); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get monitoring config"),
		})
	}
	if err := e.kubeClient.DeleteSecret(ctx.Request().Context(), fmt.Sprintf("%s-secret", name)); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
		})
	}
	return ctx.NoContent(http.StatusNoContent)
}
