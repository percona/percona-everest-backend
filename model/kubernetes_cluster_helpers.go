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
package model

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CreateKubernetesClusterParams parameters for KubernetesCluster record creation.
type CreateKubernetesClusterParams struct {
	Name      string
	Namespace *string
}

// KubernetesCluster represents db model for KubernetesCluster.
type KubernetesCluster struct {
	ID        string
	Name      string
	Namespace string

	CreatedAt time.Time
	UpdatedAt time.Time
}

const defaultK8sNamespace = "percona-everest"

// CreateKubernetesCluster creates a KubernetesCluster record.
func (db *Database) CreateKubernetesCluster(_ context.Context, params CreateKubernetesClusterParams) (*KubernetesCluster, error) {
	namespace := defaultK8sNamespace
	if params.Namespace != nil {
		namespace = *params.Namespace
	}

	k := &KubernetesCluster{
		ID:        uuid.NewString(),
		Name:      params.Name,
		Namespace: namespace,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	err := db.gormDB.Create(k).Error
	if err != nil {
		return nil, err
	}

	return k, nil
}

// ListKubernetesClusters returns all available KubernetesCluster records.
func (db *Database) ListKubernetesClusters(_ context.Context) ([]KubernetesCluster, error) {
	var clusters []KubernetesCluster
	err := db.gormDB.Find(&clusters).Error
	if err != nil {
		return nil, err
	}
	return clusters, nil
}

// GetKubernetesCluster returns KubernetesCluster record by its ID.
func (db *Database) GetKubernetesCluster(_ context.Context, id string) (*KubernetesCluster, error) {
	cluster := &KubernetesCluster{
		ID: id,
	}
	err := db.gormDB.First(cluster).Error
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

// DeleteKubernetesCluster deletes a Kubernetes cluster by its ID.
func (db *Database) DeleteKubernetesCluster(_ context.Context, id string) error {
	return db.gormDB.Delete(&KubernetesCluster{ID: id}).Error
}
