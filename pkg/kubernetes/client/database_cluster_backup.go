package client

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListDatabaseClusterBackups returns list of managed database clusters.
func (c *Client) ListDatabaseClusterBackups(ctx context.Context) (*everestv1alpha1.DatabaseClusterBackupList, error) {
	return c.customClientSet.DBClusterBackups(c.namespace).List(ctx, metav1.ListOptions{})
}

// GetDatabaseClusterBackup returns database clusters by provided name.
func (c *Client) GetDatabaseClusterBackup(ctx context.Context, name string) (*everestv1alpha1.DatabaseClusterBackup, error) {
	return c.customClientSet.DBClusterBackups(c.namespace).Get(ctx, name, metav1.GetOptions{})
}
