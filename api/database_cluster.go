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
	"fmt"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/percona/percona-everest-backend/model"
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

	names := objectStorageNamesFrom(*dbc)
	err = e.createBackupStorages(ctx.Request().Context(), kubernetesID, names)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not create ObjectStorage"),
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
	err := e.deleteBackupStorages(ctx.Request().Context(), kubernetesID, name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not delete ObjectStorages"),
		})
	}
	return e.proxyKubernetes(ctx, kubernetesID, name)
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

	newNames := objectStorageNamesFrom(*dbc)
	err = e.updateBackupStorages(ctx.Request().Context(), kubernetesID, name, newNames)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not update ObjectStorages"),
		})
	}

	return e.proxyKubernetes(ctx, kubernetesID, name)
}

// GetDatabaseClusterCredentials returns credentials for the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseClusterCredentials(ctx echo.Context, kubernetesID string, name string) error {
	k, kubeClient, code, err := e.initKubeClient(ctx, kubernetesID)
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

func (e *EverestServer) createBackupStorages(c context.Context, kubernetesID string, names map[string]struct{}) error {
	if len(names) == 0 {
		return nil
	}

	k, err := e.storage.GetKubernetesCluster(c, kubernetesID)
	if err != nil {
		return errors.Wrap(err, "Could not create k8s cluster")
	}
	everestClient, err := kubernetes.NewFromSecretsStorage(
		c, e.secretsStorage, k.ID, k.Namespace, e.l)
	if err != nil {
		return errors.Wrap(err, "Could not create k8s client")
	}

	processed := make([]string, 0, len(names))
	for name := range names {
		err = e.createBackupStorage(c, everestClient, name, k.Namespace)
		if err != nil {
			e.rollbackCreatedBackupStorages(c, processed, everestClient, k.Namespace)
			return errors.Wrap(err, fmt.Sprintf("Could not create CRs for %s", name))
		}
		processed = append(processed, name)
	}
	return nil
}

func (e *EverestServer) createBackupStorage(c context.Context, everestClient *kubernetes.Kubernetes, name, namespace string) error {
	// if storage already exists - do nothing
	_, err := everestClient.GetObjectStorage(c, name, namespace)
	if err == nil {
		return nil
	}
	if !k8serrors.IsNotFound(err) {
		return errors.Wrap(err, "Could not check if ObjectStorage exists")
	}

	// get the storage data from database
	bstorage, err := e.storage.GetBackupStorage(c, name)
	if err != nil {
		return errors.Wrap(err, "Could not get backup storage")
	}

	// get the storage secrets data from secrets storage
	secrets, err := e.getStorageSecrets(c, *bstorage)
	if err != nil {
		return errors.Wrap(err, "Failed to get secret from secrets storage")
	}

	// create the storage and the related secrets
	err = everestClient.CreateObjectStorage(c, namespace, *bstorage, secrets)
	if err != nil {
		return errors.Wrap(err, "Could not create ObjectStorage")
	}

	return nil
}

func (e *EverestServer) rollbackCreatedBackupStorages(c context.Context, toDelete []string, everestClient *kubernetes.Kubernetes, namespace string) {
	for _, name := range toDelete {
		err := e.deleteBackupStorage(c, everestClient, name, namespace, nil)
		if err != nil {
			e.l.Error(errors.Wrap(err, fmt.Sprintf("Failed to rollback created ObjectStorage %s", name)))
		}
	}
}

func (e *EverestServer) deleteBackupStorages(c context.Context, kubernetesID string, dbClusterName string) error {
	// create everest k8s client for the current kubernetesID
	k, err := e.storage.GetKubernetesCluster(c, kubernetesID)
	if err != nil {
		return errors.Wrap(err, "Could not create k8s cluster")
	}
	everestClient, err := kubernetes.NewFromSecretsStorage(
		c, e.secretsStorage, k.ID, k.Namespace, e.l)
	if err != nil {
		return errors.Wrap(err, "Could not create k8s client")
	}

	// get the list of the ObjectStorage names that should be deleted along with the cluster
	names, err := getAllowedToDeleteNames(c, everestClient, dbClusterName, nil)
	if err != nil {
		return errors.Wrap(err, "Failed to check ObjectStorages before deletion")
	}

	for name := range names {
		err = e.deleteBackupStorage(c, everestClient, name, k.Namespace, &dbClusterName)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Could not delete CRs for %s", name))
		}
	}
	return nil
}

func (e *EverestServer) deleteBackupStorage(c context.Context, everestClient *kubernetes.Kubernetes, name, namespace string, parentDBCluster *string) error {
	var exceptCluster string
	if parentDBCluster != nil {
		exceptCluster = *parentDBCluster
	}

	err := everestClient.DeleteObjectStorage(c, name, namespace, exceptCluster)
	if err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrap(err, "Could not delete backup storage")
	}

	return nil
}

func (e *EverestServer) rollbackDeletedBackupStorages(c context.Context, toDelete []string, everestClient *kubernetes.Kubernetes, namespace string) {
	for _, name := range toDelete {
		err := e.createBackupStorage(c, everestClient, name, namespace)
		if err != nil {
			e.l.Error(errors.Wrap(err, fmt.Sprintf("Failed to rollback deleted ObjectStorage %s", name)))
		}
	}
}

