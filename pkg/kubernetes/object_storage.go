package kubernetes

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
)

// CreateObjectStorage creates an ObjectStorage.
func (k *Kubernetes) CreateObjectStorage(ctx context.Context, objectStorage *everestv1alpha1.ObjectStorage) error {
	return k.client.CreateObjectStorage(ctx, objectStorage)
}
