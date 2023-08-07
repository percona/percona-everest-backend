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
	"sync"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // load all auth plugins
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// APIKind represents resource kind in kubernetes.
type APIKind string

const (
	defaultQPSLimit   = 100
	defaultBurstLimit = 150

	// DBClusterAPIKind represents a database cluster.
	DBClusterAPIKind APIKind = "databaseclusters"
	// DBClusterRestoreAPIKind represents a database cluster restore.
	DBClusterRestoreAPIKind APIKind = "databaseclusterrestores"
	// DBEngineAPIKind represents a database engine.
	DBEngineAPIKind APIKind = "databaseengines"
)

// Client is the internal client for Kubernetes.
type Client struct {
	clientset     kubernetes.Interface
	everestClient *rest.RESTClient
	restConfig    *rest.Config
	namespace     string
	clusterName   string
}

//nolint:gochecknoglobals
var addToScheme sync.Once

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
	cl, err := c.everestRESTClient(c.restConfig)
	if err != nil {
		return err
	}
	c.everestClient = cl
	_, err = c.GetServerVersion()
	return err
}

func (c *Client) everestRESTClient(cfg *rest.Config) (*rest.RESTClient, error) {
	config := *cfg
	config.ContentConfig.GroupVersion = &everestv1alpha1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	var err error
	addToScheme.Do(func() {
		err = everestv1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme)
		metav1.AddToGroupVersion(scheme.Scheme, everestv1alpha1.GroupVersion)
	})

	if err != nil {
		return nil, err
	}

	return rest.RESTClientFor(&config)
}

// ClusterName returns the name of the k8s cluster.
func (c *Client) ClusterName() string {
	return c.clusterName
}

// GetServerVersion returns server version.
func (c *Client) GetServerVersion() (*version.Info, error) {
	return c.clientset.Discovery().ServerVersion()
}
