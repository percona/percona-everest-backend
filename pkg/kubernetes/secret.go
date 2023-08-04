package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

// GetSecret returns secret by name.
func (k *Kubernetes) GetSecret(ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	return k.client.GetSecret(ctx, name, namespace)
}
