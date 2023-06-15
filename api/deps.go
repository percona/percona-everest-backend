package api

import (
	"context"

	"github.com/percona/percona-everest-backend/model"
)

type secretsStorage interface {
	CreateSecret(ctx context.Context, id, value string) error
	GetSecret(ctx context.Context, id string) (string, error)
	UpdateSecret(ctx context.Context, id, value string) error
}

type storage interface {
	CreateKubernetesCluster(ctx context.Context, params model.CreateKubernetesClusterParams) (*model.KubernetesCluster, error)
	ListKubernetesClusters(ctx context.Context) ([]model.KubernetesCluster, error)
	GetKubernetesCluster(ctx context.Context, id string) (*model.KubernetesCluster, error)

	CreateBackupStorage(ctx context.Context, params model.CreateBackupStorageParams) (*model.BackupStorage, error)
	ListBackupStorages(ctx context.Context) ([]model.BackupStorage, error)
	GetBackupStorage(ctx context.Context, id string) (*model.BackupStorage, error)
	UpdateBackupStorage(ctx context.Context, params model.UpdateBackupStorageParams) (*model.BackupStorage, error)
	DeleteBackupStorage(ctx context.Context, id string) (*model.BackupStorage, error)
}
