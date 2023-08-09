package api

import "github.com/labstack/echo/v4"

// ListDatabaseClusterBackups returns list of the created database cluster backups on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusterBackups(ctx echo.Context, kubernetesID string) error {
	return e.proxyKubernetes(ctx, kubernetesID, "")
}

// CreateDatabaseBackup creates a database cluster backup on the specified kubernetes cluster
func (e *EverestServer) CreateDatabaseClusterBackup(ctx echo.Context, kubernetesID string) error {
	return e.proxyKubernetes(ctx, kubernetesID, "")
}

// DeleteDatabaseClusterBackup deletes the specified cluster backup on the specified kubernetes cluster
func (e *EverestServer) DeleteDatabaseClusterBackup(ctx echo.Context, kubernetesID string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesID, name)
}

// GetDatabaseClusterBacup returns the specified cluster backup on the specified kubernetes cluster
func (e *EverestServer) GetDatabaseClusterBackup(ctx echo.Context, kubernetesID string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesID, name)
}

// UpdateDatabaseClusterBackup replaces the specified cluster backup on the specified kubernetes cluster
func (e *EverestServer) UpdateDatabaseClusterBackup(ctx echo.Context, kubernetesID string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesID, name)
}
