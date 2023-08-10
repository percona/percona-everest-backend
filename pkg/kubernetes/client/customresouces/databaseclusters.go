package customresouces

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const (
	dbClustersAPIKind = "databaseclusters"
)

// DBClusters returns a db cluster client.
func (c *Client) DBClusters(namespace string) DBClusterInterface { //nolint:ireturn
	return &dbClusterClient{
		restClient: c.restClient,
		namespace:  namespace,
	}
}

type dbClusterClient struct {
	restClient rest.Interface
	namespace  string
}

// DBClusterInterface supports list, get and watch methods.
type DBClusterInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*everestv1alpha1.DatabaseClusterList, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*everestv1alpha1.DatabaseCluster, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

// List lists database clusters based on opts.
func (c *dbClusterClient) List(ctx context.Context, opts metav1.ListOptions) (*everestv1alpha1.DatabaseClusterList, error) {
	result := &everestv1alpha1.DatabaseClusterList{}
	err := c.restClient.
		Get().
		Namespace(c.namespace).
		Resource(dbClustersAPIKind).
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
		Resource(dbClustersAPIKind).
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
		Resource(dbClustersAPIKind).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}
