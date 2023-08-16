package customresouces

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const (
	backupStorageAPIKind = "backupstorages"
)

// BackupStorage returns a db cluster client.
func (c *Client) BackupStorage( //nolint:ireturn
	namespace string,
) BackupStoragesInterface {
	return &client{
		restClient: c.restClient,
		namespace:  namespace,
	}
}

// BackupStoragesInterface supports methods to work with BackupStorages.
type BackupStoragesInterface interface {
	Post(ctx context.Context, storage *everestv1alpha1.BackupStorage, opts metav1.CreateOptions) (*everestv1alpha1.BackupStorage, error)
	Update(ctx context.Context, storage *everestv1alpha1.BackupStorage, opts metav1.UpdateOptions) (*everestv1alpha1.BackupStorage, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*everestv1alpha1.BackupStorage, error)
}

type client struct {
	restClient rest.Interface
	namespace  string
}

// Post creates a resource.
func (c *client) Post(
	ctx context.Context,
	storage *everestv1alpha1.BackupStorage,
	opts metav1.CreateOptions,
) (*everestv1alpha1.BackupStorage, error) {
	result := &everestv1alpha1.BackupStorage{}
	err := c.restClient.
		Post().
		Namespace(c.namespace).
		Resource(backupStorageAPIKind).Body(storage).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).Into(result)
	return result, err
}

// Update creates a resource.
func (c *client) Update(
	ctx context.Context,
	storage *everestv1alpha1.BackupStorage,
	opts metav1.UpdateOptions,
) (*everestv1alpha1.BackupStorage, error) {
	result := &everestv1alpha1.BackupStorage{}
	err := c.restClient.
		Put().Name(storage.Name).
		Namespace(c.namespace).
		Resource(backupStorageAPIKind).Body(storage).
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
		Delete().Name(name).
		Namespace(c.namespace).
		Resource(backupStorageAPIKind).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).Error()
}

// Get retrieves database cluster based on opts.
func (c *client) Get(
	ctx context.Context,
	name string,
	opts metav1.GetOptions,
) (*everestv1alpha1.BackupStorage, error) {
	result := &everestv1alpha1.BackupStorage{}
	err := c.restClient.
		Get().
		Namespace(c.namespace).
		Resource(backupStorageAPIKind).
		VersionedParams(&opts, scheme.ParameterCodec).
		Name(name).
		Do(ctx).
		Into(result)
	return result, err
}
