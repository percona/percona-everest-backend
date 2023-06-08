// Package api contains the API server implementation.
//
//nolint:golint,revive,stylecheck //for the sake of using 'someId' instead of the recommended 'someID', since it's generated.
package api

//go:generate ../bin/oapi-codegen --config=server.cfg.yml  ../docs/spec/openapi.yml

import (
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
)

// EverestServer represents the server struct.
type EverestServer struct {
	Storage        storage
	SecretsStorage secretsStorage
}

type List struct {
	Items string `json:"items"`
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
