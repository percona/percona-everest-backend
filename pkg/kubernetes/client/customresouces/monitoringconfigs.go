package customresouces

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const monitoringConfigAPIKind = "monitoringconfigs"

// MonitoringConfig returns a db cluster client.
func (c *Client) MonitoringConfig( //nolint:ireturn
	namespace string,
) MonitoringConfigsInterface {
	return &monitoringConfigClient{
		restClient: c.restClient,
		namespace:  namespace,
	}
}

// MonitoringConfigsInterface supports methods to work with MonitoringConfig.
type MonitoringConfigsInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*everestv1alpha1.MonitoringConfigList, error)
	Post(ctx context.Context, storage *everestv1alpha1.MonitoringConfig, opts metav1.CreateOptions) (*everestv1alpha1.MonitoringConfig, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*everestv1alpha1.MonitoringConfig, error)
}

type monitoringConfigClient struct {
	restClient rest.Interface
	namespace  string
}

// List lists database clusters based on opts.
func (c *monitoringConfigClient) List(ctx context.Context, opts metav1.ListOptions) (*everestv1alpha1.MonitoringConfigList, error) {
	result := &everestv1alpha1.MonitoringConfigList{}
	err := c.restClient.
		Get().
		Namespace(c.namespace).
		Resource(monitoringConfigAPIKind).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}

// Post creates a resource.
func (c *monitoringConfigClient) Post(
	ctx context.Context,
	storage *everestv1alpha1.MonitoringConfig,
	opts metav1.CreateOptions,
) (*everestv1alpha1.MonitoringConfig, error) {
	result := &everestv1alpha1.MonitoringConfig{}
	err := c.restClient.
		Post().
		Namespace(c.namespace).
		Resource(monitoringConfigAPIKind).Body(storage).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).Into(result)
	return result, err
}

// Delete creates a resource.
func (c *monitoringConfigClient) Delete(
	ctx context.Context,
	name string,
	opts metav1.DeleteOptions,
) error {
	return c.restClient.
		Delete().Name(name).
		Namespace(c.namespace).
		Resource(monitoringConfigAPIKind).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).Error()
}

// Get retrieves database cluster based on opts.
func (c *monitoringConfigClient) Get(
	ctx context.Context,
	name string,
	opts metav1.GetOptions,
) (*everestv1alpha1.MonitoringConfig, error) {
	result := &everestv1alpha1.MonitoringConfig{}
	err := c.restClient.
		Get().
		Namespace(c.namespace).
		Resource(monitoringConfigAPIKind).
		VersionedParams(&opts, scheme.ParameterCodec).
		Name(name).
		Do(ctx).
		Into(result)
	return result, err
}
