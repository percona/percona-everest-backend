package api

import (
	"context"
	"log"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/percona/percona-everest-backend/model"
)

// ListBackupStorages List of the created backup storages.
func (e *EverestServer) ListBackupStorages(ctx echo.Context) error {
	list, err := e.Storage.ListBackupStorages(ctx.Request().Context())
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, list)
}

// CreateBackupStorage Create a new backup storage object.
func (e *EverestServer) CreateBackupStorage(ctx echo.Context) error {
	var params CreateBackupStorageParams
	if err := ctx.Bind(&params); err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	c := ctx.Request().Context()

	s, err := e.Storage.CreateBackupStorage(c, model.CreateBackupStorageParams{
		Name:       params.Name,
		BucketName: params.BucketName,
		URL:        params.Url,
		Region:     params.Region,
	})
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	// create secrets
	err = e.createBackupStorageSecrets(c, s.ID, params.AccessKey, params.SecretKey)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, s)
}

// DeleteBackupStorage Delete the specified backup storage.
func (e *EverestServer) DeleteBackupStorage(ctx echo.Context, backupStorageID string) error {
	s, err := e.Storage.DeleteBackupStorage(ctx.Request().Context(), backupStorageID)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, s)
}

// GetBackupStorage Get the specified backup storage.
func (e *EverestServer) GetBackupStorage(ctx echo.Context, backupStorageID string) error {
	s, err := e.Storage.GetBackupStorage(ctx.Request().Context(), backupStorageID)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, s)
}

// UpdateBackupStorage update of the specified backup storage.
func (e *EverestServer) UpdateBackupStorage(ctx echo.Context, backupStorageID string) error {
	var params UpdateBackupStorageParams
	if err := ctx.Bind(&params); err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	c := ctx.Request().Context()

	s, err := e.Storage.UpdateBackupStorage(c, model.UpdateBackupStorageParams{
		ID:         backupStorageID,
		Name:       params.Name,
		BucketName: params.BucketName,
		URL:        params.Url,
		Region:     params.Region,
	})
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	// update secrets
	err = e.updateBackupStorageSecrets(c, backupStorageID, params.AccessKey, params.SecretKey)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, s)
}

func (e *EverestServer) createBackupStorageSecrets(ctx context.Context, storageID, accessKey, secretKey string) error {
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return e.SecretsStorage.CreateSecret(gCtx, accessKeyPrefix(storageID), accessKey)
	})

	g.Go(func() error {
		return e.SecretsStorage.CreateSecret(gCtx, secretKeyPrefix(storageID), secretKey)
	})

	return g.Wait()
}

func (e *EverestServer) updateBackupStorageSecrets(ctx context.Context, storageID string, accessKey, secretKey *string) error {
	g, gCtx := errgroup.WithContext(ctx)

	if accessKey != nil {
		g.Go(func() error {
			return e.SecretsStorage.UpdateSecret(gCtx, accessKeyPrefix(storageID), *accessKey)
		})
	}

	if secretKey != nil {
		g.Go(func() error {
			return e.SecretsStorage.UpdateSecret(gCtx, secretKeyPrefix(storageID), *secretKey)
		})
	}

	return g.Wait()
}

func accessKeyPrefix(id string) string {
	return "access-key-" + id
}

func secretKeyPrefix(id string) string {
	return "secret-key-" + id
}
