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

//go:generate ../bin/oapi-codegen --config=server.cfg.yml  ../docs/spec/openapi.yml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/percona/percona-everest-backend/cmd/config"
	"github.com/percona/percona-everest-backend/model"
)

const (
	pgStorageName   = "postgres"
	pgMigrationsDir = "migrations"
)

// EverestServer represents the server struct.
type EverestServer struct {
	config         *config.EverestConfig
	l              *zap.SugaredLogger
	storage        storage
	secretsStorage secretsStorage
}

// List represents a general object with the list of items.
type List struct {
	Items string `json:"items"`
}

// NewEverestServer creates and configures everest API.
func NewEverestServer(c *config.EverestConfig, l *zap.SugaredLogger) (*EverestServer, error) {
	e := &EverestServer{
		config: c,
		l:      l,
	}
	err := e.initStorages()

	return e, err
}

func (e *EverestServer) initStorages() error {
	db, err := model.NewDatabase(pgStorageName, e.config.DSN, pgMigrationsDir)
	if err != nil {
		return err
	}
	e.storage = db
	e.secretsStorage = db // so far the db implements both interfaces - the regular storage and the secrets storage
	_, err = db.Migrate()
	return err
}

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

func (e *EverestServer) doK8sRequest(ctx echo.Context, kubernetesID, resourceName string, body any) ([]byte, int, error) {
	cluster, err := e.storage.GetKubernetesCluster(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("Could not find Kubernetes cluster")
	}

	encodedSecret, err := e.secretsStorage.GetSecret(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", newConfigGetter(encodedSecret).loadFromString)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	j, err := json.Marshal(body)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("Could not marshal database cluster")
	}

	url := &url.URL{
		Host:   strings.TrimPrefix(config.Host, "https://"),
		Scheme: "https",
		Path:   buildProxiedURL(ctx.Request().URL.Path, kubernetesID, resourceName, cluster.Namespace),
	}
	req, err := http.NewRequest(ctx.Request().Method, url.String(), bytes.NewBuffer(j))
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("Could not create Kubernetes request")
	}

	transport, err := rest.TransportFor(config)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	client := &http.Client{Transport: transport}
	res, err := client.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("Could not send request to Kubernetes")
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Errorf("Could not get response from Kubernetes. Status HTTP %d", res.StatusCode)
	}

	if res.StatusCode >= http.StatusBadRequest {
		e.l.Errorf("Received non-2xx response from Kubernetes. HTTP %d", res.StatusCode)
		e.l.Debug(string(b))
		return nil, http.StatusInternalServerError, errors.New("Received invalid response from Kubernetes")
	}

	return b, http.StatusOK, nil
}

func (e *EverestServer) assignFieldBetweenStructs(from any, to any) error {
	fromJson, err := json.Marshal(from)
	if err != nil {
		return errors.New("Could not marshal field")
	}

	if err := json.Unmarshal(fromJson, to); err != nil {
		return errors.New("Could not unmarshal field")
	}

	return nil
}
