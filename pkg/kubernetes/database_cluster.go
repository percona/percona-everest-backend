package kubernetes

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
)

// ListDatabaseClusters returns list of managed database clusters.
func (k *Kubernetes) ListDatabaseClusters(ctx context.Context) (*everestv1alpha1.DatabaseClusterList, error) {
	return k.client.ListDatabaseClusters(ctx)
}

// GetDatabaseCluster returns database clusters by provided name.
func (k *Kubernetes) GetDatabaseCluster(ctx context.Context, name string) (*everestv1alpha1.DatabaseCluster, error) {
	return k.client.GetDatabaseCluster(ctx, name)
}
