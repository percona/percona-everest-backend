package api

import (
	"log"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/percona/percona-everest-backend/model"
)

// ListKubernetesClusters returns list of k8s clusters.
func (e *EverestServer) ListKubernetesClusters(ctx echo.Context) error {
	list, err := e.Storage.ListKubernetesClusters(ctx)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, list)
}

// RegisterKubernetesCluster registers a k8s cluster in Everest server.
func (e *EverestServer) RegisterKubernetesCluster(ctx echo.Context) error {
	var params CreateKubernetesClusterParams
	if err := ctx.Bind(&params); err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	_, err := clientcmd.BuildConfigFromKubeconfigGetter("", newConfigGetter(params.Kubeconfig).loadFromString)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	k, err := e.Storage.CreateKubernetesCluster(ctx, model.CreateKubernetesClusterParams{
		Name:      params.Name,
		Namespace: params.Namespace,
	})
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	err = e.SecretsStorage.CreateSecret(ctx, k.ID, params.Kubeconfig)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, k)
}

// GetKubernetesCluster Get the specified kubernetes cluster.
func (e *EverestServer) GetKubernetesCluster(ctx echo.Context, kubernetesID string) error {
	k, err := e.Storage.GetKubernetesCluster(ctx, kubernetesID)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, k)
}
