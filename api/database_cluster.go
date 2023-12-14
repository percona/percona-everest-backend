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
	"sort"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
)

// CreateDatabaseCluster creates a new db cluster inside the given k8s cluster.
func (e *EverestServer) CreateDatabaseCluster(ctx echo.Context) error {
	dbc := &DatabaseCluster{}
	if err := e.getBodyFromContext(ctx, dbc); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not get DatabaseCluster from the request body"),
		})
	}

	if err := e.validateDatabaseClusterCR(ctx, dbc); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	return e.proxyKubernetes(ctx, "")
}

// ListDatabaseClusters lists the created database clusters on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusters(ctx echo.Context) error {
	return e.proxyKubernetes(ctx, "")
}

// DeleteDatabaseCluster deletes a database cluster on the specified kubernetes cluster.
func (e *EverestServer) DeleteDatabaseCluster(ctx echo.Context, name string) error {
	return e.proxyKubernetes(ctx, name)
}

// GetDatabaseCluster retrieves the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseCluster(ctx echo.Context, name string) error {
	return e.proxyKubernetes(ctx, name)
}

// UpdateDatabaseCluster replaces the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseCluster(ctx echo.Context, name string) error {
	dbc := &DatabaseCluster{}
	if err := e.getBodyFromContext(ctx, dbc); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not get DatabaseCluster from the request body"),
		})
	}

	if err := e.validateDatabaseClusterCR(ctx, dbc); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	oldDB, err := e.kubeClient.GetDatabaseCluster(ctx.Request().Context(), name)
	if err != nil {
		return errors.Join(err, errors.New("could not get old Database Cluster"))
	}
	if err := validateDatabaseClusterOnUpdate(dbc, oldDB); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	return e.proxyKubernetes(ctx, name)
}

// GetDatabaseClusterCredentials returns credentials for the specified database cluster.
func (e *EverestServer) GetDatabaseClusterCredentials(ctx echo.Context, name string) error {
	databaseCluster, err := e.kubeClient.GetDatabaseCluster(ctx.Request().Context(), name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	secret, err := e.kubeClient.GetSecret(ctx.Request().Context(), databaseCluster.Spec.Engine.UserSecretsName)
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

// GetDatabaseClusterPitr returns the point-in-time recovery related information for the specified database cluster.
func (e *EverestServer) GetDatabaseClusterPitr(ctx echo.Context, name string) error {
	databaseCluster, err := e.kubeClient.GetDatabaseCluster(ctx.Request().Context(), name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	response := &DatabaseClusterPitr{}
	if !databaseCluster.Spec.Backup.Enabled || !databaseCluster.Spec.Backup.PITR.Enabled {
		return ctx.JSON(http.StatusOK, response)
	}

	backups, err := e.kubeClient.ListDatabaseClusterBackups(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	if len(backups.Items) == 0 {
		return ctx.JSON(http.StatusOK, response)
	}

	latestBackup := latestSuccessfulBackup(backups.Items, databaseCluster.Spec.Engine.Type)

	uploadInterval := getDefaultUploadInterval(databaseCluster.Spec.Engine.Type)
	now := time.Now()
	// delete nanoseconds since they're not accepted by restoration
	now = now.Truncate(time.Duration(now.Nanosecond()) * time.Nanosecond)
	// heuristic: latest restorable date is now minus uploadInterval
	latest := now.Truncate(time.Duration(uploadInterval) * time.Second).UTC()
	response.LatestDate = &latest
	earliest := latestBackup.Status.CreatedAt.UTC()
	response.EarliestDate = &earliest
	response.LatestBackupName = &latestBackup.Name
	response.Gaps = &latestBackup.Status.Gaps

	return ctx.JSON(http.StatusOK, response)
}

func latestSuccessfulBackup(backups []everestv1alpha1.DatabaseClusterBackup, engineType everestv1alpha1.EngineType) *everestv1alpha1.DatabaseClusterBackup {
	sort.Sort(BackupsByCreatedAt(backups))
	for _, backup := range backups {
		if successStatus(backup.Status.State, engineType) {
			return &backup
		}
	}
	return nil
}

// BackupsByCreatedAt is a slice of DatabaseClusterBackup that can be sorted by the CreatedAt field.
type BackupsByCreatedAt []everestv1alpha1.DatabaseClusterBackup

func (a BackupsByCreatedAt) Len() int      { return len(a) }
func (a BackupsByCreatedAt) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BackupsByCreatedAt) Less(i, j int) bool {
	if a[i].Status.CreatedAt == nil {
		return true
	}
	if a[j].Status.CreatedAt == nil {
		return false
	}
	return a[i].Status.CreatedAt.After(a[j].Status.CreatedAt.Time)
}

func successStatus(state everestv1alpha1.BackupState, engineType everestv1alpha1.EngineType) bool {
	var successState string
	switch engineType {
	case everestv1alpha1.DatabaseEnginePXC:
		successState = "Succeeded"
	case everestv1alpha1.DatabaseEnginePSMDB:
		successState = "ready"
	case everestv1alpha1.DatabaseEnginePostgresql:
		successState = "Succeeded"
	}
	return string(state) == successState
}

func getDefaultUploadInterval(engineType everestv1alpha1.EngineType) int {
	switch engineType {
	case everestv1alpha1.DatabaseEnginePXC:
		return 60
	case everestv1alpha1.DatabaseEnginePSMDB:
		return 600
	case everestv1alpha1.DatabaseEnginePostgresql:
		return 60
	}
	return 0
}
