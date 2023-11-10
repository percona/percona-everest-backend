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

package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListBackupStorages lists backup storages.
func (e *EverestServer) ListBackupStorages(ctx echo.Context) error {
	backupList, err := e.kubeClient.ListBackupStorages(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not list backup storages"),
		})
	}

	result := make([]BackupStorage, 0, len(backupList.Items))
	for _, bs := range backupList.Items {
		s := bs
		result = append(result, BackupStorage{
			Type:        BackupStorageType(bs.Spec.Type),
			Name:        s.Name,
			Description: &s.Spec.Description,
			BucketName:  s.Spec.Bucket,
			Region:      s.Spec.Region,
			Url:         &s.Spec.EndpointURL,
			Status:      s.Status.Status,
		})
	}

	return ctx.JSON(http.StatusOK, result)
}

// CreateBackupStorage creates a new backup storage object.
func (e *EverestServer) CreateBackupStorage(ctx echo.Context) error {
	params, err := validateCreateBackupStorageRequest(ctx, e.l)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	c := ctx.Request().Context()
	s, err := e.kubeClient.GetBackupStorage(c, params.Name)
	if err != nil && !k8serrors.IsNotFound(err) {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed getting a backup storage from the Kubernetes cluster"),
		})
	}
	// TODO: Change the design of operator's structs so they return nil struct so
	// if s != nil passes
	if s != nil && s.Name != "" {
		return ctx.JSON(http.StatusConflict, Error{
			Message: pointer.ToString(fmt.Sprintf("Backup storage %s already exists", params.Name)),
		})
	}

	bsData := &everestv1alpha1.BackupStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: e.kubeClient.Namespace(),
		},
		Spec: everestv1alpha1.BackupStorageSpec{
			Type:                  everestv1alpha1.BackupStorageType(params.Type),
			Bucket:                params.BucketName,
			Region:                params.Region,
			CredentialsSecretName: "",
		},
		Status: {
			Status: "Initializing",
		},
	}
	if params.Url != nil {
		bsData.Spec.EndpointURL = *params.Url
	}
	if params.Description != nil {
		bsData.Spec.Description = *params.Description
	}
	bs, err := e.kubeClient.CreateBackupStorage(c, bsData)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not create a backup storage"),
		})
	}

	secret, err := e.createBackupStorageSecret(c, bs, params.AccessKey, params.SecretKey)
	if err != nil {
		e.l.Error(err)

		if err := e.kubeClient.DeleteBackupStorage(c, bs.Name); err != nil {
			// TODO: We shall clean up the backup storage. Options:
			// 1. Retry delete here
			// 2. Backend to poll for not-ready storages and delete
			// 3. Operator to poll for not-ready storages and delete
			e.l.Errorf("Could not delete backup storage %s in initializing state", bs.Name)
		}

		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not create a secret for the backup storage"),
		})
	}

	bs.Status.Status = "Ready"
	bs.Spec.CredentialsSecretName = secret.Name

	if err := e.kubeClient.UpdateBackupStorage(c, bs); err != nil {
		e.l.Error(err)
		if err := e.kubeClient.DeleteBackupStorage(c, bs.Name); err != nil {
			// TODO: We shall clean up the backup storage. Options:
			// 1. Retry delete here
			// 2. Backend to poll for not-ready storages and delete
			// 3. Operator to poll for not-ready storages and delete
			e.l.Errorf("Could not delete backup storage %s in initializing state", bs.Name)
		}

		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not finish backup storage configuration"),
		})

	}

	result := BackupStorage{
		Type:        BackupStorageType(bs.Spec.Type),
		Name:        bs.Name,
		Description: &bs.Spec.Description,
		BucketName:  bs.Spec.Bucket,
		Region:      bs.Spec.Region,
		Url:         &bs.Spec.EndpointURL,
		Status:      bs.Status.Status,
	}

	return ctx.JSON(http.StatusOK, result)
}

