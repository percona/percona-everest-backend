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

	"github.com/percona/percona-everest-backend/model"
)

const pgErrUniqueViolation = "unique_violation"

type secretsStorage interface {
	CreateSecret(ctx context.Context, id, value string) error
	GetSecret(ctx context.Context, id string) (string, error)
	UpdateSecret(ctx context.Context, id, value string) error
	DeleteSecret(ctx context.Context, id string) (string, error)
}

type storage interface {
	backupStorageStorage
	kubernetesClusterStorage
	pmmInstanceStorage
}

type kubernetesClusterStorage interface {
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
