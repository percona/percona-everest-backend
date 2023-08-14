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
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/percona/percona-everest-backend/model"
)

// ErrMonitoringConfigInUse is returned when a monitoring config is in use.
var ErrMonitoringConfigInUse error = errors.New("monitoring config is in use")

// EnsureMonitoringConfigExists makes sure a monitoring config for the provided monitoring instance
// exists in Kubernetes. If it does not, it is created.
func (k *Kubernetes) EnsureMonitoringConfigExists(ctx context.Context, mi *model.MonitoringInstance, secrets secretGetter) error {
	_, err := k.client.GetMonitoringConfig(ctx, mi.Name)
	if err == nil {
		return nil
	}

	if !k8serrors.IsNotFound(err) {
		return errors.Wrap(err, "could not get monitoring config from kubernetes")
	}

	miSecrets, err := mi.Secrets(ctx, secrets)
	if err != nil {
		return errors.Wrap(err, "could not get monitoring instance secrets from secrets storage")
	}

	if err = k.CreateMonitoringConfig(ctx, mi, miSecrets); err != nil {
		return errors.Wrap(err, "could not create monitoring config")
	}

	return nil
}

// CreateMonitoringConfig creates a MonitoringConfig.
func (k *Kubernetes) CreateMonitoringConfig(ctx context.Context, mi *model.MonitoringInstance, secretData map[string]string) error {
	return k.createConfigWithSecret(ctx, mi.Name, k.namespace, secretData, func(secretName, namespace string) error {
		mc := &everestv1alpha1.MonitoringConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      mi.Name,
				Namespace: namespace,
			},
			Spec: everestv1alpha1.MonitoringConfigSpec{
				Type:                  everestv1alpha1.MonitoringType(mi.Type),
				CredentialsSecretName: secretName,
			},
		}

		switch mi.Type {
		case model.PMMMonitoringInstanceType:
			mc.Spec.PMM = everestv1alpha1.PMMConfig{
				URL:   mi.URL,
				Image: "percona/pmm-client:latest",
			}
		default:
			return errors.Errorf("monitoring instance type %s not supported", mi.Type)
		}

		return k.client.CreateMonitoringConfig(ctx, mc)
	})
}

// DeleteMonitoringConfig deletes a MonitoringConfig.
func (k *Kubernetes) DeleteMonitoringConfig(ctx context.Context, mi *model.MonitoringInstance) error {
	used, err := k.isMonitoringConfigInUse(ctx, mi.Name)
	if err != nil {
		return errors.Wrap(err, "could not check if monitoring config is in use")
	}
	if used {
		return ErrMonitoringConfigInUse
	}

	if err := k.client.DeleteMonitoringConfig(ctx, mi.Name); err != nil {
		return errors.Wrap(err, "could not delete monitoring config")
	}

	return k.DeleteSecret(ctx, mi.SecretName(), k.namespace)
}

// GetMonitoringConfigsBySecretName returns a list of monitoring configs which use
// the provided secret name.
func (k *Kubernetes) GetMonitoringConfigsBySecretName(
	ctx context.Context, secretName string,
) ([]*everestv1alpha1.MonitoringConfig, error) {
	mcs, err := k.client.ListMonitoringConfigs(ctx)
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

func (k *Kubernetes) isMonitoringConfigInUse(ctx context.Context, name string) (bool, error) {
	vmAgents, err := k.ListVMAgents()
	if err != nil {
		return false, errors.Wrap(err, "could not list VM agents in Kubernetes")
	}
	secretNames := make([]string, 0, len(vmAgents.Items))

	for _, vm := range vmAgents.Items {
		vm := vm
		secretNames = append(secretNames, k.SecretNamesFromVMAgent(&vm)...)
	}

	for _, secretName := range secretNames {
		mcs, err := k.GetMonitoringConfigsBySecretName(ctx, secretName)
		if err != nil {
			return false, err
		}

		for _, mc := range mcs {
			if mc.Name == name {
				return true, nil
			}
		}
	}

	dbs, err := k.ListDatabaseClusters(ctx)
	if err != nil {
		return false, errors.Wrap(err, "could not list database clusters in Kubernetes")
	}

	for _, db := range dbs.Items {
		if db.Spec.Monitoring.MonitoringConfigName == name {
			return true, nil
		}
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
			return []string{}
		}

		secretName, ok, err := unstructured.NestedString(rw, "basicAuth", "password", "name")
		if err != nil {
			// We can ignore the error because it has to be a string.
			k.l.Debug(err)
			return []string{}
		}
		if !ok {
			return []string{}
		}
		res = append(res, secretName)
	}

	return res
}
