package client

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateMonitoringConfig creates an MonitoringConfig.
func (c *Client) CreateMonitoringConfig(ctx context.Context, mc *everestv1alpha1.MonitoringConfig) error {
	_, err := c.customClientSet.MonitoringConfig(c.namespace).Post(ctx, mc, metav1.CreateOptions{})
	return err
}

// GetMonitoringConfig returns the MonitoringConfig.
func (c *Client) GetMonitoringConfig(ctx context.Context, name string) (*everestv1alpha1.MonitoringConfig, error) {
	return c.customClientSet.MonitoringConfig(c.namespace).Get(ctx, name, metav1.GetOptions{})
}

// DeleteMonitoringConfig deletes the MonitoringConfig.
func (c *Client) DeleteMonitoringConfig(ctx context.Context, name string) error {
	return c.customClientSet.MonitoringConfig(c.namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
