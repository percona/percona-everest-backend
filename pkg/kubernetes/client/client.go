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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // load all auth plugins
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/percona/percona-everest-backend/pkg/kubernetes/client/customresouces"
)

const (
	defaultQPSLimit   = 100
	defaultBurstLimit = 150

	defaultAPIURIPath  = "/api"
	defaultAPIsURIPath = "/apis"
)

// Client is the internal client for Kubernetes.
type Client struct {
	clientset       kubernetes.Interface
	customClientSet *customresouces.Client
	restConfig      *rest.Config
	restMapper      meta.RESTMapper
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
	groupResources, err := restmapper.GetAPIGroupResources(c.clientset.Discovery())
	if err != nil {
		return err
	}
	c.restMapper = restmapper.NewDiscoveryRESTMapper(groupResources)

	customClient, err := customresouces.NewForConfig(c.restConfig, c.restMapper)
	if err != nil {
		return err
	}
	c.customClientSet = customClient
	_, err = c.GetServerVersion()
	if err != nil {
		return err
	}

	return nil
}

// ClusterName returns the name of the k8s cluster.
func (c *Client) ClusterName() string {
	return c.clusterName
}

// GetServerVersion returns server version.
func (c *Client) GetServerVersion() (*version.Info, error) {
	return c.clientset.Discovery().ServerVersion()
}

// ApplyObject applies object.
func (c *Client) ApplyObject(obj runtime.Object) error {
	gvk := obj.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := c.restMapper.RESTMapping(gk, gvk.Version)
	if err != nil {
		return err
	}
	namespace, name, err := c.retrieveMetaFromObject(obj)
	if err != nil {
		return err
	}
	cli, err := c.resourceClient(mapping.GroupVersionKind.GroupVersion())
	if err != nil {
		return err
	}
	helper := resource.NewHelper(cli, mapping)
	return c.applyObject(helper, namespace, name, obj)
}

func (c *Client) applyObject(helper *resource.Helper, namespace, name string, obj runtime.Object) error {
	if _, err := helper.Get(namespace, name); err != nil {
		_, err = helper.Create(namespace, false, obj)
		if err != nil {
			return err
		}
	} else {
		_, err = helper.Replace(namespace, name, true, obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) retrieveMetaFromObject(obj runtime.Object) (string, string, error) {
	name, err := meta.NewAccessor().Name(obj)
	if err != nil {
		return "", name, err
	}
	namespace, err := meta.NewAccessor().Namespace(obj)
	if err != nil {
		return namespace, name, err
	}
	if namespace == "" {
		namespace = c.namespace
	}
	return namespace, name, nil
}

func (c *Client) resourceClient( //nolint:ireturn
	gv schema.GroupVersion,
) (rest.Interface, error) {
	cfg := c.restConfig
	cfg.ContentConfig = resource.UnstructuredPlusDefaultContentConfig()
	cfg.GroupVersion = &gv
	if len(gv.Group) == 0 {
		cfg.APIPath = defaultAPIURIPath
	} else {
		cfg.APIPath = defaultAPIsURIPath
	}
	return rest.RESTClientFor(cfg)
}

// DeleteObject deletes object from the k8s cluster.
func (c *Client) DeleteObject(obj runtime.Object) error {
	gvk := obj.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := c.restMapper.RESTMapping(gk, gvk.Version)
	if err != nil {
		return err
	}
	namespace, name, err := c.retrieveMetaFromObject(obj)
	if err != nil {
		return err
	}
	cli, err := c.resourceClient(mapping.GroupVersionKind.GroupVersion())
	if err != nil {
		return err
	}
	helper := resource.NewHelper(cli, mapping)
	err = deleteObject(helper, namespace, name)
	return err
}

func deleteObject(helper *resource.Helper, namespace, name string) error {
	if _, err := helper.Get(namespace, name); err == nil {
		_, err = helper.Delete(namespace, name)
		if err != nil {
			return err
		}
	}
	return nil
}

// ListObjects lists objects by provided group, version, kind.
func (c *Client) ListObjects(gvk schema.GroupVersionKind, into runtime.Object) error {
	helper, err := c.helperForGVK(gvk)
	if err != nil {
		return errors.Wrap(err, "could not create helper")
	}

	l, err := helper.List(c.namespace, gvk.Version, &metav1.ListOptions{})
	if err != nil {
		return err
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(l)
	if err != nil {
		return err
	}

	return runtime.DefaultUnstructuredConverter.FromUnstructured(u, into)
}

// GetObject retrieves an object by provided group, version, kind and name.
func (c *Client) GetObject(gvk schema.GroupVersionKind, name string, into runtime.Object) error {
	helper, err := c.helperForGVK(gvk)
	if err != nil {
		return errors.Wrap(err, "could not create helper")
	}

	l, err := helper.Get(c.namespace, name)
	if err != nil {
		return errors.Wrap(err, "failed to get object using helper")
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(l)
	if err != nil {
		return errors.Wrap(err, "failed to convert object to unstructured")
	}

	return runtime.DefaultUnstructuredConverter.FromUnstructured(u, into)
}

func (c *Client) helperForGVK(gvk schema.GroupVersionKind) (*resource.Helper, error) {
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := c.restMapper.RESTMapping(gk, gvk.Version)
	if err != nil {
		return nil, err
	}
	cli, err := c.resourceClient(mapping.GroupVersionKind.GroupVersion())
	if err != nil {
		return nil, err
	}

	return resource.NewHelper(cli, mapping), nil
}
