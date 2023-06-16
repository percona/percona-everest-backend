package api

import (
	"log"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

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
		result = append(result, BackupStorage{
			Id:     bs.ID,
			Name:   bs.Name,
			Region: bs.Region,
			Url:    bs.URL,
		})
	}

	return ctx.JSON(http.StatusOK, result)
}

// CreateBackupStorage Create a new backup storage object.
// rollbacks are implemented without transactions bc the secrets storage is going to be moved out of pg.
func (e *EverestServer) CreateBackupStorage(ctx echo.Context) error {
	params, err := validateCreateBackupStorageRequest(ctx)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	c := ctx.Request().Context()

	accessKeyID := uuid.NewString()
	err = e.SecretsStorage.CreateSecret(c, accessKeyID, params.AccessKey)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	secretKeyID := uuid.NewString()
	err = e.SecretsStorage.CreateSecret(c, secretKeyID, params.SecretKey)
	if err != nil {
		log.Println(err)
		// rollback the created accessKey secret
		_, dError := e.SecretsStorage.DeleteSecret(c, accessKeyID)
		if dError != nil {
			log.Printf("Inconsistent DB state, manual intervention required. Can not delete the secret with id = %s", accessKeyID)
		}

		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	s, err := e.Storage.CreateBackupStorage(c, model.CreateBackupStorageParams{
		Name:        params.Name,
		BucketName:  params.BucketName,
		URL:         params.Url,
		Region:      params.Region,
		AccessKeyID: accessKeyID,
		SecretKeyID: secretKeyID,
	})
	if err != nil {
		log.Println(err)

		// rollback the chagnes - delete secrets
		_, dError := e.SecretsStorage.DeleteSecret(c, accessKeyID)
		if dError != nil {
			log.Printf("Inconsistent DB state, manual intervention required. Can not delete the secret with id = %s", accessKeyID)
		}

		_, dError = e.SecretsStorage.DeleteSecret(c, secretKeyID)
		if dError != nil {
			log.Printf("Inconsistent DB state, manual intervention required. Can not delete the secret with id = %s", secretKeyID)
		}

		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	result := BackupStorage{
		Id:     s.ID,
		Name:   s.Name,
		Region: s.Region,
		Url:    s.URL,
	}

	return ctx.JSON(http.StatusOK, result)
}

// DeleteBackupStorage Delete the specified backup storage.
func (e *EverestServer) DeleteBackupStorage(ctx echo.Context, backupStorageID string) error {
	c := ctx.Request().Context()
	bs, err := e.Storage.GetBackupStorage(c, backupStorageID)
	if err != nil {
		log.Println(err)
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
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	result := BackupStorage{
		Id:     s.ID,
		Name:   s.Name,
		Region: s.Region,
		Url:    s.URL,
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

	s, err := e.Storage.GetBackupStorage(c, backupStorageID)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	var newAccessKeyID, newSecretKeyID *string
	var oldAccessKey, oldSecretKey *string

	if params.AccessKey != nil {
		newID := uuid.NewString()
		newAccessKeyID = &newID
		oldAccessKey, err = e.SecretsStorage.ReplaceSecret(c, s.AccessKeyID, *newAccessKeyID, *params.AccessKey)
		if err != nil {
			log.Println(err)
			return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
		}
	}

	if params.SecretKey != nil { //nolint:nestif
		newID := uuid.NewString()
		newSecretKeyID = &newID
		oldSecretKey, err = e.SecretsStorage.ReplaceSecret(c, s.SecretKeyID, *newSecretKeyID, *params.SecretKey)
		if err != nil {
			// rollback the accessKey to the old value
			if params.AccessKey != nil {
				_, err = e.SecretsStorage.ReplaceSecret(c, *newAccessKeyID, s.AccessKeyID, *oldAccessKey)
				if err != nil {
					log.Printf("Inconsistent DB state, manual intervention required. Can not revert changes over the secret with id = %s", s.AccessKeyID)
				}
			}

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
	if err != nil { //nolint:nestif
		log.Println(err)

		// rollback accessKey to the old values
		if params.AccessKey != nil {
			_, rErr := e.SecretsStorage.ReplaceSecret(c, *newAccessKeyID, s.AccessKeyID, *oldAccessKey)
			if rErr != nil {
				log.Printf("Inconsistent DB state, manual intervention required. Can not revert changes over the secret with id = %s", s.AccessKeyID)
			}
		}
		// rollback secretKey to the old values
		if params.SecretKey != nil {
			_, rErr := e.SecretsStorage.ReplaceSecret(c, *newSecretKeyID, s.SecretKeyID, *oldSecretKey)
			if rErr != nil {
				log.Printf("Inconsistent DB state, manual intervention required. Can not revert changes over the secret with id = %s", s.SecretKeyID)
			}
		}

		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	result := BackupStorage{
		Id:     updated.ID,
		Name:   updated.Name,
		Region: updated.Region,
		Url:    updated.URL,
	}

	return ctx.JSON(http.StatusOK, result)
}