func (e *EverestServer) updateBackupStorages(c context.Context, kubernetesID, dbClusterName string, newNames map[string]struct{}) error {
	if len(newNames) == 0 {
		return nil
	}

	k, err := e.storage.GetKubernetesCluster(c, kubernetesID)
	if err != nil {
		return errors.Wrap(err, "Could not create k8s cluster")
	}
	everestClient, err := kubernetes.NewFromSecretsStorage(
		c, e.secretsStorage, k.ID, k.Namespace, e.l)
	if err != nil {
		return errors.Wrap(err, "Could not create k8s client")
	}

	// get the old database cluster
	oldCluster, err := everestClient.GetDatabaseCluster(c, dbClusterName)
	if err != nil {
		return errors.Wrap(err, "Could not get old DBCluster")
	}

	// get the list of the ObjectStorages that was used in the old cluster
	oldNames := withObjectStorageNamesFromDBCluster(make(map[string]struct{}), *oldCluster)

	// try to create all storages that are new
	toCreate := uniqueKeys(oldNames, newNames)
	processed := make([]string, 0, len(toCreate))
	for name := range toCreate {
		err = e.createBackupStorage(c, everestClient, name, k.Namespace)
		if err != nil {
			e.rollbackCreatedBackupStorages(c, processed, everestClient, k.Namespace)
			return errors.Wrap(err, fmt.Sprintf("Could not create CRs for %s", name))
		}
		processed = append(processed, name)
	}

	// try to delete all storages that are not mentioned in the updated dbCluster anymore
	tryingToDelete := uniqueKeys(newNames, oldNames)
	// get the list of the ObjectStorage names that could be deleted
	toDelete, err := getAllowedToDeleteNames(c, everestClient, dbClusterName, tryingToDelete)
	if err != nil {
		return errors.Wrap(err, "Failed to check ObjectStorages before deletion")
	}

	processed = make([]string, 0, len(toDelete))
	for name := range toDelete {
		err = e.deleteBackupStorage(c, everestClient, name, k.Namespace, &oldCluster.Name)
		if err != nil {
			e.rollbackDeletedBackupStorages(c, processed, everestClient, k.Namespace)
			return errors.Wrap(err, fmt.Sprintf("Could not delete CRs for %s", name))
		}
		processed = append(processed, name)
	}
	return nil
}

func (e *EverestServer) getStorageSecrets(ctx context.Context, bs model.BackupStorage) (map[string]string, error) {
	secretKey, err := e.secretsStorage.GetSecret(ctx, bs.SecretKeyID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get secretKey")
	}
	accessKey, err := e.secretsStorage.GetSecret(ctx, bs.AccessKeyID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get accessKey")
	}
	return map[string]string{
		bs.SecretKeyID: secretKey,
		bs.AccessKeyID: accessKey,
	}, nil
}

func objectStorageNamesFrom(dbc DatabaseCluster) map[string]struct{} {
	names := make(map[string]struct{})
	if dbc.Spec == nil {
		return names
	}

	if dbc.Spec.DataSource != nil {
		names[dbc.Spec.DataSource.ObjectStorageName] = struct{}{}
	}

	if dbc.Spec.Backup == nil || dbc.Spec.Backup.Schedules == nil {
		return names
	}
	for _, schedule := range *dbc.Spec.Backup.Schedules {
		names[schedule.ObjectStorageName] = struct{}{}
	}

	return names
}

func getAllowedToDeleteNames(c context.Context, everestClient *kubernetes.Kubernetes, dbClusterName string, subset map[string]struct{}) (map[string]struct{}, error) {
	// get all existing dbClusters
	clusters, err := everestClient.ListDatabaseClusters(c)
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

	// figure out what ObjectStorages are used by other DBClusters
	inUse := objectStorageNamesFromDBClustersList(otherClusters)
	//  figure out what ObjectStorages are used in the cluster we're trying to delete
	var toDelete = subset
	if toDelete == nil {
		toDelete = withObjectStorageNamesFromDBCluster(make(map[string]struct{}), toDeleteCluster)
	}

	// figure out what ObjectStorages we're allowed to delete
	allowedToDelete := make(map[string]struct{}, len(toDelete))
	for name := range toDelete {
		if _, ok := inUse[name]; !ok {
			// add to the allowed list only that are not in use by other clusters
			allowedToDelete[name] = struct{}{}
		}
	}

	return allowedToDelete, nil
}

func withObjectStorageNamesFromDBCluster(existing map[string]struct{}, dbc everestv1alpha1.DatabaseCluster) map[string]struct{} {
	if dbc.Spec.DataSource != nil && dbc.Spec.DataSource.ObjectStorageName != "" {
		existing[dbc.Spec.DataSource.ObjectStorageName] = struct{}{}
	}

	for _, schedule := range dbc.Spec.Backup.Schedules {
		if schedule.ObjectStorageName != "" {
			existing[schedule.ObjectStorageName] = struct{}{}
		}
	}

	return existing
}

func objectStorageNamesFromDBClustersList(dbClusters []everestv1alpha1.DatabaseCluster) map[string]struct{} {
	names := make(map[string]struct{})

	for _, dbc := range dbClusters {
		names = withObjectStorageNamesFromDBCluster(names, dbc)
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
