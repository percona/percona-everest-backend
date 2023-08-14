// percona-everest-backend
// Copyright (C) 2023 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package api contains the API server implementation.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func (e *EverestServer) proxyKubernetes(ctx echo.Context, kubernetesID, resourceName string) error {
	cluster, err := e.storage.GetKubernetesCluster(ctx.Request().Context(), kubernetesID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get a Kubernetes cluster"),
		})
	}
	encodedSecret, err := e.secretsStorage.GetSecret(ctx.Request().Context(), kubernetesID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not retrieve kubeconfig"),
		})
	}

	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", newConfigGetter(encodedSecret).loadFromString)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not build kubeconfig"),
		})
	}
	reverseProxy := httputil.NewSingleHostReverseProxy(
		&url.URL{
			Host:   strings.TrimPrefix(config.Host, "https://"),
			Scheme: "https",
		})
	transport, err := rest.TransportFor(config)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not create REST transport"),
		})
	}
	reverseProxy.Transport = transport
	reverseProxy.ErrorHandler = everestErrorHandler(cluster.Name, e.l)
	req := ctx.Request()
	req.URL.Path = buildProxiedURL(ctx.Request().URL.Path, kubernetesID, resourceName, cluster.Namespace)
	reverseProxy.ServeHTTP(ctx.Response(), req)
	return nil
}

func buildProxiedURL(uri, kubernetesID, resourceName, namespace string) string {
	// cut the /kubernetes part
	uri = strings.TrimPrefix(uri, "/v1/kubernetes/"+kubernetesID)

	// cut the resource name if present
	uri = strings.TrimSuffix(uri, resourceName)

	// remove kebab-case
	uri = strings.ReplaceAll(uri, "-", "")
	return fmt.Sprintf("/apis/everest.percona.com/v1alpha1/namespaces/%s%s%s", namespace, uri, resourceName)
}

func everestErrorHandler(clusterName string, logger *zap.SugaredLogger) func(http.ResponseWriter, *http.Request, error) {
	errorMessage := fmt.Sprintf("%s kubernetes cluster is unavailable", clusterName)
	b, err := json.Marshal(Error{Message: pointer.ToString(errorMessage)})
	if err != nil {
		logger.Error(err.Error())
	}
	return func(res http.ResponseWriter, req *http.Request, err error) {
		if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, os.ErrDeadlineExceeded) {
			res.WriteHeader(http.StatusInternalServerError)
			if _, err := res.Write(b); err != nil {
				logger.Error(err.Error())
			}
			return
		}
		res.WriteHeader(http.StatusBadGateway)
		if _, err := res.Write(b); err != nil {
			logger.Error(err.Error())
		}
	}
}
