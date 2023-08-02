package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

// CreateSecret creates an ObjectStorage.
func (k *Kubernetes) CreateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	return k.client.CreateSecret(ctx, secret)
}

// DeleteSecret deletes an ObjectStorage.
func (k *Kubernetes) DeleteSecret(ctx context.Context, name, namespace string) error {
	return k.client.DeleteSecret(ctx, name, namespace)
}
