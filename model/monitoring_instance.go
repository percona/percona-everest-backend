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

// Package model ...
package model

import (
	"context"
	"fmt"
	"time"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// MonitoringInstanceType defines type of monitoring used by an instance.
type MonitoringInstanceType string

// PMMMonitoringInstanceType refers to PMM as a monitoring type.
const PMMMonitoringInstanceType = "pmm"

// MonitoringInstance represents a monitoring instance.
type MonitoringInstance struct {
	Type MonitoringInstanceType
	Name string `gorm:"primary_key"`
	URL  string
	// ID of API key in secret storage
	APIKeySecretID string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// SecretName returns the name of the k8s secret as referenced by the k8s MonitoringConfig resource.
func (m *MonitoringInstance) SecretName() string {
	return fmt.Sprintf("%s-secret", m.Name)
}

// Secrets returns all monitoring instance secrets from secrets storage.
func (m *MonitoringInstance) Secrets(ctx context.Context, getSecret func(ctx context.Context, id string) (string, error)) (map[string]string, error) {
	apiKey, err := getSecret(ctx, m.APIKeySecretID)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"apiKey": apiKey,
	}, nil
}

// K8sResource returns a resource which shall be created when storing this struct in Kubernetes.
func (m *MonitoringInstance) K8sResource( //nolint:ireturn
	namespace string,
) (runtime.Object, error) {
	mc := &everestv1alpha1.MonitoringConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: namespace,
		},
		Spec: everestv1alpha1.MonitoringConfigSpec{
			Type:                  everestv1alpha1.MonitoringType(m.Type),
			CredentialsSecretName: m.SecretName(),
		},
	}

	switch m.Type {
	case PMMMonitoringInstanceType:
		mc.Spec.PMM = everestv1alpha1.PMMConfig{
			URL:   m.URL,
			Image: "percona/pmm-client:2",
		}
	default:
		return nil, fmt.Errorf("monitoring instance type %s not supported", m.Type)
	}

	return mc, nil
}
