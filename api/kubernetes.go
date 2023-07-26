package api

import (
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/percona/percona-everest-backend/model"
)

// ListKubernetesClusters returns list of k8s clusters.
func (e *EverestServer) ListKubernetesClusters(ctx echo.Context) error {
	list, err := e.storage.ListKubernetesClusters(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	result := make([]KubernetesCluster, 0, len(list))
	for _, k := range list {
		result = append(result, KubernetesCluster{
			Id:   k.ID,
			Name: k.Name,
		})
	}

	return ctx.JSON(http.StatusOK, result)
}

// RegisterKubernetesCluster registers a k8s cluster in Everest server.
func (e *EverestServer) RegisterKubernetesCluster(ctx echo.Context) error {
	var params CreateKubernetesClusterParams
	if err := ctx.Bind(&params); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	c := ctx.Request().Context()

	_, err := clientcmd.BuildConfigFromKubeconfigGetter("", newConfigGetter(params.Kubeconfig).loadFromString)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	k, err := e.storage.CreateKubernetesCluster(c, model.CreateKubernetesClusterParams{
		Name:      params.Name,
		Namespace: params.Namespace,
	})
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	err = e.secretsStorage.CreateSecret(c, k.ID, params.Kubeconfig)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	result := KubernetesCluster{
		Id:   k.ID,
		Name: k.Name,
	}
	return ctx.JSON(http.StatusOK, result)
}

// GetKubernetesCluster Get the specified kubernetes cluster.
func (e *EverestServer) GetKubernetesCluster(ctx echo.Context, kubernetesID string) error {
	k, err := e.storage.GetKubernetesCluster(ctx.Request().Context(), kubernetesID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	result := KubernetesCluster{
		Id:   k.ID,
		Name: k.Name,
	}
	return ctx.JSON(http.StatusOK, result)
}
