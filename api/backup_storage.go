package api

import (
	"context"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/percona/percona-everest-backend/model"
)

// ListBackupStorages List of the created backup storages.
func (e *EverestServer) ListBackupStorages(ctx echo.Context) error {
	list, err := e.storage.ListBackupStorages(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	result := make([]BackupStorage, 0, len(list))
	for _, bs := range list {
		s := bs
		result = append(result, BackupStorage{
			Type:        BackupStorageType(bs.Type),
			Name:        s.Name,
			Description: s.Description,
			BucketName:  s.BucketName,
			Region:      s.Region,
			Url:         &s.URL,
		})
	}

	return ctx.JSON(http.StatusOK, result)
}

// CreateBackupStorage Create a new backup storage object.
// rollbacks are implemented without transactions bc the secrets storage is going to be moved out of pg.
func (e *EverestServer) CreateBackupStorage(ctx echo.Context) error { //nolint:funlen,cyclop
	params, err := validateCreateBackupStorageRequest(ctx)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	c := ctx.Request().Context()
	var accessKeyID, secretKeyID string

	existingStorage, err := e.storage.GetBackupStorage(c, params.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	if existingStorage != nil {
		err = errors.Errorf("Storage %s already exists", params.Name)
		e.l.Error(err)
		return ctx.JSON(http.StatusConflict, Error{Message: pointer.ToString(err.Error())})
	}

	defer func() {
		if err == nil {
			return
		}

		// rollback the changes - delete secrets
		if accessKeyID != "" {
			_, dError := e.secretsStorage.DeleteSecret(c, accessKeyID)
			if dError != nil {
				e.l.Errorf("Failed to delete unused secret with id = %s", accessKeyID)
			}
		}

		if secretKeyID != "" {
			_, dError := e.secretsStorage.DeleteSecret(c, secretKeyID)
			if dError != nil {
				e.l.Errorf("Failed to delete unused secret with id = %s", secretKeyID)
			}
		}
	}()

	accessKeyID = uuid.NewString()
	err = e.secretsStorage.CreateSecret(c, accessKeyID, params.AccessKey)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	secretKeyID = uuid.NewString()
	err = e.secretsStorage.CreateSecret(c, secretKeyID, params.SecretKey)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	var url string
	if params.Url != nil {
		url = *params.Url
	}

	s, err := e.storage.CreateBackupStorage(c, model.CreateBackupStorageParams{
		Name:        params.Name,
		Description: params.Description,
		Type:        string(params.Type),
		BucketName:  params.BucketName,
		URL:         url,
		Region:      params.Region,
		AccessKeyID: accessKeyID,
		SecretKeyID: secretKeyID,
	})
	if err != nil {
		e.l.Error(err)
		// TODO do not throw DB errors to API, e.g. duplicated key handling
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	result := BackupStorage{
		Type:        BackupStorageType(s.Type),
		Name:        s.Name,
		Description: s.Description,
		BucketName:  s.BucketName,
		Region:      s.Region,
		Url:         &s.URL,
	}

	k8sID, err := e.currentKubernetesID(c)
	if err != nil {
		err = errors.Wrap(err, "Failed to create a backup storage")
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	err = e.everestK8s.ApplyObjectStorage(
		ctx,
		k8sID,
		result,
		map[string]string{
			secretKeyID: params.SecretKey,
			accessKeyID: params.AccessKey,
		},
	)
	if err != nil {
		err = errors.Wrap(err, "Failed to create a backup storage in the current k8s cluster")
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, result)
}

// DeleteBackupStorage Delete the specified backup storage.
func (e *EverestServer) DeleteBackupStorage(ctx echo.Context, backupStorageID string) error { //nolint:cyclop
	c := ctx.Request().Context()
	bs, err := e.storage.GetBackupStorage(c, backupStorageID)
	if err != nil {
		e.l.Error(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString(err.Error())})
		}
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	deletedAccessKey, err := e.secretsStorage.DeleteSecret(c, bs.AccessKeyID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	deletedSecretKey, err := e.secretsStorage.DeleteSecret(c, bs.SecretKeyID)
	if err != nil {
		e.l.Error(err)

		// rollback the changes - put the deleted secret back
		cErr := e.secretsStorage.CreateSecret(c, bs.SecretKeyID, deletedAccessKey)
		if cErr != nil {
			e.l.Errorf("Inconsistent DB state, manual intervention required. Can not revert changes over the secret with id = %s", bs.AccessKeyID)
		}
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
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

		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	k8sID, err := e.currentKubernetesID(c)
	if err != nil {
		err = errors.Wrap(err, "Failed to create a backup storage")
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	err = e.everestK8s.RemoveObjectStorage(ctx, k8sID, bs.Name)
	if err != nil {
		err = errors.Wrap(err, "Failed to create a backup storage in the current k8s cluster")
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetBackupStorage Get the specified backup storage.
func (e *EverestServer) GetBackupStorage(ctx echo.Context, backupStorageID string) error {
	s, err := e.storage.GetBackupStorage(ctx.Request().Context(), backupStorageID)
	if err != nil {
		e.l.Error(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString(err.Error())})
		}
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	result := BackupStorage{
		Description: s.Description,
		Type:        BackupStorageType(s.Type),
		BucketName:  s.BucketName,
		Name:        s.Name,
		Region:      s.Region,
		Url:         &s.URL,
	}

	return ctx.JSON(http.StatusOK, result)
}

// UpdateBackupStorage update of the specified backup storage.
func (e *EverestServer) UpdateBackupStorage(ctx echo.Context, name string) error { //nolint:funlen,cyclop
	params, err := validateUpdateBackupStorageRequest(ctx)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	c := ctx.Request().Context()

	// check data access
	s, err := e.checkStorageAccessByUpdate(c, name, *params)
	if err != nil {
		e.l.Error(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString(err.Error())})
		}
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	var newAccessKeyID, newSecretKeyID *string
	defer func() {
		if err == nil {
			return
		}

		// if an error appeared - cleanup the created secrets
		if newAccessKeyID != nil {
			_, err = e.secretsStorage.DeleteSecret(c, *newAccessKeyID)
			if err != nil {
				e.l.Errorf("Failed to delete unused secret, please delete it manually. id = %s", *newAccessKeyID)
			}
		}

		if newSecretKeyID != nil {
			_, err = e.secretsStorage.DeleteSecret(c, *newSecretKeyID)
			if err != nil {
				e.l.Errorf("Failed to delete unused secret, please delete it manually. id = %s", *newSecretKeyID)
			}
		}
	}()

	if params.AccessKey != nil {
		newID := uuid.NewString()
		newAccessKeyID = &newID

		// create new AccessKey
		err = e.secretsStorage.CreateSecret(c, newID, *params.AccessKey)
		if err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
		}
	}

	if params.SecretKey != nil {
		newID := uuid.NewString()
		newSecretKeyID = &newID

		// create new SecretKey
		err = e.secretsStorage.CreateSecret(c, newID, *params.SecretKey)
		if err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
		}
	}

	updated, err := e.storage.UpdateBackupStorage(c, model.UpdateBackupStorageParams{
		Name:        name,
		Description: params.Description,
		BucketName:  params.BucketName,
		URL:         params.Url,
		Region:      params.Region,
		AccessKeyID: newAccessKeyID,
		SecretKeyID: newSecretKeyID,
	})
	if err != nil {
		e.l.Errorf("Failed to update backup storage with name = %s", name)
		// TODO: do not throw DB errors to API, e.g. duplicated key handling
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	// delete old AccessKey
	if params.AccessKey != nil {
		_, cErr := e.secretsStorage.DeleteSecret(c, s.AccessKeyID)
		if cErr != nil {
			e.l.Errorf("Failed to delete unused secret, please delete it manually. id = %s", s.AccessKeyID)
		}
	}

	// delete old SecretKey
	if params.SecretKey != nil {
		_, cErr := e.secretsStorage.DeleteSecret(c, s.SecretKeyID)
		if cErr != nil {
			e.l.Errorf("Failed to delete unused secret, please delete it manually. id = %s", s.SecretKeyID)
		}
	}

	result := BackupStorage{
		Type:        BackupStorageType(updated.Type),
		Name:        updated.Name,
		Description: updated.Description,
		BucketName:  updated.BucketName,
		Region:      updated.Region,
		Url:         &updated.URL,
	}

	return ctx.JSON(http.StatusOK, result)
}

func (e *EverestServer) checkStorageAccessByUpdate(ctx context.Context, storageName string, params UpdateBackupStorageParams) (*model.BackupStorage, error) {
	s, err := e.storage.GetBackupStorage(ctx, storageName)
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

func (e *EverestServer) currentKubernetesID(ctx context.Context) (string, error) {
	clusters, err := e.storage.ListKubernetesClusters(ctx)
	if err != nil {
		return "", err
	}

	if len(clusters) == 0 {
		return "", errors.Errorf("No k8s cluster registred")
	}

	// The first one is the current one
	return clusters[0].ID, nil
}
