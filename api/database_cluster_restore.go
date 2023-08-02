package api //nolint:dupl

import "github.com/labstack/echo/v4"

// ListDatabaseClusterRestores List of the created database cluster restores on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusterRestores(ctx echo.Context, kubernetesID string) error {
	return e.everestK8s.ProxyKubernetes(ctx, kubernetesID, "")
}

// CreateDatabaseClusterRestore Create a database cluster restore on the specified kubernetes cluster.
func (e *EverestServer) CreateDatabaseClusterRestore(ctx echo.Context, kubernetesID string) error {
	return e.everestK8s.ProxyKubernetes(ctx, kubernetesID, "")
}

// DeleteDatabaseClusterRestore Delete the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) DeleteDatabaseClusterRestore(ctx echo.Context, kubernetesID string, name string) error {
	return e.everestK8s.ProxyKubernetes(ctx, kubernetesID, name)
}

// GetDatabaseClusterRestore Returns the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseClusterRestore(ctx echo.Context, kubernetesID string, name string) error {
	return e.everestK8s.ProxyKubernetes(ctx, kubernetesID, name)
}

// UpdateDatabaseClusterRestore Replace the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseClusterRestore(ctx echo.Context, kubernetesID string, name string) error {
	return e.everestK8s.ProxyKubernetes(ctx, kubernetesID, name)
}
