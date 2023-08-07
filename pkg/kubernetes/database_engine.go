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

// ListDatabaseEngines returns list of managed database clusters.
func (k *Kubernetes) ListDatabaseEngines(ctx context.Context) (*everestv1alpha1.DatabaseEngineList, error) {
	list := &everestv1alpha1.DatabaseEngineList{}
	err := k.client.ListResources(ctx, client.DBEngineAPIKind, list, &metav1.ListOptions{})

	return list, err
}

// GetDatabaseEngine returns database clusters by provided name.
func (k *Kubernetes) GetDatabaseEngine(ctx context.Context, name string) (*everestv1alpha1.DatabaseEngine, error) {
	c := &everestv1alpha1.DatabaseEngine{}
	err := k.client.GetResource(ctx, client.DBEngineAPIKind, name, c, &metav1.GetOptions{})
	return c, err
}

// UpdateDatabaseEngine updates a database cluster by its name.
func (k *Kubernetes) UpdateDatabaseEngine(ctx context.Context, name string, engine *everestv1alpha1.DatabaseEngine) (*everestv1alpha1.DatabaseEngine, error) {
	if engine.ResourceVersion == "" {
		c, err := k.GetDatabaseEngine(ctx, name)
		if err != nil {
			return nil, err
		}

		engine.ResourceVersion = c.ResourceVersion
	}

	c := &everestv1alpha1.DatabaseEngine{}
	err := k.client.UpdateResource(ctx, client.DBEngineAPIKind, name, engine, c, &metav1.UpdateOptions{})
	return c, err
}
