package client

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateMonitoringConfig creates an monitoringConfig.
func (c *Client) CreateMonitoringConfig(ctx context.Context, storage *everestv1alpha1.MonitoringConfig) error {
	_, err := c.customClientSet.MonitoringConfig(storage.Namespace).Create(ctx, storage, metav1.CreateOptions{})
	return err
}

// UpdateMonitoringConfig updates an monitoringConfig.
func (c *Client) UpdateMonitoringConfig(ctx context.Context, storage *everestv1alpha1.MonitoringConfig) error {
	_, err := c.customClientSet.MonitoringConfig(storage.Namespace).Update(ctx, storage, metav1.UpdateOptions{})
	return err
}

// GetMonitoringConfig returns the monitoringConfig.
func (c *Client) GetMonitoringConfig(ctx context.Context, name string) (*everestv1alpha1.MonitoringConfig, error) {
	return c.customClientSet.MonitoringConfig(c.namespace).Get(ctx, name, metav1.GetOptions{})
}

// ListMonitoringConfigs returns the monitoringConfig.
func (c *Client) ListMonitoringConfigs(ctx context.Context) (*everestv1alpha1.MonitoringConfigList, error) {
	return c.customClientSet.MonitoringConfig(c.namespace).List(ctx, metav1.ListOptions{})
}

// DeleteMonitoringConfig deletes the monitoringConfig.
func (c *Client) DeleteMonitoringConfig(ctx context.Context, name string) error {
	return c.customClientSet.MonitoringConfig(c.namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
