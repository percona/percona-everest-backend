package api //nolint:dupl

import (
	"log"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/percona/percona-everest-backend/pkg/kubernetes"
)

// CreateDatabaseCluster creates a new db cluster inside the given k8s cluster.
func (e *EverestServer) CreateDatabaseCluster(ctx echo.Context, kubernetesID string) error {
	return e.proxyKubernetes(ctx, kubernetesID, "")
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

func (e *EverestServer) GetDatabaseClusterCredentials(ctx echo.Context, kubernetesID string, name string) error {
	cluster, err := e.storage.GetKubernetesCluster(ctx.Request().Context(), kubernetesID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	kubeClient, err := kubernetes.NewFromSecretsStorage(ctx.Request().Context(), e.secretsStorage, cluster.ID, cluster.Namespace, logrus.NewEntry(logrus.StandardLogger()))
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	databaseCluster, err := kubeClient.GetDatabaseCluster(ctx.Request().Context(), name)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	secret, err := kubeClient.GetSecret(ctx.Request().Context(), databaseCluster.Spec.Engine.UserSecretsName, cluster.Namespace)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	_ = secret

	return nil
}
