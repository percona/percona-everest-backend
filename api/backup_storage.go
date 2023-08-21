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
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/percona/percona-everest-backend/model"
	"github.com/percona/percona-everest-backend/pkg/configs"
	"github.com/percona/percona-everest-backend/pkg/kubernetes"
)

// ListBackupStorages lists backup storages.
func (e *EverestServer) ListBackupStorages(ctx echo.Context) error {
	list, err := e.storage.ListBackupStorages(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not list backup storages"),
		})
	}

	result := make([]BackupStorage, 0, len(list))
	for _, bs := range list {
		s := bs
		result = append(result, BackupStorage{
			Type:        BackupStorageType(bs.Type),
			Name:        s.Name,
			Description: &s.Description,
			BucketName:  s.BucketName,
			Region:      s.Region,
			Url:         &s.URL,
		})
	}

	return ctx.JSON(http.StatusOK, result)
}

// CreateBackupStorage creates a new backup storage object.
// Rollbacks are implemented without transactions bc the secrets storage is going to be moved out of pg.
func (e *EverestServer) CreateBackupStorage(ctx echo.Context) error { //nolint:funlen,cyclop
	params, err := validateCreateBackupStorageRequest(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	c := ctx.Request().Context()

	existingStorage, err := e.storage.GetBackupStorage(c, nil, params.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Failed to get BackupStorage"),
		})
	}
	if existingStorage != nil {
		err = errors.Errorf("Storage %s already exists", params.Name)
		e.l.Error(err)
		return ctx.JSON(http.StatusConflict, Error{Message: pointer.ToString(err.Error())})
	}

	var accessKeyID, secretKeyID *string
	defer e.cleanUpNewSecretsOnUpdateError(err, accessKeyID, secretKeyID)

	accessKeyID, secretKeyID, err = e.createSecrets(c, &params.AccessKey, &params.SecretKey)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	var url string
	if params.Url != nil {
		url = *params.Url
	}

	var description string
	if params.Description != nil {
		description = *params.Description
	}

	s, err := e.storage.CreateBackupStorage(c, model.CreateBackupStorageParams{
		Name:        params.Name,
		Description: description,
		Type:        string(params.Type),
		BucketName:  params.BucketName,
		URL:         url,
		Region:      params.Region,
		AccessKeyID: *accessKeyID,
		SecretKeyID: *secretKeyID,
	})
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code.Name() == pgErrUniqueViolation {
				return ctx.JSON(http.StatusBadRequest, Error{
					Message: pointer.ToString("Backup storage with the same name already exists. " + pgErr.Detail),
				})
			}
		}

		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not create a new backup storage"),
		})
	}

	result := BackupStorage{
		Type:        BackupStorageType(s.Type),
		Name:        s.Name,
		Description: &s.Description,
		BucketName:  s.BucketName,
		Region:      s.Region,
		Url:         &s.URL,
	}

	return ctx.JSON(http.StatusOK, result)
}

// DeleteBackupStorage deletes the specified backup storage.
func (e *EverestServer) DeleteBackupStorage(ctx echo.Context, backupStorageName string) error {
	c := ctx.Request().Context()
	bs, err := e.storage.GetBackupStorage(c, nil, backupStorageName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString("Could not find backup storage")})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get backup storage"),
		})
	}

	err = e.storage.Transaction(func(tx *gorm.DB) error {
		err := e.storage.DeleteBackupStorage(c, backupStorageName, tx)
		if err != nil {
			e.l.Error(err)
			return errors.New("Could not delete backup storage")
		}

		ks, err := e.storage.ListKubernetesClusters(c)
		if err != nil {
			return errors.Wrap(err, "Could not list Kubernetes clusters")
		}

		go configs.DeleteConfigFromK8sClusters( //nolint:contextcheck
			context.Background(), ks, bs,
			e.initKubeClient, kubernetes.IsBackupStorageConfigInUse, e.l,
		)

		return nil
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString(err.Error()),
		})
	}

	go func() {
		if _, err := e.secretsStorage.DeleteSecret(c, bs.AccessKeyID); err != nil {
			e.l.Error(errors.Wrap(err, "could not delete access key from secrets storage"))
		}
	}()

	go func() {
		if _, err := e.secretsStorage.DeleteSecret(c, bs.SecretKeyID); err != nil {
			e.l.Error(errors.Wrap(err, "could not delete secret key from secrets storage"))
		}
	}()

	return ctx.NoContent(http.StatusNoContent)
}

// GetBackupStorage retrieves the specified backup storage.
func (e *EverestServer) GetBackupStorage(ctx echo.Context, backupStorageID string) error {
	s, err := e.storage.GetBackupStorage(ctx.Request().Context(), nil, backupStorageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString("Could not find backup storage")})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get backup storage"),
		})
	}

	result := BackupStorage{
		Description: &s.Description,
		Type:        BackupStorageType(s.Type),
		BucketName:  s.BucketName,
		Name:        s.Name,
		Region:      s.Region,
		Url:         &s.URL,
	}

	return ctx.JSON(http.StatusOK, result)
}

