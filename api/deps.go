package api

import (
	"context"

	"github.com/percona/percona-everest-backend/model"
)

type secretsStorage interface {
	CreateSecret(ctx context.Context, id, value string) error
	GetSecret(ctx context.Context, id string) (string, error)
	UpdateSecret(ctx context.Context, id, value string) error
	DeleteSecret(ctx context.Context, id string) (string, error)
}

type storage interface {
	backupStorageStorage
	kubernetesCluster
	pmmInstanceStorage
}

type kubernetesCluster interface {
	CreateKubernetesCluster(ctx context.Context, params model.CreateKubernetesClusterParams) (*model.KubernetesCluster, error)
	ListKubernetesClusters(ctx context.Context) ([]model.KubernetesCluster, error)
	GetKubernetesCluster(ctx context.Context, id string) (*model.KubernetesCluster, error)
	DeleteKubernetesCluster(ctx context.Context, id string) error
}

type backupStorageStorage interface {
	CreateBackupStorage(ctx context.Context, params model.CreateBackupStorageParams) (*model.BackupStorage, error)
	ListBackupStorages(ctx context.Context) ([]model.BackupStorage, error)
	GetBackupStorage(ctx context.Context, id string) (*model.BackupStorage, error)
	UpdateBackupStorage(ctx context.Context, params model.UpdateBackupStorageParams) (*model.BackupStorage, error)
	DeleteBackupStorage(ctx context.Context, id string) error
}

type pmmInstanceStorage interface {
	CreatePMMInstance(pmm *model.PMMInstance) (*model.PMMInstance, error)
	ListPMMInstances() ([]model.PMMInstance, error)
	GetPMMInstance(ID string) (*model.PMMInstance, error)
	DeletePMMInstance(ID string) error
	UpdatePMMInstance(ID string, params model.UpdatePMMInstanceParams) error
}
