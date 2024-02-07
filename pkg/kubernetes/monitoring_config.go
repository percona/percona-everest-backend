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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	monitoringConfigNameLabel = "monitoringConfigName"
)

// ListMonitoringConfigs returns list of managed monitoring configs.
func (k *Kubernetes) ListMonitoringConfigs(ctx context.Context, namespace string) (*everestv1alpha1.MonitoringConfigList, error) {
	return k.client.ListMonitoringConfigs(ctx, namespace)
}

// GetMonitoringConfig returns monitoring configs by provided name.
func (k *Kubernetes) GetMonitoringConfig(ctx context.Context, namespace, name string) (*everestv1alpha1.MonitoringConfig, error) {
	return k.client.GetMonitoringConfig(ctx, namespace, name)
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
func (k *Kubernetes) DeleteMonitoringConfig(ctx context.Context, namespace, name string) error {
	return k.client.DeleteMonitoringConfig(ctx, namespace, name)
}

// IsMonitoringConfigUsed checks that a backup storage by provided name is used across k8s cluster.
func (k *Kubernetes) IsMonitoringConfigUsed(ctx context.Context, namespace, monitoringConfigName string) (bool, error) {
	_, err := k.client.GetMonitoringConfig(ctx, namespace, monitoringConfigName)
	if err != nil {
		return false, err
	}
	options := metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{
				monitoringConfigNameLabel: monitoringConfigName,
			},
		}),
	}
	list, err := k.client.ListDatabaseClusters(ctx, options)
	if err != nil {
		return false, err
	}
	if len(list.Items) > 0 {
		return true, nil
	}
	return false, nil
}

// SecretNamesFromVMAgent returns a list of secret names as used by VMAgent's remoteWrite password field.
func (k *Kubernetes) SecretNamesFromVMAgent(vmAgent *unstructured.Unstructured) []string {
	rws, ok, err := unstructured.NestedSlice(vmAgent.Object, "spec", "remoteWrite")
	if err != nil {
		// We can ignore the error because it has to be an interface.
		k.l.Debug(err)
	}
	if !ok {
		return []string{}
	}

	res := make([]string, 0, len(rws))
	for _, rw := range rws {
		rw, ok := rw.(map[string]interface{})
		if !ok {
			continue
		}

		secretName, ok, err := unstructured.NestedString(rw, "basicAuth", "password", "name")
		if err != nil {
			// We can ignore the error because it has to be a string.
			k.l.Debug(err)
			continue
		}
		if !ok {
			continue
		}
		res = append(res, secretName)
	}

	return res
}

// GetMonitoringConfigsBySecretName returns a list of monitoring configs which use
// the provided secret name.
func (k *Kubernetes) GetMonitoringConfigsBySecretName(
	ctx context.Context, namespace, secretName string,
) ([]*everestv1alpha1.MonitoringConfig, error) {
	mcs, err := k.client.ListMonitoringConfigs(ctx, namespace)
	if err != nil {
		return nil, err
	}

	res := make([]*everestv1alpha1.MonitoringConfig, 0, 1)
	for _, mc := range mcs.Items {
		mc := mc
		if mc.Spec.CredentialsSecretName == secretName {
			res = append(res, &mc)
		}
	}

	return res, nil
}
