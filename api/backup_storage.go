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
		})
	}

	return ctx.JSON(http.StatusOK, result)
}

// CreateBackupStorage creates a new backup storage object.
func (e *EverestServer) CreateBackupStorage(ctx echo.Context) error { //nolint:funlen //FIXME
	params, err := validateCreateBackupStorageRequest(ctx, e.l)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	c := ctx.Request().Context()
	if code, err := e.checkBackupStorageExists(c, params); err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}
	_, err = e.kubeClient.CreateSecret(c, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: e.kubeClient.Namespace(),
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"AWS_SECRET_ACCESS_KEY": params.SecretKey,
			"AWS_ACCESS_KEY_ID":     params.AccessKey,
		},
	})
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed creating the secret for the backup storage"),
		})
	}
	bs := &everestv1alpha1.BackupStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: e.kubeClient.Namespace(),
		},
		Spec: everestv1alpha1.BackupStorageSpec{
			Type:                  everestv1alpha1.BackupStorageType(params.Type),
			Bucket:                params.BucketName,
			Region:                params.Region,
			CredentialsSecretName: params.Name,
		},
	}
	if params.Url != nil {
		bs.Spec.EndpointURL = *params.Url
	}
	if params.Description != nil {
		bs.Spec.Description = *params.Description
	}
	err = e.kubeClient.CreateBackupStorage(c, bs)
	if err != nil {
		e.l.Error(err)
		// TODO: Move this logic to the operator
		dErr := e.kubeClient.DeleteSecret(c, params.Name)
		if dErr != nil {
			return ctx.JSON(http.StatusInternalServerError, Error{
				Message: pointer.ToString("Failing cleaning up the secret because failed creating backup storage"),
			})
		}
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed creating backup storage"),
		})
	}
	result := BackupStorage{
		Type:        BackupStorageType(params.Type),
		Name:        params.Name,
		Description: params.Description,
		BucketName:  params.BucketName,
		Region:      params.Region,
		Url:         params.Url,
	}

	return ctx.JSON(http.StatusOK, result)
}

// DeleteBackupStorage deletes the specified backup storage.
func (e *EverestServer) DeleteBackupStorage(ctx echo.Context, backupStorageName string) error {
	used, err := e.kubeClient.BackupStorageIsUsed(ctx.Request().Context(), backupStorageName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctx.JSON(http.StatusNotFound, Error{
				Message: pointer.ToString("Backup storage is not found"),
			})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to check the backup storage is used"),
		})
	}
	if used {
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString(fmt.Sprintf("Backup storage %s is used", backupStorageName)),
		})
	}
	if err := e.kubeClient.DeleteBackupStorage(ctx.Request().Context(), backupStorageName); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctx.JSON(http.StatusNotFound, Error{
				Message: pointer.ToString("Backup storage is not found"),
			})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
		})
	}
	if err := e.kubeClient.DeleteSecret(ctx.Request().Context(), backupStorageName); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctx.JSON(http.StatusNotFound, Error{
				Message: pointer.ToString("Secret is not found"),
			})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
		})
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetBackupStorage retrieves the specified backup storage.
func (e *EverestServer) GetBackupStorage(ctx echo.Context, backupStorageName string) error {
	s, err := e.kubeClient.GetBackupStorage(ctx.Request().Context(), backupStorageName)
	if err != nil {
		e.l.Error(err)
		if k8serrors.IsNotFound(err) {
			return ctx.JSON(http.StatusNotFound, Error{
				Message: pointer.ToString("Backup storage is not found"),
			})
		}
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
	})
}

// UpdateBackupStorage updates of the specified backup storage.
func (e *EverestServer) UpdateBackupStorage(ctx echo.Context, backupStorageName string) error { //nolint:funlen //FIXME
	c := ctx.Request().Context()
	bs, err := e.kubeClient.GetBackupStorage(c, backupStorageName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctx.JSON(http.StatusNotFound, Error{
				Message: pointer.ToString("Backup storage is not found"),
			})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed getting backup storage"),
		})
	}
	params, err := validateUpdateBackupStorageRequest(ctx, bs, e.l)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	if params.AccessKey != nil && params.SecretKey != nil {
		_, err = e.kubeClient.UpdateSecret(c, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      backupStorageName,
				Namespace: e.kubeClient.Namespace(),
			},
			Type: corev1.SecretTypeOpaque,
			StringData: map[string]string{
				"AWS_SECRET_ACCESS_KEY": *params.SecretKey,
				"AWS_ACCESS_KEY_ID":     *params.AccessKey,
			},
		})
		if err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{
				Message: pointer.ToString("Failed updating the secret"),
			})
		}
	}
	if params.BucketName != nil {
		bs.Spec.Bucket = *params.BucketName
	}
	if params.Region != nil {
		bs.Spec.Bucket = *params.Region
	}
	if params.Url != nil {
		bs.Spec.EndpointURL = *params.Url
	}
	if params.Description != nil {
		bs.Spec.Description = *params.Description
	}

	err = e.kubeClient.UpdateBackupStorage(c, bs)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed updating backup storage"),
		})
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

func (e *EverestServer) checkBackupStorageExists(c context.Context, params *CreateBackupStorageParams) (int, error) {
	s, err := e.kubeClient.GetBackupStorage(c, params.Name)
	if err != nil && !k8serrors.IsNotFound(err) {
		e.l.Error(err)
		return http.StatusInternalServerError, errors.New("failed getting a backup storage from the Kubernetes cluster")
	}
	if s != nil && s.Name != "" {
		return http.StatusConflict, fmt.Errorf("storage %s already exists", params.Name)
	}
	return http.StatusOK, nil
}
