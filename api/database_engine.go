package api

import "github.com/labstack/echo/v4"

// ListDatabaseEngines List of the available database engines on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseEngines(ctx echo.Context, kubernetesID string) error {
	return e.everestK8s.ProxyKubernetes(ctx, kubernetesID, "")
}

// GetDatabaseEngine Get the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseEngine(ctx echo.Context, kubernetesID string, name string) error {
	return e.everestK8s.ProxyKubernetes(ctx, kubernetesID, name)
}

// UpdateDatabaseEngine Get the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseEngine(ctx echo.Context, kubernetesID string, name string) error {
	return e.everestK8s.ProxyKubernetes(ctx, kubernetesID, name)
}
