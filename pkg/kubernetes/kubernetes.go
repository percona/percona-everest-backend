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
	"errors"
	"strings"

	"go.uber.org/zap"
	"k8s.io/client-go/rest"

	"github.com/percona/percona-everest-backend/pkg/kubernetes/client"
)

type (
	// ClusterType defines type of cluster.
	ClusterType string
)

const (
	// ClusterTypeUnknown is for unknown type.
	ClusterTypeUnknown ClusterType = "unknown"
	// ClusterTypeMinikube is for minikube.
	ClusterTypeMinikube ClusterType = "minikube"
	// ClusterTypeEKS is for EKS.
	ClusterTypeEKS ClusterType = "eks"
	// ClusterTypeGeneric is a generic type.
	ClusterTypeGeneric ClusterType = "generic"

	configMapName = "everest-configuration"
)

// Kubernetes is a client for Kubernetes.
type Kubernetes struct {
	client    client.KubeClientConnector
	l         *zap.SugaredLogger
	namespace string
}

// NewInCluster creates a new kubernetes client using incluster authentication.
func NewInCluster(l *zap.SugaredLogger) (*Kubernetes, error) {
	client, err := client.NewInCluster()
	if err != nil {
		return nil, err
	}
	return &Kubernetes{
		client:    client,
		l:         l,
		namespace: client.Namespace(),
	}, nil
}

// Config returns rest config.
func (k *Kubernetes) Config() *rest.Config {
	return k.client.Config()
}

// Namespace returns the current namespace.
func (k *Kubernetes) Namespace() string {
	return k.namespace
}

// ClusterName returns the name of the k8s cluster.
func (k *Kubernetes) ClusterName() string {
	return k.client.ClusterName()
}

// GetEverestID returns the ID of the namespace where everest is deployed.
func (k *Kubernetes) GetEverestID(ctx context.Context) (string, error) {
	namespace, err := k.client.GetNamespace(ctx, k.namespace)
	if err != nil {
		return "", err
	}
	return string(namespace.UID), nil
}

// GetClusterType tries to guess the underlying kubernetes cluster based on storage class.
func (k *Kubernetes) GetClusterType(ctx context.Context) (ClusterType, error) {
	storageClasses, err := k.client.GetStorageClasses(ctx)
	if err != nil {
		return ClusterTypeUnknown, err
	}
	for _, storageClass := range storageClasses.Items {
		if strings.Contains(storageClass.Provisioner, "aws") {
			return ClusterTypeEKS, nil
		}
		if strings.Contains(storageClass.Provisioner, "minikube") ||
			strings.Contains(storageClass.Provisioner, "kubevirt.io/hostpath-provisioner") ||
			strings.Contains(storageClass.Provisioner, "standard") {
			return ClusterTypeMinikube, nil
		}
	}
	return ClusterTypeGeneric, nil
}

// GetPersistedNamespaces returns list of persisted namespaces.
func (k *Kubernetes) GetPersistedNamespaces(ctx context.Context, namespace string) ([]string, error) {
	var namespaces []string
	cMap, err := k.client.GetConfigMap(ctx, namespace, configMapName)
	if err != nil {
		return namespaces, err
	}
	// FIXME: If we decide to separate the installation and the namespaces this key can be empty/nonexistent
	v, ok := cMap.Data["namespaces"]
	if !ok {
		return namespaces, errors.New("`namespaces` key does not exist in the configmap")
	}
	namespaces = strings.Split(v, ",")
	return namespaces, nil
}
