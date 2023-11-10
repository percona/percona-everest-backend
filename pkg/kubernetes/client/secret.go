package client

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetSecret returns secret by name.
func (c *Client) GetSecret(ctx context.Context, name string) (*corev1.Secret, error) {
	return c.clientset.CoreV1().Secrets(c.namespace).Get(ctx, name, metav1.GetOptions{})
}

// UpdateSecret updates k8s Secret.
func (c *Client) UpdateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	return c.clientset.CoreV1().Secrets(c.namespace).Update(ctx, secret, metav1.UpdateOptions{})
}

// CreateSecret creates k8s Secret.
func (c *Client) CreateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	return c.clientset.CoreV1().Secrets(c.namespace).Create(ctx, secret, metav1.CreateOptions{})
}

// DeleteSecret deletes the k8s Secret.
func (c *Client) DeleteSecret(ctx context.Context, name string) error {
	return c.clientset.CoreV1().Secrets(c.namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
