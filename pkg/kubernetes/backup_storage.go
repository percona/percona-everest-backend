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

// Package kubernetes ...
//
//nolint:dupl
package kubernetes

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
)

// ListBackupStorages returns list of managed database clusters.
func (k *Kubernetes) ListBackupStorages(ctx context.Context) (*everestv1alpha1.BackupStorageList, error) {
	return k.client.ListBackupStorages(ctx)
}

// GetBackupStorage returns database clusters by provided name.
func (k *Kubernetes) GetBackupStorage(ctx context.Context, name string) (*everestv1alpha1.BackupStorage, error) {
	return k.client.GetBackupStorage(ctx, name)
}

// CreateBackupStorage returns database clusters by provided name.
func (k *Kubernetes) CreateBackupStorage(ctx context.Context, storage *everestv1alpha1.BackupStorage) error {
	return k.client.CreateBackupStorage(ctx, storage)
}

// UpdateBackupStorage returns database clusters by provided name.
func (k *Kubernetes) UpdateBackupStorage(ctx context.Context, storage *everestv1alpha1.BackupStorage) error {
	return k.client.UpdateBackupStorage(ctx, storage)
}

// DeleteBackupStorage returns database clusters by provided name.
func (k *Kubernetes) DeleteBackupStorage(ctx context.Context, name string) error {
	return k.client.DeleteBackupStorage(ctx, name)
}
