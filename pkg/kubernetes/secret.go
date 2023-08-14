package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

// GetSecret returns secret by name.
func (k *Kubernetes) GetSecret(ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	return k.client.GetSecret(ctx, name, namespace)
}

// CreateSecret creates an ObjectStorage.
func (k *Kubernetes) CreateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	return k.client.CreateSecret(ctx, secret)
}

// UpdateSecret creates an ObjectStorage.
func (k *Kubernetes) UpdateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	return k.client.UpdateSecret(ctx, secret)
}

// DeleteSecret deletes an ObjectStorage.
func (k *Kubernetes) DeleteSecret(ctx context.Context, name, namespace string) error {
	return k.client.DeleteSecret(ctx, name, namespace)
}
