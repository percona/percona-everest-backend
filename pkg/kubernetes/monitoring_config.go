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

// ListMonitoringConfigs returns list of managed monitoring configs.
func (k *Kubernetes) ListMonitoringConfigs(ctx context.Context) (*everestv1alpha1.MonitoringConfigList, error) {
	return k.client.ListMonitoringConfigs(ctx)
}

// GetMonitoringConfig returns monitoring configs by provided name.
func (k *Kubernetes) GetMonitoringConfig(ctx context.Context, name string) (*everestv1alpha1.MonitoringConfig, error) {
	return k.client.GetMonitoringConfig(ctx, name)
}

// CreateMonitoringConfig returns monitoring configs by provided name.
func (k *Kubernetes) CreateMonitoringConfig(ctx context.Context, storage *everestv1alpha1.MonitoringConfig) error {
	return k.client.CreateMonitoringConfig(ctx, storage)
}

// UpdateMonitoringConfig returns monitoring configs by provided name.
func (k *Kubernetes) UpdateMonitoringConfig(ctx context.Context, storage *everestv1alpha1.MonitoringConfig) error {
	return k.client.UpdateMonitoringConfig(ctx, storage)
}

// DeleteMonitoringConfig returns monitoring configs by provided name.
func (k *Kubernetes) DeleteMonitoringConfig(ctx context.Context, name string) error {
	return k.client.DeleteMonitoringConfig(ctx, name)
}