// DeleteBackupStorage deletes the specified backup storage.
func (e *EverestServer) DeleteBackupStorage(ctx echo.Context, backupStorageName string) error {
	used, err := e.kubeClient.IsBackupStorageUsed(ctx.Request().Context(), backupStorageName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctx.JSON(http.StatusNotFound, Error{
				Message: pointer.ToString("Backup storage not found"),
			})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to check the backup storage is used"),
		})
	}
	if used {
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString(fmt.Sprintf("Backup storage %s is in use", backupStorageName)),
		})
	}
	if err := e.kubeClient.DeleteBackupStorage(ctx.Request().Context(), backupStorageName); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctx.NoContent(http.StatusNoContent)
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to delete a backup storage"),
		})
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (e *EverestServer) backupSecretData(secretKey, accessKey string) map[string]string {
	return map[string]string{
		"AWS_SECRET_ACCESS_KEY": secretKey,
		"AWS_ACCESS_KEY_ID":     accessKey,
	}
}

// GetBackupStorage retrieves the specified backup storage.
func (e *EverestServer) GetBackupStorage(ctx echo.Context, backupStorageName string) error {
	s, err := e.kubeClient.GetBackupStorage(ctx.Request().Context(), backupStorageName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctx.JSON(http.StatusNotFound, Error{
				Message: pointer.ToString("Backup storage not found"),
			})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed getting backup storage"),
		})
	}
	return ctx.JSON(http.StatusOK, BackupStorage{
		Type:        BackupStorageType(s.Spec.Type),
		Name:        s.Name,
		Description: &s.Spec.Description,
		BucketName:  s.Spec.Bucket,
		Region:      s.Spec.Region,
		Url:         &s.Spec.EndpointURL,
		Status:      s.Status.Status,
	})
}

// UpdateBackupStorage updates of the specified backup storage.
func (e *EverestServer) UpdateBackupStorage(ctx echo.Context, backupStorageName string) error { //nolint:funlen
	c := ctx.Request().Context()
	bs, err := e.kubeClient.GetBackupStorage(c, backupStorageName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctx.JSON(http.StatusNotFound, Error{
				Message: pointer.ToString("Backup storage not found"),
			})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed getting backup storage"),
		})
	}

	if bs.Status.Status != "Ready" {
		return ctx.JSON(http.StatusNotFound, Error{
			Message: pointer.ToString("Backup storage is not ready"),
		})
	}

	params, err := validateUpdateBackupStorageRequest(ctx, bs, e.l)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	var secret *corev1.Secret
	if params.AccessKey != nil && params.SecretKey != nil {
		secret, err = e.createBackupStorageSecret(c, bs, *params.AccessKey, *params.SecretKey)
		if err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{
				Message: pointer.ToString(fmt.Sprintf("Could not create secret for backup storage %s", backupStorageName)),
			})
		}
	}
	if params.BucketName != nil {
		bs.Spec.Bucket = *params.BucketName
	}
	if params.Region != nil {
		bs.Spec.Region = *params.Region
	}
	if params.Url != nil {
		bs.Spec.EndpointURL = *params.Url
	}
	if params.Description != nil {
		bs.Spec.Description = *params.Description
	}

	var prevSecretName string
	if secret != nil {
		prevSecretName = bs.Spec.CredentialsSecretName
		bs.Spec.CredentialsSecretName = secret.Name
	}

	if err = e.kubeClient.UpdateBackupStorage(c, bs); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed updating backup storage"),
		})
	}

	if prevSecretName != "" {
		if err := e.kubeClient.DeleteSecret(c, prevSecretName); err != nil {
			// TODO: We shall clean up the old secret. Options:
			// 1. Retry delete here
			// 2. Backend to poll for not-used secrets and delete
			// 3. Operator to poll for not-used secrets and delete
			e.l.Errorf("Could not delete secret %s", prevSecretName)
		}
	}

	result := BackupStorage{
		Type:        BackupStorageType(bs.Spec.Type),
		Name:        bs.Name,
		Description: params.Description,
		BucketName:  bs.Spec.Bucket,
		Region:      bs.Spec.Region,
		Url:         &bs.Spec.EndpointURL,
	}

	return ctx.JSON(http.StatusOK, result)
}

func (e *EverestServer) createBackupStorageSecret(
	ctx context.Context, bs *everestv1alpha1.BackupStorage,
	accessKey, secretKey string,
) (*corev1.Secret, error) {
	secretData := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: bs.Name + "-",
			Namespace:    e.kubeClient.Namespace(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: bs.APIVersion,
					Kind:       bs.Kind,
					Name:       bs.Name,
					UID:        bs.UID,
				},
			},
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: e.backupSecretData(secretKey, accessKey),
	}

	return e.kubeClient.CreateSecret(ctx, secretData)
}
