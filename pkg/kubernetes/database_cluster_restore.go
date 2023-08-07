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
package kubernetes

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/percona/percona-everest-backend/pkg/kubernetes/client"
)

// ListDatabaseClusterRestores returns a list of database cluster restores.
func (k *Kubernetes) ListDatabaseClusterRestores(ctx context.Context) (*everestv1alpha1.DatabaseClusterRestoreList, error) {
	list := &everestv1alpha1.DatabaseClusterRestoreList{}
	err := k.client.ListResources(ctx, client.DBClusterRestoreAPIKind, list, &metav1.ListOptions{})

	return list, err
}

// GetDatabaseClusterRestore returns a database cluster restore by provided name.
func (k *Kubernetes) GetDatabaseClusterRestore(ctx context.Context, name string) (*everestv1alpha1.DatabaseClusterRestore, error) {
	c := &everestv1alpha1.DatabaseClusterRestore{}
	err := k.client.GetResource(ctx, client.DBClusterRestoreAPIKind, name, c, &metav1.GetOptions{})
	return c, err
}

// CreateDatabaseClusterRestore creates a database cluster restore.
func (k *Kubernetes) CreateDatabaseClusterRestore(ctx context.Context, cluster *everestv1alpha1.DatabaseClusterRestore) (*everestv1alpha1.DatabaseClusterRestore, error) {
	c := &everestv1alpha1.DatabaseClusterRestore{}
	err := k.client.CreateResource(ctx, client.DBClusterRestoreAPIKind, cluster, c, &metav1.CreateOptions{})
	return c, err
}

// UpdateDatabaseClusterRestore updates a database cluster restore by its name.
func (k *Kubernetes) UpdateDatabaseClusterRestore(ctx context.Context, name string, cluster *everestv1alpha1.DatabaseClusterRestore) (*everestv1alpha1.DatabaseClusterRestore, error) {
	if cluster.ResourceVersion == "" {
		c, err := k.GetDatabaseClusterRestore(ctx, name)
		if err != nil {
			return nil, err
		}

		cluster.ResourceVersion = c.ResourceVersion
	}

	c := &everestv1alpha1.DatabaseClusterRestore{}
	err := k.client.UpdateResource(ctx, client.DBClusterRestoreAPIKind, name, cluster, c, &metav1.UpdateOptions{})
	return c, err
}

// DeleteDatabaseClusterRestore deletes a database cluster restore.
func (k *Kubernetes) DeleteDatabaseClusterRestore(ctx context.Context, name string) error {
	return k.client.DeleteResource(ctx, client.DBClusterRestoreAPIKind, name, &metav1.DeleteOptions{})
}