// UpdateBackupStorage updates of the specified backup storage.
func (e *EverestServer) UpdateBackupStorage(ctx echo.Context, backupStorageName string) error {
	params, err := validateUpdateBackupStorageRequest(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	c := ctx.Request().Context()

	// check data access
	s, err := e.checkStorageAccessByUpdate(c, backupStorageName, *params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString("Could not find backup storage")})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not connect to backup storage"),
		})
	}

	var newAccessKeyID, newSecretKeyID *string
	defer e.cleanUpNewSecretsOnUpdateError(err, newAccessKeyID, newSecretKeyID)

	newAccessKeyID, newSecretKeyID, err = e.createSecrets(c, params.AccessKey, params.SecretKey)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Failed to create secrets")})
	}

	httpStatusCode := http.StatusInternalServerError
	var bs *model.BackupStorage
	err = e.storage.Transaction(func(tx *gorm.DB) error {
		var err error
		httpStatusCode, err = e.updateBackupStorage(c, tx, backupStorageName, params, newAccessKeyID, newSecretKeyID)
		if err != nil {
			return err
		}

		bs, err = e.storage.GetBackupStorage(c, tx, backupStorageName)
		if err != nil {
			e.l.Error(err)
			return errors.New("Could not find updated backup storage")
		}

		ks, err := e.storage.ListKubernetesClusters(c)
		if err != nil {
			return errors.Wrap(err, "Could not list Kubernetes clusters")
		}

		go configs.UpdateConfigInAllK8sClusters( //nolint:contextcheck
			context.Background(), ks, bs,
			e.secretsStorage.GetSecret, e.initKubeClient, e.l,
		)

		return nil
	})
	if err != nil {
		return ctx.JSON(httpStatusCode, Error{
			Message: pointer.ToString(err.Error()),
		})
	}

	e.deleteOldSecretsAfterUpdate(c, params, s)

	result := BackupStorage{
		Type:        BackupStorageType(bs.Type),
		Name:        bs.Name,
		Description: &bs.Description,
		BucketName:  bs.BucketName,
		Region:      bs.Region,
		Url:         &bs.URL,
	}

	return ctx.JSON(http.StatusOK, result)
}

func (e *EverestServer) createSecrets(
	ctx context.Context,
	accessKey, secretKey *string,
) (*string, *string, error) {
	var newAccessKeyID, newSecretKeyID *string
	if accessKey != nil {
		newID := uuid.NewString()
		newAccessKeyID = &newID

		// create new AccessKey
		err := e.secretsStorage.CreateSecret(ctx, newID, *accessKey)
		if err != nil {
			e.l.Error(err)
			return newAccessKeyID, newSecretKeyID, errors.New("Could not store access key in secrets storage")
		}
	}

	if secretKey != nil {
		newID := uuid.NewString()
		newSecretKeyID = &newID

		// create new SecretKey
		err := e.secretsStorage.CreateSecret(ctx, newID, *secretKey)
		if err != nil {
			e.l.Error(err)
			return newAccessKeyID, newSecretKeyID, errors.New("Could not store secret key in secrets storage")
		}
	}

	return newAccessKeyID, newSecretKeyID, nil
}

func (e *EverestServer) deleteOldSecretsAfterUpdate(ctx context.Context, params *UpdateBackupStorageParams, s *model.BackupStorage) {
	// delete old AccessKey
	if params.AccessKey != nil {
		_, cErr := e.secretsStorage.DeleteSecret(ctx, s.AccessKeyID)
		if cErr != nil {
			e.l.Errorf("Failed to delete unused secret, please delete it manually. id = %s", s.AccessKeyID)
		}
	}

	// delete old SecretKey
	if params.SecretKey != nil {
		_, cErr := e.secretsStorage.DeleteSecret(ctx, s.SecretKeyID)
		if cErr != nil {
			e.l.Errorf("Failed to delete unused secret, please delete it manually. id = %s", s.SecretKeyID)
		}
	}
}

func (e *EverestServer) cleanUpNewSecretsOnUpdateError(err error, newAccessKeyID, newSecretKeyID *string) {
	if err == nil {
		return
	}

	ctx := context.Background()

	// if an error appeared - cleanup the created secrets
	if newAccessKeyID != nil {
		_, err = e.secretsStorage.DeleteSecret(ctx, *newAccessKeyID)
		if err != nil {
			e.l.Errorf("Failed to delete unused secret, please delete it manually. id = %s", *newAccessKeyID)
		}
	}

	if newSecretKeyID != nil {
		_, err = e.secretsStorage.DeleteSecret(ctx, *newSecretKeyID)
		if err != nil {
			e.l.Errorf("Failed to delete unused secret, please delete it manually. id = %s", *newSecretKeyID)
		}
	}
}

func (e *EverestServer) checkStorageAccessByUpdate(ctx context.Context, storageName string, params UpdateBackupStorageParams) (*model.BackupStorage, error) {
	s, err := e.storage.GetBackupStorage(ctx, nil, storageName)
	if err != nil {
		return nil, err
	}

	accessKey, err := e.secretsStorage.GetSecret(ctx, s.AccessKeyID)
	if err != nil {
		return nil, err
	}

	secretKey, err := e.secretsStorage.GetSecret(ctx, s.SecretKeyID)
	if err != nil {
		return nil, err
	}

	oldData := &storageData{
		accessKey: accessKey,
		secretKey: secretKey,
		storage:   *s,
	}

	err = validateStorageAccessByUpdate(oldData, params)
	if err != nil {
		return nil, err
	}

	return &oldData.storage, nil
}

func (e *EverestServer) updateBackupStorage(
	ctx context.Context, tx *gorm.DB, backupStorageName string, params *UpdateBackupStorageParams,
	newAccessKeyID, newSecretKeyID *string,
) (int, error) {
	err := e.storage.UpdateBackupStorage(ctx, tx, model.UpdateBackupStorageParams{
		Name:        backupStorageName,
		Description: params.Description,
		BucketName:  params.BucketName,
		URL:         params.Url,
		Region:      params.Region,
		AccessKeyID: newAccessKeyID,
		SecretKeyID: newSecretKeyID,
	})
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code.Name() == pgErrUniqueViolation {
				return http.StatusBadRequest, errors.New("Backup storage with the same name already exists. " + pgErr.Detail)
			}
		}

		e.l.Error(err)
		return http.StatusInternalServerError, errors.New("Could not update backup storage")
	}

	return 0, nil
}
