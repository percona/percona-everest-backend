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
	"encoding/json"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/pkg/errors"

	"github.com/percona/percona-everest-backend/model"
	"github.com/percona/percona-everest-backend/pkg/configs"
	"github.com/percona/percona-everest-backend/pkg/kubernetes"
)

// CreateDatabaseCluster creates a new db cluster inside the given k8s cluster.
func (e *EverestServer) CreateDatabaseCluster(ctx echo.Context, kubernetesID string) error {
	dbc, err := getDBCfromContext(ctx)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not get DatabaseCluster from the request body"),
		})
	}

	names := backupStorageNamesFrom(*dbc)

	k, kubeClient, code, err := e.initKubeClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}

	err = e.createK8SBackupStorages(ctx.Request().Context(), k, kubeClient, names)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not create BackupStorage"),
		})
	}

	return e.proxyKubernetes(ctx, kubernetesID, "")
}

// ListDatabaseClusters List of the created database clusters on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusters(ctx echo.Context, kubernetesID string) error {
	return e.proxyKubernetes(ctx, kubernetesID, "")
}

// DeleteDatabaseCluster Create a database cluster on the specified kubernetes cluster.
func (e *EverestServer) DeleteDatabaseCluster(ctx echo.Context, kubernetesID string, name string) error {
	_, kubeClient, code, err := e.initKubeClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}

	db, err := kubeClient.GetDatabaseCluster(ctx.Request().Context(), name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get database cluster"),
		})
	}

	proxyErr := e.proxyKubernetes(ctx, kubernetesID, name)
	if proxyErr != nil {
		return proxyErr
	}

	go e.deleteK8SBackupStorages(context.Background(), kubeClient, db)

	if db.Spec.Monitoring != nil && db.Spec.Monitoring.MonitoringConfigName != "" {
		go e.deleteK8SMonitoringConfig(context.Background(), kubeClient, db.Spec.Monitoring.MonitoringConfigName)
	}

	return nil
}

// GetDatabaseCluster Get the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseCluster(ctx echo.Context, kubernetesID string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesID, name)
}

// UpdateDatabaseCluster Replace the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseCluster(ctx echo.Context, kubernetesID string, name string) error {
	dbc, err := getDBCfromContext(ctx)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not get DatabaseCluster from the request body"),
		})
	}

	k, kubeClient, code, err := e.initKubeClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}

	newNames := backupStorageNamesFrom(*dbc)
	err = e.updateK8SBackupStorages(ctx.Request().Context(), k, kubeClient, name, newNames)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not update BackupStorages"),
		})
	}

	return e.proxyKubernetes(ctx, kubernetesID, name)
}

