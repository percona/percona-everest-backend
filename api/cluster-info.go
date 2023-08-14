package api

import (
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
)

// GetKubernetesClusterInfo returns the cluster type and storage classes of a kubernetes cluster.
func (e *EverestServer) GetKubernetesClusterInfo(ctx echo.Context, kubernetesID string) error {
	_, kubeClient, code, err := e.initKubeClient(ctx, kubernetesID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(code, Error{Message: pointer.ToString("Failed building connection to the Kubernetes cluster")})
	}
	clusterType, err := kubeClient.GetClusterType(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Failed getting Kubernetes cluster provider")})
	}
	storagesList, err := kubeClient.GetStorageClasses(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Failed getting storage classes")})
	}
	classNames := make([]string, len(storagesList.Items))
	for i, storageClass := range storagesList.Items {
		classNames[i] = storageClass.Name
	}

	return ctx.JSON(http.StatusOK, &KubernetesClusterInfo{ClusterType: string(clusterType), StorageClassNames: classNames})
}
