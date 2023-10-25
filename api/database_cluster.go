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
	"errors"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
)

// CreateDatabaseCluster creates a new db cluster inside the given k8s cluster.
func (e *EverestServer) CreateDatabaseCluster(ctx echo.Context, kubernetesID string) error {
	dbc := &DatabaseCluster{}
	if err := e.getBodyFromContext(ctx, dbc); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not get DatabaseCluster from the request body"),
		})
	}

	if err := e.validateDatabaseClusterCR(ctx, kubernetesID, dbc); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	//backupNames := backupStorageNamesFrom(dbc)
	//err = e.createK8SBackupStorages(ctx.Request().Context(), kubeClient, backupNames)
	//if err != nil {
	//	e.l.Error(err)
	//	return ctx.JSON(http.StatusInternalServerError, Error{
	//		Message: pointer.ToString("Could not create BackupStorage"),
	//	})
	//}

	// if monitoringName := monitoringNameFrom(dbc); monitoringName != "" {
	//	i, err := e.storage.GetMonitoringInstance(monitoringName)
	//	if err != nil {
	//		return ctx.JSON(http.StatusBadRequest, Error{
	//			Message: pointer.ToString("Could not find monitoring instance"),
	//		})
	//	}

	//	err = kubeClient.EnsureConfigExists(ctx.Request().Context(), i, e.secretsStorage.GetSecret)
	//	if err != nil {
	//		e.l.Error(err)
	//		return ctx.JSON(http.StatusBadRequest, Error{
	//			Message: pointer.ToString("Could not create monitoring config in Kubernetes"),
	//		})
	//	}
	//}

	return e.proxyKubernetes(ctx, kubernetesID, "")
}

// ListDatabaseClusters lists the created database clusters on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusters(ctx echo.Context, kubernetesID string) error {
	return e.proxyKubernetes(ctx, kubernetesID, "")
}

// DeleteDatabaseCluster deletes a database cluster on the specified kubernetes cluster.
func (e *EverestServer) DeleteDatabaseCluster(ctx echo.Context, kubernetesID string, name string) error {
	//db, err := e.kubeClient.GetDatabaseCluster(ctx.Request().Context(), name)
	//if err != nil {
	//	e.l.Error(err)
	//	return ctx.JSON(http.StatusInternalServerError, Error{
	//		Message: pointer.ToString("Could not get database cluster"),
	//	})
	//}

	proxyErr := e.proxyKubernetes(ctx, kubernetesID, name)
	if proxyErr != nil {
		return proxyErr
	}

	// At this point the proxy already sent a response to the API user.
	// We check if the response was successful to continue with cleanup.
	if ctx.Response().Status >= http.StatusMultipleChoices {
		return nil
	}

	// names := kubernetes.BackupStorageNamesFromDBCluster(db)
	// e.waitGroup.Add(1)
	// go e.deleteK8SBackupStorages(context.Background(), kubeClient, names)

	//if db.Spec.Monitoring != nil && db.Spec.Monitoring.MonitoringConfigName != "" {
	//	e.waitGroup.Add(1)
	//	go e.deleteK8SMonitoringConfig(context.Background(), kubeClient, db.Spec.Monitoring.MonitoringConfigName)
	//}

	return nil
}

// GetDatabaseCluster retrieves the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseCluster(ctx echo.Context, kubernetesID string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesID, name)
}

