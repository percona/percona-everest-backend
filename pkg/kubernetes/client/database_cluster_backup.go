package client

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListDatabaseClusterBackups returns list of managed database cluster backups.
func (c *Client) ListDatabaseClusterBackups(ctx context.Context, options metav1.ListOptions) (*everestv1alpha1.DatabaseClusterBackupList, error) {
	return c.customClientSet.DBClusterBackups(c.namespace).List(ctx, options)
}

// GetDatabaseClusterBackup returns database cluster backups by provided name.
func (c *Client) GetDatabaseClusterBackup(ctx context.Context, name string) (*everestv1alpha1.DatabaseClusterBackup, error) {
	return c.customClientSet.DBClusterBackups(c.namespace).Get(ctx, name, metav1.GetOptions{})
}
