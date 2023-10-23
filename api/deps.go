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

	"gorm.io/gorm"

	"github.com/percona/percona-everest-backend/model"
)

const pgErrUniqueViolation = "unique_violation"

type secretsStorage interface {
	PutSecret(ctx context.Context, id, value string) error
	GetSecret(ctx context.Context, id string) (string, error)
	DeleteSecret(ctx context.Context, id string) error
}

type storage interface {
	backupStorageStorage
	kubernetesClusterStorage
	monitoringInstanceStorage
	settingsStorage

	Close() error
	Transaction(fn func(tx *gorm.DB) error) error
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
	GetBackupStorage(ctx context.Context, tx *gorm.DB, name string) (*model.BackupStorage, error)
	UpdateBackupStorage(ctx context.Context, tx *gorm.DB, params model.UpdateBackupStorageParams) error
	DeleteBackupStorage(ctx context.Context, name string, tx *gorm.DB) error
}

type monitoringInstanceStorage interface {
	CreateMonitoringInstance(pmm *model.MonitoringInstance) (*model.MonitoringInstance, error)
	ListMonitoringInstances() ([]model.MonitoringInstance, error)
	GetMonitoringInstance(name string) (*model.MonitoringInstance, error)
	DeleteMonitoringInstance(name string, tx *gorm.DB) error
	UpdateMonitoringInstance(name string, params model.UpdateMonitoringInstanceParams) error
}

type settingsStorage interface {
	GetEverestID(ctx context.Context) (string, error)
	GetSettingByKey(ctx context.Context, key string) (string, error)
	InitSettings(ctx context.Context) error
}