// UpdateDatabaseCluster replaces the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseCluster(ctx echo.Context, kubernetesID string, name string) error {
	dbc := &DatabaseCluster{}
	if err := e.getBodyFromContext(ctx, dbc); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not get DatabaseCluster from the request body"),
		})
	}

	if err := e.validateDatabaseClusterCR(ctx, kubernetesID, dbc); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	oldDB, err := e.kubeClient.GetDatabaseCluster(ctx.Request().Context(), name)
	if err != nil {
		return errors.Join(err, errors.New("could not get old Database Cluster"))
	}
	if err := validateDatabaseClusterOnUpdate(dbc, oldDB); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	//newMonitoringName := monitoringNameFrom(dbc)
	//newBackupNames := backupStorageNamesFrom(dbc)
	//err = e.createResources(ctx.Request().Context(), oldDB, kubeClient, newMonitoringName, newBackupNames)
	//if err != nil {
	//	return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	//}

	proxyErr := e.proxyKubernetes(ctx, kubernetesID, name)
	if proxyErr != nil {
		return proxyErr
	}

	// At this point the proxy already sent a response to the API user.
	// We check if the response was successful to continue with cleanup.
	if ctx.Response().Status >= http.StatusMultipleChoices {
		return nil
	}
	// e.waitGroup.Add(1)
	// go e.deleteBackupStoragesOnUpdate(context.Background(), kubeClient, oldDB, newBackupNames)
	// e.waitGroup.Add(1)
	// go e.deleteMonitoringInstanceOnUpdate(context.Background(), kubeClient, oldDB, newMonitoringName)

	return nil
}

// GetDatabaseClusterCredentials returns credentials for the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseClusterCredentials(ctx echo.Context, kubernetesID string, name string) error {
	databaseCluster, err := e.kubeClient.GetDatabaseCluster(ctx.Request().Context(), name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	secret, err := e.kubeClient.GetSecret(ctx.Request().Context(), databaseCluster.Spec.Engine.UserSecretsName, "percona-everest") // FIXME
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	response := &DatabaseClusterCredential{}
	switch databaseCluster.Spec.Engine.Type {
	case everestv1alpha1.DatabaseEnginePXC:
		response.Username = pointer.ToString("root")
		response.Password = pointer.ToString(string(secret.Data["root"]))
	case everestv1alpha1.DatabaseEnginePSMDB:
		response.Username = pointer.ToString(string(secret.Data["MONGODB_USER_ADMIN_USER"]))
		response.Password = pointer.ToString(string(secret.Data["MONGODB_USER_ADMIN_PASSWORD"]))
	case everestv1alpha1.DatabaseEnginePostgresql:
		response.Username = pointer.ToString("postgres")
		response.Password = pointer.ToString(string(secret.Data["password"]))
	default:
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("Unsupported database engine")})
	}

	return ctx.JSON(http.StatusOK, response)
}

func backupStorageNamesFrom(dbc *DatabaseCluster) map[string]struct{} {
	names := make(map[string]struct{})
	if dbc.Spec == nil {
		return names
	}

	if dbc.Spec.DataSource != nil && dbc.Spec.DataSource.BackupSource != nil {
		names[dbc.Spec.DataSource.BackupSource.BackupStorageName] = struct{}{}
	}

	if dbc.Spec.Backup == nil || dbc.Spec.Backup.Schedules == nil {
		return names
	}
	for _, schedule := range *dbc.Spec.Backup.Schedules {
		names[schedule.BackupStorageName] = struct{}{}
	}

	return names
}

func monitoringNameFrom(db *DatabaseCluster) string {
	if db.Spec == nil {
		return ""
	}

	if db.Spec.Monitoring == nil {
		return ""
	}
	if db.Spec.Monitoring.MonitoringConfigName == nil {
		return ""
	}

	return *db.Spec.Monitoring.MonitoringConfigName
}

func withBackupStorageNamesFromDBCluster(existing map[string]struct{}, dbc everestv1alpha1.DatabaseCluster) map[string]struct{} {
	if dbc.Spec.DataSource != nil && dbc.Spec.DataSource.BackupSource != nil && dbc.Spec.DataSource.BackupSource.BackupStorageName != "" {
		existing[dbc.Spec.DataSource.BackupSource.BackupStorageName] = struct{}{}
	}

	for _, schedule := range dbc.Spec.Backup.Schedules {
		if schedule.BackupStorageName != "" {
			existing[schedule.BackupStorageName] = struct{}{}
		}
	}

	return existing
}

func uniqueKeys(source, target map[string]struct{}) map[string]struct{} {
	keysNotInSource := make(map[string]struct{}, len(target))
	for key := range target {
		if _, exists := source[key]; !exists {
			keysNotInSource[key] = struct{}{}
		}
	}
	return keysNotInSource
}
