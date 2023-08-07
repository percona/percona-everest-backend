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

package kubernetes

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/percona/percona-everest-backend/pkg/kubernetes/client"
)

// ListDatabaseClusters returns list of managed database clusters.
func (k *Kubernetes) ListDatabaseClusters(ctx context.Context) (*everestv1alpha1.DatabaseClusterList, error) {
	list := &everestv1alpha1.DatabaseClusterList{}
	err := k.client.ListResources(ctx, client.DBClusterAPIKind, list, &metav1.ListOptions{})

	return list, err
}

// GetDatabaseCluster returns database clusters by provided name.
func (k *Kubernetes) GetDatabaseCluster(ctx context.Context, name string) (*everestv1alpha1.DatabaseCluster, error) {
	c := &everestv1alpha1.DatabaseCluster{}
	err := k.client.GetResource(ctx, client.DBClusterAPIKind, name, c, &metav1.GetOptions{})
	return c, err
}

// CreateDatabaseCluster creates a database cluster.
func (k *Kubernetes) CreateDatabaseCluster(ctx context.Context, cluster *everestv1alpha1.DatabaseCluster) (*everestv1alpha1.DatabaseCluster, error) {
	c := &everestv1alpha1.DatabaseCluster{}
	err := k.client.CreateResource(ctx, client.DBClusterAPIKind, cluster, c, &metav1.CreateOptions{})
	return c, err
}

// UpdateDatabaseCluster updates a database cluster by its name.
func (k *Kubernetes) UpdateDatabaseCluster(ctx context.Context, name string, cluster *everestv1alpha1.DatabaseCluster) (*everestv1alpha1.DatabaseCluster, error) {
	if cluster.ResourceVersion == "" {
		c, err := k.GetDatabaseCluster(ctx, name)
		if err != nil {
			return nil, err
		}

		cluster.ResourceVersion = c.ResourceVersion
	}

	c := &everestv1alpha1.DatabaseCluster{}
	err := k.client.UpdateResource(ctx, client.DBClusterAPIKind, name, cluster, c, &metav1.UpdateOptions{})
	return c, err
}

// DeleteDatabaseCluster deletes a database cluster.
func (k *Kubernetes) DeleteDatabaseCluster(ctx context.Context, name string) error {
	return k.client.DeleteResource(ctx, client.DBClusterAPIKind, name, &metav1.DeleteOptions{})
}
