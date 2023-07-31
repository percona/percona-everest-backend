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

// Package client provides a way to communicate with a k8s cluster.
package client

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // load all auth plugins
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/percona/percona-everest-backend/pkg/kubernetes/client/database"
	"github.com/percona/percona-everest-backend/pkg/kubernetes/client/objectstorage"
)

const (
	defaultQPSLimit   = 100
	defaultBurstLimit = 150
)

// Client is the internal client for Kubernetes.
type Client struct {
	clientset           kubernetes.Interface
	dbClusterClient     *database.DBClusterClient
	objectStorageClient *objectstorage.Client
	restConfig          *rest.Config
	namespace           string
	clusterName         string
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
	dbClusterClient, err := database.NewForConfig(c.restConfig)
	if err != nil {
		return err
	}
	c.dbClusterClient = dbClusterClient
	_, err = c.GetServerVersion()
	if err != nil {
		return err
	}

	objectStorageClient, err := objectstorage.NewForConfig(c.restConfig)
	if err != nil {
		return err
	}
	c.objectStorageClient = objectStorageClient

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
	return c.dbClusterClient.DBClusters(c.namespace).List(ctx, metav1.ListOptions{})
}

// GetDatabaseCluster returns database clusters by provided name.
func (c *Client) GetDatabaseCluster(ctx context.Context, name string) (*everestv1alpha1.DatabaseCluster, error) {
	cluster, err := c.dbClusterClient.DBClusters(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

// CreateObjectStorage creates an objectStorage.
func (c *Client) CreateObjectStorage(ctx context.Context, storage *everestv1alpha1.ObjectStorage) error {
	return c.objectStorageClient.ObjectStorage(storage.Namespace).Post(ctx, storage, metav1.CreateOptions{})
}

// CreateSecret creates k8s Secret.
func (c *Client) CreateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	return c.clientset.CoreV1().Secrets(secret.Namespace).Create(ctx, secret, metav1.CreateOptions{})
}
