package api

import (
	"github.com/labstack/echo/v4"

	"github.com/percona/percona-everest-backend/model"
)

type secretsStorage interface {
	CreateSecret(ctx echo.Context, id, value string) error
	GetSecret(ctx echo.Context, id string) (string, error)
}

type storage interface {
	CreateKubernetesCluster(ctx echo.Context, params model.CreateKubernetesClusterParams) (*model.KubernetesCluster, error)
	ListKubernetesClusters(ctx echo.Context) ([]model.KubernetesCluster, error)
	GetKubernetesCluster(ctx echo.Context, id string) (*model.KubernetesCluster, error)

	CreateBackupStorage(ctx echo.Context, params model.CreateBackupStorageParams) (*model.BackupStorage, error)
	ListBackupStorages(ctx echo.Context) ([]model.BackupStorage, error)
	GetBackupStorage(ctx echo.Context, id string) (*model.BackupStorage, error)
	UpdateBackupStorage(ctx echo.Context, params model.UpdateBackupStorageParams) (*model.BackupStorage, error)
}
