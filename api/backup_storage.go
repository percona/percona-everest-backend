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
	"github.com/labstack/echo/v4"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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
			Type: BackupStorageType(bs.Spec.Type),
			Name: s.Name,
			// Description: &s.Spec.Description, //FIXME
			BucketName: s.Spec.Bucket,
			Region:     s.Spec.Region,
			Url:        &s.Spec.EndpointURL,
		})
	}

	return ctx.JSON(http.StatusOK, result)
}

// CreateBackupStorage creates a new backup storage object.
// Rollbacks are implemented without transactions bc the secrets storage is going to be moved out of pg.
func (e *EverestServer) CreateBackupStorage(ctx echo.Context) error {
	params, err := validateCreateBackupStorageRequest(ctx, e.l)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	c := ctx.Request().Context()
	s, err := e.kubeClient.GetBackupStorage(c, params.Name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
		})
	}
	if s != nil {
		err = fmt.Errorf("storage %s already exists", params.Name)
		e.l.Error(err)
		return ctx.JSON(http.StatusConflict, Error{Message: pointer.ToString(err.Error())})
	}
	_, err = e.kubeClient.CreateSecret(c, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-secret", params.Name),
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"AWS_SECRET_ACCESS_KEY": params.AccessKey,
			"AWS_ACCESS_KEY_ID":     params.SecretKey,
		},
	})
	if err != nil {

		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
		})
	}
	err = e.kubeClient.CreateBackupStorage(c, &everestv1alpha1.BackupStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name: params.Name,
		},
		Spec: everestv1alpha1.BackupStorageSpec{
			Type:                  everestv1alpha1.BackupStorageType(params.Type),
			Bucket:                params.BucketName,
			Region:                params.Region,
			EndpointURL:           *params.Url,
			CredentialsSecretName: fmt.Sprintf("%s-secret", params.Name),
		},
	})
	if err != nil {

		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
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
	if err := e.kubeClient.DeleteBackupStorage(ctx.Request().Context(), backupStorageName); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
		})
	}
	if err := e.kubeClient.DeleteSecret(ctx.Request().Context(), fmt.Sprintf("%s-secret", backupStorageName)); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
		})
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetBackupStorage retrieves the specified backup storage.
func (e *EverestServer) GetBackupStorage(ctx echo.Context, backupStorageID string) error {
	// TODO: Implement the logic
	return ctx.JSON(http.StatusOK, nil)
}

// UpdateBackupStorage updates of the specified backup storage.
func (e *EverestServer) UpdateBackupStorage(ctx echo.Context, backupStorageName string) error {
	params, err := validateUpdateBackupStorageRequest(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	c := ctx.Request().Context()
	if params.AccessKey != nil && params.SecretKey != nil {
		_, err = e.kubeClient.UpdateSecret(c, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-secret", backupStorageName),
			},
			Type: corev1.SecretTypeOpaque,
			StringData: map[string]string{
				"AWS_SECRET_ACCESS_KEY": *params.AccessKey,
				"AWS_ACCESS_KEY_ID":     *params.SecretKey,
			},
		})
	}
	if err != nil {

		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
		})
	}
	err = e.kubeClient.UpdateBackupStorage(c, &everestv1alpha1.BackupStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name: backupStorageName,
		},
		Spec: everestv1alpha1.BackupStorageSpec{
			Bucket:                *params.BucketName,
			Region:                *params.Region,
			EndpointURL:           *params.Url,
			CredentialsSecretName: fmt.Sprintf("%s-secret", backupStorageName),
		},
	})
	if err != nil {

		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
		})
	}
	//err = validateStorageAccessByUpdate(ctx, oldData, params, e.l)
	//if err != nil {
	//	return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	//}

	return nil
}
