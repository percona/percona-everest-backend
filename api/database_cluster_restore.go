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
	"github.com/labstack/echo/v4"
)

// ListDatabaseClusterRestores List of the created database cluster restores on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusterRestores(ctx echo.Context, kubernetesID string) error {
	return e.proxyKubernetes(ctx, kubernetesID, "")
}

// CreateDatabaseClusterRestore Create a database cluster restore on the specified kubernetes cluster.
func (e *EverestServer) CreateDatabaseClusterRestore(ctx echo.Context, kubernetesID string) error {
	restore := &DatabaseClusterRestore{}
	if err := e.getBodyFromContext(ctx, restore); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not get DatabaseClusterRestore from the request body"),
		})
	}

	if restore.Spec != nil && restore.Spec.DataSource.BackupSource != nil && restore.Spec.DataSource.BackupSource.BackupStorageName != "" {
		_, kubeClient, code, err := e.initKubeClient(ctx.Request().Context(), kubernetesID)
		if err != nil {
			return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
		}

		bsNames := map[string]struct{}{
			restore.Spec.DataSource.BackupSource.BackupStorageName: {},
		}
		if err := e.createK8SBackupStorages(ctx.Request().Context(), kubeClient, bsNames); err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{
				Message: pointer.ToString("Could not create BackupStorage"),
			})
		}
	}

	return e.proxyKubernetes(ctx, kubernetesID, "")
}

// DeleteDatabaseClusterRestore Delete the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) DeleteDatabaseClusterRestore(ctx echo.Context, kubernetesID string, name string) error {
	_, kubeClient, code, err := e.initKubeClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}

	restore, err := kubeClient.GetDatabaseClusterRestore(ctx.Request().Context(), name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get database cluster restore"),
		})
	}

	proxyErr := e.proxyKubernetes(ctx, kubernetesID, name)
	if proxyErr != nil {
		return proxyErr
	}

	// At this point the proxy already sent a response to the API user.
	// We check if the response was successful to continue with cleanup.
	if ctx.Response().Status >= http.StatusMultipleChoices {
		return nil
	}

	if restore.Spec.DataSource.BackupSource != nil && restore.Spec.DataSource.BackupSource.BackupStorageName != "" {
		bsNames := map[string]struct{}{
			restore.Spec.DataSource.BackupSource.BackupStorageName: {},
		}
		go e.deleteK8SBackupStorages(context.Background(), kubeClient, bsNames)
	}

	return nil
}

// GetDatabaseClusterRestore Returns the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseClusterRestore(ctx echo.Context, kubernetesID string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesID, name)
}

// UpdateDatabaseClusterRestore Replace the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseClusterRestore(ctx echo.Context, kubernetesID string, name string) error {
	_, kubeClient, code, err := e.initKubeClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}

	newRestore := &DatabaseClusterRestore{}
	if err := e.getBodyFromContext(ctx, newRestore); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not get DatabaseClusterRestore from the request body"),
		})
	}

	oldRestore, err := kubeClient.GetDatabaseClusterRestore(ctx.Request().Context(), name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get database cluster restore"),
		})
	}
	proxyErr := e.proxyKubernetes(ctx, kubernetesID, name)
	if proxyErr != nil {
		return proxyErr
	}

	// if there were no changes in BackupStorageName field - do nothing
	if oldRestore.Spec.DataSource.BackupSource == nil || newRestore.Spec.DataSource.BackupSource == nil ||
		oldRestore.Spec.DataSource.BackupSource.BackupStorageName == newRestore.Spec.DataSource.BackupSource.BackupStorageName {
		return nil
	}

	// need to create the new BackupStorages CRs
	toCreateNames := map[string]struct{}{
		newRestore.Spec.DataSource.BackupSource.BackupStorageName: {},
	}
	if err = e.createK8SBackupStorages(context.Background(), kubeClient, toCreateNames); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not create BackupStorage"),
		})
	}

	// need to delete unused BackupStorages CRs
	toDeleteNames := map[string]struct{}{
		oldRestore.Spec.DataSource.BackupSource.BackupStorageName: {},
	}

	go e.deleteK8SBackupStorages(context.Background(), kubeClient, toDeleteNames)
	return nil
}
