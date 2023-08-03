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
			Id:         s.ID,
			Type:       BackupStorageType(bs.Type),
			Name:       s.Name,
			BucketName: s.BucketName,
			Region:     s.Region,
			Url:        &s.URL,
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
	var accessKeyID, secretKeyID string

	defer func() {
		if err == nil {
			return
		}

		// rollback the changes - delete secrets
		if accessKeyID != "" {
			_, dError := e.secretsStorage.DeleteSecret(c, accessKeyID)
			if dError != nil {
				e.l.Errorf(
					"Failed to delete unused secret with id = %s. The secret needs to be deleted manually",
					accessKeyID,
				)
			}
		}

		if secretKeyID != "" {
			_, dError := e.secretsStorage.DeleteSecret(c, secretKeyID)
			if dError != nil {
				e.l.Errorf(
					"Failed to delete unused secret with id = %s. The secret needs to be deleted manually",
					secretKeyID,
				)
			}
		}
	}()

	accessKeyID = uuid.NewString()
	err = e.secretsStorage.CreateSecret(c, accessKeyID, params.AccessKey)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not store access key in secrets storage"),
		})
	}

	secretKeyID = uuid.NewString()
	err = e.secretsStorage.CreateSecret(c, secretKeyID, params.SecretKey)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not store secret key in secrets storage"),
		})
	}

	var url string
	if params.Url != nil {
		url = *params.Url
	}

	s, err := e.storage.CreateBackupStorage(c, model.CreateBackupStorageParams{
		Name:        params.Name,
		Type:        string(params.Type),
		BucketName:  params.BucketName,
		URL:         url,
		Region:      params.Region,
		AccessKeyID: accessKeyID,
		SecretKeyID: secretKeyID,
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
		Id:         s.ID,
		Type:       BackupStorageType(s.Type),
		Name:       s.Name,
		BucketName: s.BucketName,
		Region:     s.Region,
		Url:        &s.URL,
	}

	return ctx.JSON(http.StatusOK, result)
}

// DeleteBackupStorage deletes the specified backup storage.
func (e *EverestServer) DeleteBackupStorage(ctx echo.Context, backupStorageID string) error {
	c := ctx.Request().Context()
	bs, err := e.storage.GetBackupStorage(c, backupStorageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString("Could not find backup storage")})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get backup storage"),
		})
	}

	deletedAccessKey, err := e.secretsStorage.DeleteSecret(c, bs.AccessKeyID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not delete access key from secrets storage"),
		})
	}

	deletedSecretKey, err := e.secretsStorage.DeleteSecret(c, bs.SecretKeyID)
	if err != nil {
		e.l.Error(err)

		// rollback the changes - put the deleted secret back
		cErr := e.secretsStorage.CreateSecret(c, bs.SecretKeyID, deletedAccessKey)
		if cErr != nil {
			e.l.Errorf("Inconsistent DB state, manual intervention required. Can not revert changes over the secret with id = %s", bs.AccessKeyID)
		}
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not delete secret key from secrets storage"),
		})
	}

	err = e.storage.DeleteBackupStorage(c, backupStorageID)
	if err != nil {
		e.l.Error(err)

		// rollback the changes - put the deleted secrets back
		cErr := e.secretsStorage.CreateSecret(c, bs.AccessKeyID, deletedAccessKey)
		if cErr != nil {
			e.l.Errorf("Inconsistent DB state, manual intervention required. Can not revert changes over the secret with id = %s", bs.AccessKeyID)
		}
		cErr = e.secretsStorage.CreateSecret(c, bs.SecretKeyID, deletedSecretKey)
		if cErr != nil {
			e.l.Errorf("Inconsistent DB state, manual intervention required. Can not revert changes over the secret with id = %s", bs.SecretKeyID)
		}

		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not delete backup storage"),
		})
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetBackupStorage retrieves the specified backup storage.
func (e *EverestServer) GetBackupStorage(ctx echo.Context, backupStorageID string) error {
	s, err := e.storage.GetBackupStorage(ctx.Request().Context(), backupStorageID)
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
		Id:         s.ID,
		Type:       BackupStorageType(s.Type),
		BucketName: s.BucketName,
		Name:       s.Name,
		Region:     s.Region,
		Url:        &s.URL,
	}

	return ctx.JSON(http.StatusOK, result)
}

// UpdateBackupStorage updates of the specified backup storage.
func (e *EverestServer) UpdateBackupStorage(ctx echo.Context, backupStorageID string) error {
	params, err := validateUpdateBackupStorageRequest(ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	c := ctx.Request().Context()

	// check data access
	s, err := e.checkStorageAccessByUpdate(c, backupStorageID, *params)
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

	newAccessKeyID, newSecretKeyID, err = e.maybeCreateSecretsDuringUpdate(c, params)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	updated, httpStatusCode, err := e.updateBackupStorage(c, backupStorageID, params, newAccessKeyID, newSecretKeyID)
	if err != nil {
		return ctx.JSON(httpStatusCode, Error{Message: pointer.ToString(err.Error())})
	}

	e.deleteOldSecretsAfterUpdate(c, params, s)

	result := BackupStorage{
		Id:         updated.ID,
		Type:       BackupStorageType(updated.Type),
		Name:       updated.Name,
		BucketName: updated.BucketName,
		Region:     updated.Region,
		Url:        &updated.URL,
	}

	return ctx.JSON(http.StatusOK, result)
}

func (e *EverestServer) maybeCreateSecretsDuringUpdate(
	ctx context.Context, params *UpdateBackupStorageParams,
) (*string, *string, error) {
	var newAccessKeyID, newSecretKeyID *string
	if params.AccessKey != nil {
		newID := uuid.NewString()
		newAccessKeyID = &newID

		// create new AccessKey
		err := e.secretsStorage.CreateSecret(ctx, newID, *params.AccessKey)
		if err != nil {
			e.l.Error(err)
			return newAccessKeyID, newSecretKeyID, errors.New("Could not store access key in secrets storage")
		}
	}

	if params.SecretKey != nil {
		newID := uuid.NewString()
		newSecretKeyID = &newID

		// create new SecretKey
		err := e.secretsStorage.CreateSecret(ctx, newID, *params.SecretKey)
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

func (e *EverestServer) checkStorageAccessByUpdate(ctx context.Context, storageID string, params UpdateBackupStorageParams) (*model.BackupStorage, error) {
	s, err := e.storage.GetBackupStorage(ctx, storageID)
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
	ctx context.Context, backupStorageID string, params *UpdateBackupStorageParams,
	newAccessKeyID, newSecretKeyID *string,
) (*model.BackupStorage, int, error) {
	updated, err := e.storage.UpdateBackupStorage(ctx, model.UpdateBackupStorageParams{
		ID:          backupStorageID,
		Name:        params.Name,
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
				return nil, http.StatusBadRequest, errors.New("Backup storage with the same name already exists. " + pgErr.Detail)
			}
		}

		e.l.Error(err)
		return nil, http.StatusInternalServerError, errors.New("Could not update backup storage")
	}

	return updated, 0, nil
}
