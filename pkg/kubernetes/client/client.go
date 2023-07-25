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
	"bytes"
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yamlSerializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // load all auth plugins
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/percona/percona-everest-backend/pkg/kubernetes/client/database"
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
	dbClusterClient *database.DBClusterClient
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
	dbClusterClient, err := database.NewForConfig(c.restConfig)
	if err != nil {
		return err
	}
	c.dbClusterClient = dbClusterClient
	_, err = c.GetServerVersion()
	return err
}

func (c *Client) kubeClient() (client.Client, error) { //nolint:ireturn
	rcl, err := rest.HTTPClientFor(c.restConfig)
	if err != nil {
		return nil, err
	}

	rm, err := apiutil.NewDynamicRESTMapper(c.restConfig, rcl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dynamic rest mapper")
	}

	cl, err := client.New(c.restConfig, client.Options{
		Scheme: scheme.Scheme,
		Mapper: rm,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}
	return cl, nil
}

// ClusterName returns the name of the k8s cluster.
func (c *Client) ClusterName() string {
	return c.clusterName
}

// GetServerVersion returns server version.
func (c *Client) GetServerVersion() (*version.Info, error) {
	return c.clientset.Discovery().ServerVersion()
}

// ListDatabaseClusters returns list of managed PCX clusters.
func (c *Client) ListDatabaseClusters(ctx context.Context) (*everestv1alpha1.DatabaseClusterList, error) {
	return c.dbClusterClient.DBClusters(c.namespace).List(ctx, metav1.ListOptions{})
}

// GetDatabaseCluster returns PXC clusters by provided name.
func (c *Client) GetDatabaseCluster(ctx context.Context, name string) (*everestv1alpha1.DatabaseCluster, error) {
	cluster, err := c.dbClusterClient.DBClusters(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

// ApplyObject applies object.
func (c *Client) ApplyObject(obj runtime.Object) error {
	groupResources, err := restmapper.GetAPIGroupResources(c.clientset.Discovery())
	if err != nil {
		return err
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	gvk := obj.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := mapper.RESTMapping(gk, gvk.Version)
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

// ApplyFile accepts manifest file contents, parses into []runtime.Object
// and applies them against the cluster.
func (c *Client) ApplyFile(fileBytes []byte) error {
	objs, err := c.getObjects(fileBytes)
	if err != nil {
		return err
	}
	for i := range objs {
		err := c.ApplyObject(objs[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) getObjects(f []byte) ([]runtime.Object, error) {
	objs := []runtime.Object{}
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(f), 100)
	var err error
	for {
		var rawObj runtime.RawExtension
		if err = decoder.Decode(&rawObj); err != nil {
			break
		}

		obj, _, err := yamlSerializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return nil, err
		}

		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return nil, err
		}

		objs = append(objs, &unstructured.Unstructured{Object: unstructuredMap})
	}

	return objs, nil //nolint:nilerr
}

// DeleteFile accepts manifest file contents parses into []runtime.Object
// and deletes them from the cluster.
func (c *Client) DeleteFile(fileBytes []byte) error {
	objs, err := c.getObjects(fileBytes)
	if err != nil {
		return err
	}
	for i := range objs {
		err := c.DeleteObject(objs[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteObject deletes object from the k8s cluster.
func (c *Client) DeleteObject(obj runtime.Object) error {
	groupResources, err := restmapper.GetAPIGroupResources(c.clientset.Discovery())
	if err != nil {
		return err
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	gvk := obj.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := mapper.RESTMapping(gk, gvk.Version)
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

// DeleteAllMonitoringResources deletes all resources related to monitoring from k8s cluster.
func (c *Client) DeleteAllMonitoringResources(ctx context.Context) error {
	cl, err := c.kubeClient()
	if err != nil {
		return err
	}

	opts := []client.DeleteAllOfOption{
		client.MatchingLabels{"everest.percona.com/type": "monitoring"},
		client.InNamespace(c.namespace),
	}

	for _, o := range c.monitoringResourceTypesForRemoval() {
		if err := cl.DeleteAllOf(ctx, o, opts...); err != nil {
			return err
		}
	}

	return nil
}

// monitoringResourceTypesForRemoval returns a list of object types in k8s cluster to be removed
// when deleting all monitoring resources from a k8s cluster.
func (c *Client) monitoringResourceTypesForRemoval() []client.Object {
	vmNodeScrape := &unstructured.Unstructured{}
	vmNodeScrape.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "operator.victoriametrics.com",
		Kind:    "VMNodeScrape",
		Version: "v1beta1",
	})

	vmPodScrape := &unstructured.Unstructured{}
	vmPodScrape.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "operator.victoriametrics.com",
		Kind:    "VMPodScrape",
		Version: "v1beta1",
	})

	vmAgent := &unstructured.Unstructured{}
	vmAgent.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "operator.victoriametrics.com",
		Kind:    "VMAgent",
		Version: "v1beta1",
	})

	vmServiceScrape := &unstructured.Unstructured{}
	vmServiceScrape.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "operator.victoriametrics.com",
		Kind:    "VMServiceScrape",
		Version: "v1beta1",
	})

	return []client.Object{
		&corev1.ServiceAccount{},
		&corev1.Service{},
		&appsv1.Deployment{},

		vmNodeScrape,
		vmPodScrape,
		vmServiceScrape,
		vmAgent,
	}
}
