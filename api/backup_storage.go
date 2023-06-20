package api

import (
	"context"
	"log"
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
	list, err := e.Storage.ListBackupStorages(ctx.Request().Context())
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
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

// CreateBackupStorage Create a new backup storage object.
// rollbacks are implemented without transactions bc the secrets storage is going to be moved out of pg.
func (e *EverestServer) CreateBackupStorage(ctx echo.Context) error { //nolint:funlen,cyclop
	params, err := validateCreateBackupStorageRequest(ctx)
	if err != nil {
		log.Println(err)
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
			_, dError := e.SecretsStorage.DeleteSecret(c, accessKeyID)
			if dError != nil {
				log.Printf("Failed to delete unused secret with id = %s", accessKeyID)
			}
		}

		if secretKeyID != "" {
			_, dError := e.SecretsStorage.DeleteSecret(c, secretKeyID)
			if dError != nil {
				log.Printf("Failed to delete unused secret with id = %s", secretKeyID)
			}
		}
	}()

	accessKeyID = uuid.NewString()
	err = e.SecretsStorage.CreateSecret(c, accessKeyID, params.AccessKey)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	secretKeyID = uuid.NewString()
	err = e.SecretsStorage.CreateSecret(c, secretKeyID, params.SecretKey)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	var url string
	if params.Url != nil {
		url = *params.Url
	}

	s, err := e.Storage.CreateBackupStorage(c, model.CreateBackupStorageParams{
		Name:        params.Name,
		Type:        string(params.Type),
		BucketName:  params.BucketName,
		URL:         url,
		Region:      params.Region,
		AccessKeyID: accessKeyID,
		SecretKeyID: secretKeyID,
	})
	if err != nil {
		log.Println(err)
		// TODO do not throw DB errors to API, e.g. duplicated key handling
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
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

// DeleteBackupStorage Delete the specified backup storage.
func (e *EverestServer) DeleteBackupStorage(ctx echo.Context, backupStorageID string) error {
	c := ctx.Request().Context()
	bs, err := e.Storage.GetBackupStorage(c, backupStorageID)
	if err != nil {
		log.Println(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString(err.Error())})
		}
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	deletedAccessKey, err := e.SecretsStorage.DeleteSecret(c, bs.AccessKeyID)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	deletedSecretKey, err := e.SecretsStorage.DeleteSecret(c, bs.SecretKeyID)
	if err != nil {
		log.Println(err)

		// rollback the changes - put the deleted secret back
		cErr := e.SecretsStorage.CreateSecret(c, bs.SecretKeyID, deletedAccessKey)
		if cErr != nil {
			log.Printf("Inconsistent DB state, manual intervention required. Can not revert changes over the secret with id = %s", bs.AccessKeyID)
		}
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	err = e.Storage.DeleteBackupStorage(c, backupStorageID)
	if err != nil {
		log.Println(err)

		// rollback the changes - put the deleted secrets back
		cErr := e.SecretsStorage.CreateSecret(c, bs.AccessKeyID, deletedAccessKey)
		if cErr != nil {
			log.Printf("Inconsistent DB state, manual intervention required. Can not revert changes over the secret with id = %s", bs.AccessKeyID)
		}
		cErr = e.SecretsStorage.CreateSecret(c, bs.SecretKeyID, deletedSecretKey)
		if cErr != nil {
			log.Printf("Inconsistent DB state, manual intervention required. Can not revert changes over the secret with id = %s", bs.SecretKeyID)
		}

		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.NoContent(http.StatusNoContent)
}

// GetBackupStorage Get the specified backup storage.
func (e *EverestServer) GetBackupStorage(ctx echo.Context, backupStorageID string) error {
	s, err := e.Storage.GetBackupStorage(ctx.Request().Context(), backupStorageID)
	if err != nil {
		log.Println(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusNotFound, Error{Message: pointer.ToString(err.Error())})
		}
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
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

// UpdateBackupStorage update of the specified backup storage.
func (e *EverestServer) UpdateBackupStorage(ctx echo.Context, backupStorageID string) error { //nolint:funlen,cyclop
	params, err := validateUpdateBackupStorageRequest(ctx)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	c := ctx.Request().Context()

	// check data access
	s, err := e.checkStorageAccessByUpdate(c, backupStorageID, *params)
	if err != nil {
		log.Println(err)
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
			_, err = e.SecretsStorage.DeleteSecret(c, *newAccessKeyID)
			if err != nil {
				log.Printf("Failed to delete unused secret, please delete it manually. id = %s", *newAccessKeyID)
			}
		}

		if newSecretKeyID != nil {
			_, err = e.SecretsStorage.DeleteSecret(c, *newSecretKeyID)
			if err != nil {
				log.Printf("Failed to delete unused secret, please delete it manually. id = %s", *newSecretKeyID)
			}
		}
	}()

	if params.AccessKey != nil {
		newID := uuid.NewString()
		newAccessKeyID = &newID

		// create new AccessKey
		err = e.SecretsStorage.CreateSecret(c, newID, *params.AccessKey)
		if err != nil {
			log.Println(err)
			return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
		}
	}

	if params.SecretKey != nil {
		newID := uuid.NewString()
		newSecretKeyID = &newID

		// create new SecretKey
		err = e.SecretsStorage.CreateSecret(c, newID, *params.SecretKey)
		if err != nil {
			log.Println(err)
			return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
		}
	}

	updated, err := e.Storage.UpdateBackupStorage(c, model.UpdateBackupStorageParams{
		ID:          backupStorageID,
		Name:        params.Name,
		BucketName:  params.BucketName,
		URL:         params.Url,
		Region:      params.Region,
		AccessKeyID: newAccessKeyID,
		SecretKeyID: newSecretKeyID,
	})
	if err != nil {
		log.Printf("Failed to update backup storage with id = %s", backupStorageID)
		// TODO: do not throw DB errors to API, e.g. duplicated key handling
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	// delete old AccessKey
	if params.AccessKey != nil {
		_, cErr := e.SecretsStorage.DeleteSecret(c, s.AccessKeyID)
		if cErr != nil {
			log.Printf("Failed to delete unused secret, please delete it manually. id = %s", s.AccessKeyID)
		}
	}

	// delete old SecretKey
	if params.SecretKey != nil {
		_, cErr := e.SecretsStorage.DeleteSecret(c, s.SecretKeyID)
		if cErr != nil {
			log.Printf("Failed to delete unused secret, please delete it manually. id = %s", s.SecretKeyID)
		}
	}

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

func (e *EverestServer) checkStorageAccessByUpdate(ctx context.Context, storageID string, params UpdateBackupStorageParams) (*model.BackupStorage, error) {
	s, err := e.Storage.GetBackupStorage(ctx, storageID)
	if err != nil {
		return nil, err
	}

	accessKey, err := e.SecretsStorage.GetSecret(ctx, s.AccessKeyID)
	if err != nil {
		return nil, err
	}

	secretKey, err := e.SecretsStorage.GetSecret(ctx, s.SecretKeyID)
	if err != nil {
		return nil, err
	}

	oldData := &storageData{
		accessKey: accessKey,
		secretKey: secretKey,
		s:         *s,
	}

	err = validateStorageAccessByUpdate(oldData, params)
	if err != nil {
		return nil, err
	}

	return &oldData.s, nil
}
