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

	"github.com/percona/percona-everest-backend/model"
)

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