// GetDatabaseClusterCredentials returns credentials for the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseClusterCredentials(ctx echo.Context, kubernetesID string, name string) error {
	k, kubeClient, code, err := e.initKubeClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}

	databaseCluster, err := kubeClient.GetDatabaseCluster(ctx.Request().Context(), name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	secret, err := kubeClient.GetSecret(ctx.Request().Context(), databaseCluster.Spec.Engine.UserSecretsName, k.Namespace)
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

func (e *EverestServer) createK8SBackupStorages(ctx context.Context, k *model.KubernetesCluster, kubeClient *kubernetes.Kubernetes, names map[string]struct{}) error {
	if len(names) == 0 {
		return nil
	}

	processed := make([]string, 0, len(names))
	for name := range names {
		bs, err := e.storage.GetBackupStorage(ctx, nil, name)
		if err != nil {
			return errors.Wrap(err, "Could not get backup storage")
		}

		err = kubeClient.EnsureConfigExists(ctx, bs, e.secretsStorage.GetSecret)
		if err != nil {
			e.rollbackCreatedBackupStorages(ctx, k, processed)
			return errors.Wrapf(err, "Could not create CRs for %s", name)
		}
		processed = append(processed, name)
	}
	return nil
}

func (e *EverestServer) rollbackCreatedBackupStorages(
	ctx context.Context, k *model.KubernetesCluster, toDelete []string,
) {
	for _, name := range toDelete {
		bs, err := e.storage.GetBackupStorage(ctx, nil, name)
		if err != nil {
			e.l.Error(errors.Wrap(err, "could not get backup storage"))
			continue
		}

		configs.DeleteConfigFromK8sClusters(
			ctx, []model.KubernetesCluster{*k}, bs,
			e.initKubeClient, kubernetes.IsBackupStorageConfigInUse, e.l,
		)
	}
}

func (e *EverestServer) deleteK8SMonitoringConfig(
	ctx context.Context, kubeClient *kubernetes.Kubernetes, name string,
) {
	i, err := e.storage.GetMonitoringInstance(name)
	if err != nil {
		e.l.Error(errors.Wrap(err, "could get monitoring instance"))
		return
	}

	err = kubeClient.DeleteConfig(ctx, i, func(ctx context.Context, name string) (bool, error) {
		return kubernetes.IsMonitoringConfigInUse(ctx, name, kubeClient)
	})
	if err != nil {
		e.l.Error(errors.Wrap(err, "could not delete monitoring config in Kubernetes"))
		return
	}
}

func (e *EverestServer) deleteK8SBackupStorages(
	ctx context.Context, kubeClient *kubernetes.Kubernetes, db *everestv1alpha1.DatabaseCluster,
) {
	names := kubernetes.BackupStorageNamesFromDBCluster(db)
	for name := range names {
		bs, err := e.storage.GetBackupStorage(ctx, nil, name)
		if err != nil {
			e.l.Error(errors.Wrap(err, "could not get backup storage"))
			continue
		}

		err = kubeClient.DeleteConfig(ctx, bs, func(ctx context.Context, name string) (bool, error) {
			return kubernetes.IsBackupStorageConfigInUse(ctx, name, kubeClient)
		})
		if err != nil {
			e.l.Error(errors.Wrap(err, "could not delete backup storage config in Kubernetes"))
			continue
		}
	}
}

func (e *EverestServer) deleteK8SBackupStorage(
	ctx context.Context, kubeClient *kubernetes.Kubernetes, name string,
) error {
	bs, err := e.storage.GetBackupStorage(ctx, nil, name)
	if err != nil {
		return errors.Wrap(err, "could not find backup storage")
	}

	err = kubeClient.DeleteConfig(ctx, bs, func(ctx context.Context, name string) (bool, error) {
		return kubernetes.IsBackupStorageConfigInUse(ctx, name, kubeClient)
	})
	if err != nil {
		return errors.Wrap(err, "could not delete config in Kubernetes")
	}

	return nil
}

func (e *EverestServer) rollbackDeletedK8SBackupStorages(ctx context.Context, toCreate []string, kubeClient *kubernetes.Kubernetes) {
	for _, name := range toCreate {
		bs, err := e.storage.GetBackupStorage(ctx, nil, name)
		if err != nil {
			e.l.Error(errors.Wrapf(err, "Could not get backup storage %s", name))
			continue
		}

		err = kubeClient.EnsureConfigExists(ctx, bs, e.secretsStorage.GetSecret)
		if err != nil {
			e.l.Error(errors.Wrapf(err, "Failed to rollback deleted BackupStorage %s", name))
		}
	}
}

func (e *EverestServer) updateK8SBackupStorages(c context.Context, k *model.KubernetesCluster, kubeClient *kubernetes.Kubernetes, dbClusterName string, newNames map[string]struct{}) error {
	if len(newNames) == 0 {
		return nil
	}

	// get the old database cluster
	oldCluster, err := kubeClient.GetDatabaseCluster(c, dbClusterName)
	if err != nil {
		return errors.Wrap(err, "Could not get old DBCluster")
	}

	// get the list of the BackupStorages that was used in the old cluster
	oldNames := withBackupStorageNamesFromDBCluster(make(map[string]struct{}), *oldCluster)

	// try to create all storages that are new
	toCreate := uniqueKeys(oldNames, newNames)
	processed := make([]string, 0, len(toCreate))
	for name := range toCreate {
		bs, err := e.storage.GetBackupStorage(c, nil, name)
		if err != nil {
			return errors.Wrap(err, "Could not get backup storage")
		}

		err = kubeClient.EnsureConfigExists(c, bs, e.secretsStorage.GetSecret)
		if err != nil {
			e.rollbackCreatedBackupStorages(c, k, processed)
			return errors.Wrapf(err, "Could not create CRs for %s", name)
		}
		processed = append(processed, name)
	}

	// try to delete all storages that are not mentioned in the updated dbCluster anymore
	tryingToDelete := uniqueKeys(newNames, oldNames)
	// get the list of the BackupStorage names that could be deleted
	toDelete, err := getAllowedToDeleteNames(c, kubeClient, dbClusterName, tryingToDelete)
	if err != nil {
		return errors.Wrap(err, "Failed to check BackupStorages before deletion")
	}

	processed = make([]string, 0, len(toDelete))
	for name := range toDelete {
		err = e.deleteK8SBackupStorage(c, kubeClient, name)
		if err != nil {
			e.rollbackDeletedK8SBackupStorages(c, processed, kubeClient)
			return errors.Wrapf(err, "Could not delete CRs for %s", name)
		}
		processed = append(processed, name)
	}
	return nil
}

func backupStorageNamesFrom(dbc DatabaseCluster) map[string]struct{} {
	names := make(map[string]struct{})
	if dbc.Spec == nil {
		return names
	}

	if dbc.Spec.DataSource != nil {
		names[dbc.Spec.DataSource.BackupStorageName] = struct{}{}
	}

	if dbc.Spec.Backup == nil || dbc.Spec.Backup.Schedules == nil {
		return names
	}
	for _, schedule := range *dbc.Spec.Backup.Schedules {
		names[schedule.BackupStorageName] = struct{}{}
	}

	return names
}

func getAllowedToDeleteNames(c context.Context, kubeClient *kubernetes.Kubernetes, dbClusterName string, subset map[string]struct{}) (map[string]struct{}, error) {
	// get all existing dbClusters
	clusters, err := kubeClient.ListDatabaseClusters(c)
	if err != nil {
		return nil, errors.Wrap(err, "Could not get db clusters list")
	}

	// figure out which one we're trying to delete and what is the list of the other clusters
	var toDeleteCluster everestv1alpha1.DatabaseCluster
	otherClusters := make([]everestv1alpha1.DatabaseCluster, 0, len(clusters.Items))
	for _, dbc := range clusters.Items {
		if dbc.Name != dbClusterName {
			otherClusters = append(otherClusters, dbc)
		} else {
			toDeleteCluster = dbc
		}
	}

	// figure out what BackupStorages are used by other DBClusters
	inUse := backupStorageNamesFromDBClustersList(otherClusters)
	//  figure out what BackupStorages are used in the cluster we're trying to delete
	toDelete := subset
	if toDelete == nil {
		toDelete = withBackupStorageNamesFromDBCluster(make(map[string]struct{}), toDeleteCluster)
	}

	// figure out what BackupStorages we're allowed to delete
	allowedToDelete := make(map[string]struct{}, len(toDelete))
	for name := range toDelete {
		if _, ok := inUse[name]; !ok {
			// add to the allowed list only that are not in use by other clusters
			allowedToDelete[name] = struct{}{}
		}
	}

	return allowedToDelete, nil
}

func withBackupStorageNamesFromDBCluster(existing map[string]struct{}, dbc everestv1alpha1.DatabaseCluster) map[string]struct{} {
	if dbc.Spec.DataSource != nil && dbc.Spec.DataSource.BackupStorageName != "" {
		existing[dbc.Spec.DataSource.BackupStorageName] = struct{}{}
	}

	for _, schedule := range dbc.Spec.Backup.Schedules {
		if schedule.BackupStorageName != "" {
			existing[schedule.BackupStorageName] = struct{}{}
		}
	}

	return existing
}

func backupStorageNamesFromDBClustersList(dbClusters []everestv1alpha1.DatabaseCluster) map[string]struct{} {
	names := make(map[string]struct{})

	for _, dbc := range dbClusters {
		names = withBackupStorageNamesFromDBCluster(names, dbc)
	}

	return names
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

func getDBCfromContext(ctx echo.Context) (*DatabaseCluster, error) {
	dbc := &DatabaseCluster{}
	// GetBody creates a copy of the body to avoid "spoiling" the request before proxing
	reader, err := ctx.Request().GetBody()
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(dbc); err != nil {
		return nil, err
	}
	return dbc, nil
}
