// Copyright (C) 2023 Percona LLC
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

// Package objectstorage provides a way to operate on ObjectStorage objects in k8s.
package objectstorage

import (
	"context"
	"sync"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const (
	apiKind = "objectstorages"
)

// ClientInterface supports getting an ObjectStorage client.
type ClientInterface interface {
	ObjectStorage(namespace string) Interface
}

// Client contains a rest client.
type Client struct {
	restClient rest.Interface
}

//nolint:gochecknoglobals
var addToScheme sync.Once

// NewForConfig creates a new database cluster client based on config.
func NewForConfig(c *rest.Config) (*Client, error) {
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

	return &Client{restClient: client}, nil
}

// ObjectStorage returns a db cluster client.
func (c *Client) ObjectStorage(namespace string) Interface { //nolint:ireturn
	return &client{
		restClient: c.restClient,
		namespace:  namespace,
	}
}

// Interface supports list, get and watch methods.
type Interface interface {
	Post(ctx context.Context, storage *everestv1alpha1.ObjectStorage, opts metav1.CreateOptions) (*everestv1alpha1.ObjectStorage, error)
	Update(ctx context.Context, storage *everestv1alpha1.ObjectStorage, pt types.PatchType, opts metav1.UpdateOptions) (*everestv1alpha1.ObjectStorage, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}

type client struct {
	restClient rest.Interface
	namespace  string
}

// Post creates a resource.
func (c *client) Post(
	ctx context.Context,
	storage *everestv1alpha1.ObjectStorage,
	opts metav1.CreateOptions,
) (*everestv1alpha1.ObjectStorage, error) {
	result := &everestv1alpha1.ObjectStorage{}
	err := c.restClient.
		Post().
		Namespace(c.namespace).
		Resource(apiKind).Body(storage).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).Into(result)
	return result, err
}

// Update creates a resource.
func (c *client) Update(
	ctx context.Context,
	storage *everestv1alpha1.ObjectStorage,
	pt types.PatchType,
	opts metav1.UpdateOptions,
) (*everestv1alpha1.ObjectStorage, error) {
	result := &everestv1alpha1.ObjectStorage{}
	err := c.restClient.
		Patch(pt).
		Namespace(c.namespace).
		Resource(apiKind).Body(storage).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).Into(result)
	return result, err
}

// Delete creates a resource.
func (c *client) Delete(
	ctx context.Context,
	name string,
	opts metav1.DeleteOptions,
) error {
	return c.restClient.
		Delete().Param("name", name).
		Namespace(c.namespace).
		Resource(apiKind).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).Error()
}
