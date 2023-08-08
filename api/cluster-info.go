package api

import (
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
)

func (e *EverestServer) GetKubernetesClusterInfo(ctx echo.Context, kubernetesID string) error {
	_, kubeClient, code, err := e.initKubeClient(ctx, kubernetesID)
	if err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}
	clusterType, err := kubeClient.GetClusterType(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}
	storagesList, err := kubeClient.GetStorageClasses(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}
	var classNames []string
	for _, storageClass := range storagesList.Items {
		classNames = append(classNames, storageClass.Name)
	}

	return ctx.JSON(http.StatusOK, &KubernetesClusterInfo{ClusterType: string(clusterType), StorageClassNames: classNames})
}
