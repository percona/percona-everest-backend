// Package api contains the API server implementation.
//
//nolint:golint,revive,stylecheck //for the sake of using 'someId' instead of the recommended 'someID', since it's generated.
package api

//go:generate ../bin/oapi-codegen --config=server.cfg.yml  ../docs/spec/openapi.yml

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	vault "github.com/hashicorp/vault/api"
	"github.com/labstack/echo/v4"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// EverestServer represents the server struct.
type EverestServer struct {
	v *vault.Client
}

// NewEverestServer creates a new instance of Everest server.
func NewEverestServer() (*EverestServer, error) {
	config := vault.DefaultConfig()
	config.Address = "http://127.0.0.1:8200"
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetToken("myroot")
	return &EverestServer{v: client}, nil
}

// ListKubernetesClusters returns list of k8s clusters.
func (e *EverestServer) ListKubernetesClusters(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, nil)
}

// RegisterKubernetesCluster registers a k8s cluster in Everest server.
func (e *EverestServer) RegisterKubernetesCluster(ctx echo.Context) error {
	var params CreateKubernetesClusterParams
	if err := ctx.Bind(&params); err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	_, err := clientcmd.BuildConfigFromKubeconfigGetter("", newConfigGetter(*params.Kubeconfig).loadFromString)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	m := map[string]interface{}{
		"kubeconfig": params.Kubeconfig,
	}

	_, err = e.v.KVv2("secret").Put(context.TODO(), *params.Name, m)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	//nolint:godox
	// TODO: store in db
	id := uuid.NewString()
	k := KubernetesCluster{&id, params.Name}

	return ctx.JSON(http.StatusOK, k)
}

// GetKubernetesCluster Get the specified kubernetes cluster.
func (e *EverestServer) GetKubernetesCluster(ctx echo.Context, kubernetesId string) error {
	log.Println(kubernetesId)
	return ctx.JSON(http.StatusNotImplemented, nil)
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
	secret, err := e.v.KVv2("secret").Get(context.TODO(), kubernetesId)
	kubeconfig, ok := secret.Data["kubeconfig"].(string)
	if !ok {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", newConfigGetter(kubeconfig).loadFromString)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	data, err := json.Marshal(secret)
	if err != nil {
		log.Println(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	err = json.Unmarshal(data, config)
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
