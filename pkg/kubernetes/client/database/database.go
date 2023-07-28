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

// Package database provides a way to operate on database cluster objects in k8s.
package database

import (
	"context"
	"sync"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const (
	// DBClusterKind defines kind for DB cluster.
	DBClusterKind = "DatabaseCluster"
	apiKind       = "databaseclusters"
)

// DBClusterClientInterface supports getting a database cluster client.
type DBClusterClientInterface interface {
	DBClusters(namespace string) DBClusterInterface
}

// DBClusterClient contains a rest client.
type DBClusterClient struct {
	restClient rest.Interface
}

//nolint:gochecknoglobals
var addToScheme sync.Once

// NewForConfig creates a new database cluster client based on config.
func NewForConfig(c *rest.Config) (*DBClusterClient, error) {
	config := *c
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

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &DBClusterClient{restClient: client}, nil
}

// DBClusters returns a db cluster client.
func (c *DBClusterClient) DBClusters(namespace string) DBClusterInterface { //nolint:ireturn
	return &dbClusterClient{
		restClient: c.restClient,
		namespace:  namespace,
	}
}

// DBClusterInterface supports list, get and watch methods.
type DBClusterInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*everestv1alpha1.DatabaseClusterList, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*everestv1alpha1.DatabaseCluster, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type dbClusterClient struct {
	restClient rest.Interface
	namespace  string
}

// List lists database clusters based on opts.
func (c *dbClusterClient) List(ctx context.Context, opts metav1.ListOptions) (*everestv1alpha1.DatabaseClusterList, error) {
	result := &everestv1alpha1.DatabaseClusterList{}
	err := c.restClient.
		Get().
		Namespace(c.namespace).
		Resource(apiKind).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}

// Get retrieves database cluster based on opts.
func (c *dbClusterClient) Get(
	ctx context.Context,
	name string,
	opts metav1.GetOptions,
) (*everestv1alpha1.DatabaseCluster, error) {
	result := &everestv1alpha1.DatabaseCluster{}
	err := c.restClient.
		Get().
		Namespace(c.namespace).
		Resource(apiKind).
		VersionedParams(&opts, scheme.ParameterCodec).
		Name(name).
		Do(ctx).
		Into(result)
	return result, err
}

// Watch starts a watch based on opts.
func (c *dbClusterClient) Watch( //nolint:ireturn
	ctx context.Context,
	opts metav1.ListOptions,
) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.namespace).
		Resource(apiKind).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}
