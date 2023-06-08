// Package api contains the API server implementation.
//
//nolint:golint,revive,stylecheck //for the sake of using 'someId' instead of the recommended 'someID', since it's generated.
package api

//go:generate ../bin/oapi-codegen --config=server.cfg.yml  ../docs/spec/openapi.yml

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/percona/percona-everest-backend/model"
)

// EverestServer represents the server struct.
type EverestServer struct {
	Storage        storage
	SecretsStorage secretsStorage
}

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

	k, err := e.Storage.CreateKubernetesCluster(ctx, model.CreateKubernetesClusterParams{Name: params.Name})
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	encodedConfig := base64.StdEncoding.EncodeToString([]byte(params.Kubeconfig))

	err = e.SecretsStorage.CreateSecret(ctx, k.ID, encodedConfig)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, k)
}

// GetKubernetesCluster Get the specified kubernetes cluster.
func (e *EverestServer) GetKubernetesCluster(ctx echo.Context, kubernetesId string) error {
	k, err := e.Storage.GetKubernetesCluster(ctx, kubernetesId)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, k)
}

// CreateDatabaseCluster creates a new db cluster inside the given k8s cluster.
func (e *EverestServer) CreateDatabaseCluster(ctx echo.Context, kubernetesId string) error {
	return e.proxyKubernetes(ctx, kubernetesId, "")
}

// ListDatabaseClusterRestores List of the created database cluster restores on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusterRestores(ctx echo.Context, kubernetesId string) error {
	return e.proxyKubernetes(ctx, kubernetesId, "")
}

// CreateDatabaseClusterRestore Create a database cluster restore on the specified kubernetes cluster.
func (e *EverestServer) CreateDatabaseClusterRestore(ctx echo.Context, kubernetesId string) error {
	return e.proxyKubernetes(ctx, kubernetesId, "")
}

// DeleteDatabaseClusterRestore Delete the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) DeleteDatabaseClusterRestore(ctx echo.Context, kubernetesId string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesId, name)
}

// GetDatabaseClusterRestore Returns the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseClusterRestore(ctx echo.Context, kubernetesId string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesId, name)
}

// UpdateDatabaseClusterRestore Replace the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseClusterRestore(ctx echo.Context, kubernetesId string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesId, name)
}

// ListDatabaseClusters List of the created database clusters on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusters(ctx echo.Context, kubernetesId string) error {
	return e.proxyKubernetes(ctx, kubernetesId, "")
}

// DeleteDatabaseCluster Create a database cluster on the specified kubernetes cluster.
func (e *EverestServer) DeleteDatabaseCluster(ctx echo.Context, kubernetesId string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesId, name)
}

// GetDatabaseCluster Get the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseCluster(ctx echo.Context, kubernetesId string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesId, name)
}

// UpdateDatabaseCluster Replace the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseCluster(ctx echo.Context, kubernetesId string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesId, name)
}

// ListDatabaseEngines List of the available database engines on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseEngines(ctx echo.Context, kubernetesId string) error {
	return e.proxyKubernetes(ctx, kubernetesId, "")
}

// GetDatabaseEngine Get the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseEngine(ctx echo.Context, kubernetesId string, name string) error {
	return e.proxyKubernetes(ctx, kubernetesId, name)
}

func (e *EverestServer) proxyKubernetes(ctx echo.Context, kubernetesId, resourceName string) error {
	encodedSecret, err := e.SecretsStorage.GetSecret(ctx, kubernetesId)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", newConfigGetter(encodedSecret).loadFromString)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	reverseProxy := httputil.NewSingleHostReverseProxy(
		&url.URL{ //nolint:exhaustruct
			Host:   strings.TrimPrefix(config.Host, "https://"),
			Scheme: "https",
		})
	transport, err := rest.TransportFor(config)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	reverseProxy.Transport = transport
	req := ctx.Request()
	req.URL.Path = buildProxiedUrl(ctx.Request().URL.Path, kubernetesId, resourceName)
	reverseProxy.ServeHTTP(ctx.Response(), req)
	return nil
}

func buildProxiedUrl(uri, kubernetesId string, resourceName string) string {
	// cut the /kubernetes part
	uri = strings.TrimPrefix(uri, "/kubernetes/"+kubernetesId)

	// cut the resource name if present
	uri = strings.TrimSuffix(uri, resourceName)

	// remove kebab-case
	uri = strings.ReplaceAll(uri, "-", "")
	return fmt.Sprintf("/apis/dbaas.percona.com/v1/namespaces/%s%s%s", "default", uri, resourceName)
}

type List struct {
	Items string `json:"items"`
}
