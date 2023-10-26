package client

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateBackupStorage creates an backupStorage.
func (c *Client) CreateBackupStorage(ctx context.Context, storage *everestv1alpha1.BackupStorage) error {
	_, err := c.customClientSet.BackupStorage(storage.Namespace).Create(ctx, storage, metav1.CreateOptions{})
	return err
}

// UpdateBackupStorage updates an backupStorage.
func (c *Client) UpdateBackupStorage(ctx context.Context, storage *everestv1alpha1.BackupStorage) error {
	_, err := c.customClientSet.BackupStorage(storage.Namespace).Update(ctx, storage, metav1.UpdateOptions{})
	return err
}

// GetBackupStorage returns the backupStorage.
func (c *Client) GetBackupStorage(ctx context.Context, name string) (*everestv1alpha1.BackupStorage, error) {
	return c.customClientSet.BackupStorage(c.namespace).Get(ctx, name, metav1.GetOptions{})
}

// ListBackupStorages returns the backupStorage.
func (c *Client) ListBackupStorages(ctx context.Context) (*everestv1alpha1.BackupStorageList, error) {
	return c.customClientSet.BackupStorage(c.namespace).List(ctx, metav1.ListOptions{})
}

// DeleteBackupStorage deletes the backupStorage.
func (c *Client) DeleteBackupStorage(ctx context.Context, name string) error {
	return c.customClientSet.BackupStorage(c.namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
