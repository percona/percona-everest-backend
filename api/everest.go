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
	"context"
	"encoding/json"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/go-logr/zapr"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/percona/percona-everest-backend/cmd/config"
	"github.com/percona/percona-everest-backend/model"
	"github.com/percona/percona-everest-backend/pkg/kubernetes"
	"github.com/percona/percona-everest-backend/pkg/logger"
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

func (e *EverestServer) getK8sClient(ctx context.Context, kubernetesID string) (*kubernetes.Kubernetes, int, error) {
	k, err := e.storage.GetKubernetesCluster(ctx, kubernetesID)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("Could not find Kubernetes cluster")
	}

	l := logger.MustInitLogger()
	client, err := kubernetes.NewFromSecretsStorage(
		ctx, e.secretsStorage, k.ID,
		k.Namespace, zapr.NewLogger(l),
	)
	if err != nil {
		e.l.Error(err)
		return nil, http.StatusInternalServerError, errors.New("Could not create Kubernetes client from kubeconfig")
	}

	return client, 0, nil
}

func (e *EverestServer) assignFieldBetweenStructs(from any, to any) error {
	fromJSON, err := json.Marshal(from)
	if err != nil {
		return errors.New("Could not marshal field")
	}

	if err := json.Unmarshal(fromJSON, to); err != nil {
		return errors.New("Could not unmarshal field")
	}

	return nil
}

func (e *EverestServer) createResource(
	ctx echo.Context, kubernetesID string,
	spec any, intoSpec any, obj runtime.Object,
) error {
	cl, statusCode, err := e.getK8sClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	if err := e.assignFieldBetweenStructs(spec, intoSpec); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	err = cl.CreateResource(ctx.Request().Context(), obj, obj)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not create new resource in Kubernetes"),
		})
	}

	return ctx.NoContent(http.StatusOK)
}

func (e *EverestServer) updateResource(
	ctx echo.Context, kubernetesID string,
	name string, spec any, intoSpec any, obj runtime.Object,
) error {
	cl, statusCode, err := e.getK8sClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	if err := e.assignFieldBetweenStructs(spec, intoSpec); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	err = cl.UpdateResource(ctx.Request().Context(), name, obj, obj)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not update resource in Kubernetes"),
		})
	}

	return ctx.NoContent(http.StatusOK)
}
