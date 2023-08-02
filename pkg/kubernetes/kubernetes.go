// Copyright (C) 2017 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Package kubernetes provides functionality for kubernetes.
package kubernetes

import (
	"context"
	"encoding/base64"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"github.com/percona/percona-everest-backend/pkg/kubernetes/client"
)

// Kubernetes is a client for Kubernetes.
type Kubernetes struct {
	client     client.KubeClientConnector
	l          *zap.SugaredLogger
	kubeconfig []byte
}

type secretGetter interface {
	GetSecret(ctx context.Context, id string) (string, error)
}

// New returns new Kubernetes object.
func New(kubeconfig []byte, namespace string, l *zap.SugaredLogger) (*Kubernetes, error) {
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
	kubernetesID string, namespace string, l *zap.SugaredLogger,
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
