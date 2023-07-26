package api //nolint:dupl

import (
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
)

// CreateDatabaseCluster creates a new db cluster inside the given k8s cluster.
func (e *EverestServer) CreateDatabaseCluster(ctx echo.Context, kubernetesID string) error {
	var params CreateDatabaseClusterJSONRequestBody
	if err := ctx.Bind(&params); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	// Here we save/process any Everest related fields
	e.storage.SaveNotificationConfig(ctx, kubernetesID, params.Notification)

	// Apply e.g. resource limits
	if e.limits.IsOverTeamCPULimit(params.Spec.Engine.Resources) {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("CPU limits reached for your team")})
	}

	// Talk to k8s
	e.k8s.applyObject(everestv1alpha.DatabaseCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "everest.percona.com/v1alpha1",
			Kind:       "DatabaseCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   params.Identifier,
			Labels: params.Labels,
		},
		Spec: params.Spec,
	})

	return ctx.NoContent(200)
}

// ListDatabaseClusters List of the created database clusters on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusters(ctx echo.Context, kubernetesID string) error {
	return e.proxyKubernetes(ctx, kubernetesID, "")
}

// DeleteDatabaseCluster Create a database cluster on the specified kubernetes cluster.
func (e *EverestServer) DeleteDatabaseCluster(ctx echo.Context, kubernetesID string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesID, name)
}

// GetDatabaseCluster Get the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseCluster(ctx echo.Context, kubernetesID string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesID, name)
}

// UpdateDatabaseCluster Replace the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseCluster(ctx echo.Context, kubernetesID string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesID, name)
}
