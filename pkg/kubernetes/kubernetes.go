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

// Package kubernetes provides functionality for kubernetes.
package kubernetes

import (
	"context"
	"encoding/base64"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/percona/percona-everest-backend/pkg/kubernetes/client"
)

// Kubernetes is a client for Kubernetes.
type Kubernetes struct {
	client     client.KubeClientConnector
	l          logr.Logger
	kubeconfig []byte
}

type secretGetter interface {
	GetSecret(ctx context.Context, id string) (string, error)
}

// New returns new Kubernetes object.
func New(kubeconfig []byte, namespace string, l logr.Logger) (*Kubernetes, error) {
	client, err := client.NewFromKubeConfig(kubeconfig, namespace)
	if err != nil {
		return nil, err
	}

	return &Kubernetes{
		client:     client,
		l:          l,
		kubeconfig: kubeconfig,
	}, nil
}

// NewFromSecretsStorage returns a new Kubernetes object by retrieving the kubeconfig from a
// secrets storage.
func NewFromSecretsStorage(
	ctx context.Context, secretGetter secretGetter,
	kubernetesID string, namespace string, l logr.Logger,
) (*Kubernetes, error) {
	kubeconfigBase64, err := secretGetter.GetSecret(ctx, kubernetesID)
	if err != nil {
		return nil, errors.Wrap(err, "could not get kubeconfig from secrets storage")
	}
	kubeconfig, err := base64.StdEncoding.DecodeString(kubeconfigBase64)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode base64 kubeconfig")
	}

	return New(kubeconfig, namespace, l)
}

// ClusterName returns the name of the k8s cluster.
func (k *Kubernetes) ClusterName() string {
	return k.client.ClusterName()
}

// GetSecret returns secret by name.
func (k *Kubernetes) GetSecret(ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	return k.client.GetSecret(ctx, name, namespace)
}
