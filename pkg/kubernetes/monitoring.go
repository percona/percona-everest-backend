package kubernetes

import "context"

// DeleteAllMonitoringResources deletes all resources related to monitoring from k8s cluster.
func (k *Kubernetes) DeleteAllMonitoringResources(ctx context.Context) error {
	return k.client.DeleteAllMonitoringResources(ctx)
}
