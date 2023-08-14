package client

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateObjectStorage creates an objectStorage.
func (c *Client) CreateObjectStorage(ctx context.Context, storage *everestv1alpha1.ObjectStorage) error {
	_, err := c.customClientSet.ObjectStorage(storage.Namespace).Post(ctx, storage, metav1.CreateOptions{})
	return err
}

// UpdateObjectStorage updates an objectStorage.
func (c *Client) UpdateObjectStorage(ctx context.Context, storage *everestv1alpha1.ObjectStorage) error {
	_, err := c.customClientSet.ObjectStorage(storage.Namespace).Update(ctx, storage, metav1.UpdateOptions{})
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
