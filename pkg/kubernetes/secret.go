package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

// CreateSecret creates an ObjectStorage.
func (k *Kubernetes) CreateSecret(ctx context.Context, secret *corev1.Secret) error {
	_, err := k.client.CreateSecret(ctx, secret)
	return err
}
