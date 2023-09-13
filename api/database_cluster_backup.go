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
	"net/url"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
)

// ListDatabaseClusterBackups returns list of the created database cluster backups on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusterBackups(ctx echo.Context, kubernetesID string, name string) error {
	req := ctx.Request()
	if err := validateRFC1123(name, "name"); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("Cluster name is not RFC 1123 compatible")})
	}
	val := url.Values{}
	val.Add("labelSelector", fmt.Sprintf("clusterName=%s", name))
	req.URL.RawQuery = val.Encode()
	path := req.URL.Path
	// trim backups
	path = strings.TrimSuffix(path, "/backups")
	// trim name
	path = strings.TrimSuffix(path, name)
	path = strings.ReplaceAll(path, "database-clusters", "database-cluster-backups")
	req.URL.Path = path
	return e.proxyKubernetes(ctx, kubernetesID, "")
}

// CreateDatabaseClusterBackup creates a database cluster backup on the specified kubernetes cluster.
func (e *EverestServer) CreateDatabaseClusterBackup(ctx echo.Context, kubernetesID string) error {
	backup := &DatabaseClusterBackup{}
	if err := e.getBodyFromContext(ctx, backup); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not get DatabaseCluster from the request body"),
		})
	}

	if backup.Spec == nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("'Spec' field should not be empty")})
	}
	if backup.Spec.BackupStorageName != "" {
		_, kubeClient, code, err := e.initKubeClient(ctx.Request().Context(), kubernetesID)
		if err != nil {
			return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
		}

		bsNames := map[string]struct{}{
			backup.Spec.BackupStorageName: {},
		}
		if err := e.createK8SBackupStorages(ctx.Request().Context(), kubeClient, bsNames); err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{
				Message: pointer.ToString("Could not create BackupStorage"),
			})
		}
	}

	if err := e.validateDBClusterAccess(ctx, kubernetesID, backup.Spec.DbClusterName); err != nil {
		return err
	}

	return e.proxyKubernetes(ctx, kubernetesID, "")
}

// DeleteDatabaseClusterBackup deletes the specified cluster backup on the specified kubernetes cluster.
func (e *EverestServer) DeleteDatabaseClusterBackup(ctx echo.Context, kubernetesID string, name string) error {
	_, kubeClient, code, err := e.initKubeClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}

	backup, err := kubeClient.GetDatabaseClusterBackup(ctx.Request().Context(), name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get database cluster backup"),
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

	if backup.Spec.BackupStorageName != "" {
		bsNames := map[string]struct{}{
			backup.Spec.BackupStorageName: {},
		}
		go e.deleteK8SBackupStorages(context.Background(), kubeClient, bsNames)
	}

	return nil
}

// GetDatabaseClusterBackup returns the specified cluster backup on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseClusterBackup(ctx echo.Context, kubernetesID string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesID, name)
}
