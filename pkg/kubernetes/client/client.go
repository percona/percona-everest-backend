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

// Package client provides a way to communicate with a k8s cluster.
package client

import (
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // load all auth plugins
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/percona/percona-everest-backend/pkg/kubernetes/client/custom"
)

const (
	defaultQPSLimit   = 100
	defaultBurstLimit = 150
)

// Client is the internal client for Kubernetes.
type Client struct {
	clientset       kubernetes.Interface
	customClientSet *custom.Client
	restConfig      *rest.Config
	namespace       string
	clusterName     string
}

// NewFromKubeConfig returns new Client from a kubeconfig.
func NewFromKubeConfig(kubeconfig []byte, namespace string) (*Client, error) {
	clientConfig, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	config.QPS = defaultQPSLimit
	config.Burst = defaultBurstLimit
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	c := &Client{
		clientset:   clientset,
		restConfig:  config,
		clusterName: clientConfig.Contexts[clientConfig.CurrentContext].Cluster,
		namespace:   namespace,
	}

	err = c.initOperatorClients()
	return c, err
}

// Initializes clients for operators.
func (c *Client) initOperatorClients() error {
	customClient, err := custom.NewForConfig(c.restConfig)
	if err != nil {
		return err
	}
	c.customClientSet = customClient
	_, err = c.GetServerVersion()
	if err != nil {
		return err
	}

	return err
}

// ClusterName returns the name of the k8s cluster.
func (c *Client) ClusterName() string {
	return c.clusterName
}

// GetServerVersion returns server version.
func (c *Client) GetServerVersion() (*version.Info, error) {
	return c.clientset.Discovery().ServerVersion()
}

// ListDatabaseClusters returns list of managed database clusters.
func (c *Client) ListDatabaseClusters(ctx context.Context) (*everestv1alpha1.DatabaseClusterList, error) {
	return c.customClientSet.DBClusters(c.namespace).List(ctx, metav1.ListOptions{})
}

// GetDatabaseCluster returns database clusters by provided name.
func (c *Client) GetDatabaseCluster(ctx context.Context, name string) (*everestv1alpha1.DatabaseCluster, error) {
	return c.customClientSet.DBClusters(c.namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetSecret returns secret by name.
func (c *Client) GetSecret(ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	return c.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
}

// CreateObjectStorage creates an objectStorage.
func (c *Client) CreateObjectStorage(ctx context.Context, storage *everestv1alpha1.ObjectStorage) error {
	_, err := c.customClientSet.ObjectStorage(storage.Namespace).Post(ctx, storage, metav1.CreateOptions{})
	return err
}

// GetObjectStorage returns the objectStorage.
func (c *Client) GetObjectStorage(ctx context.Context, name, namespace string) (*everestv1alpha1.ObjectStorage, error) {
	return c.customClientSet.ObjectStorage(namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeleteObjectStorage deletes the objectStorage.
func (c *Client) DeleteObjectStorage(ctx context.Context, name, namespace string) error {
	return c.customClientSet.ObjectStorage(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// CreateSecret creates k8s Secret.
func (c *Client) CreateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	return c.clientset.CoreV1().Secrets(secret.Namespace).Create(ctx, secret, metav1.CreateOptions{})
}

// DeleteSecret deletes the k8s Secret.
func (c *Client) DeleteSecret(ctx context.Context, name, namespace string) error {
	return c.clientset.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
