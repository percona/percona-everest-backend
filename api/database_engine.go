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

// Package api ...
package api

import (
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/percona/percona-everest-backend/client"
)

// ListDatabaseEngines List of the available database engines on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseEngines(ctx echo.Context, kubernetesID string) error {
	cl, statusCode, err := e.getK8sClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	engineList, err := cl.ListDatabaseEngines(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	items := make([]client.DatabaseEngineWithName, 0, len(engineList.Items))
	res := &client.DatabaseEngineList{Items: &items}
	for _, i := range engineList.Items {
		i := i
		d, err := e.parseDBEngineObj(&i)
		if err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
		}
		*res.Items = append(*res.Items, *d)
	}

	return ctx.JSON(http.StatusOK, res)
}

// GetDatabaseEngine Get the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseEngine(ctx echo.Context, kubernetesID string, name string) error {
	cl, statusCode, err := e.getK8sClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	engine, err := cl.GetDatabaseEngine(ctx.Request().Context(), name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(statusCode, Error{
			Message: pointer.ToString("Could not get database engine"),
		})
	}

	d, err := e.parseDBEngineObj(engine)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, d)
}

// UpdateDatabaseEngine Get the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseEngine(ctx echo.Context, kubernetesID string, name string) error {
	var params UpdateDatabaseEngineJSONRequestBody
	if err := ctx.Bind(&params); err != nil {
		return err
	}

	cl, statusCode, err := e.getK8sClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	engine := &everestv1alpha1.DatabaseEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if err := e.assignFieldBetweenStructs(params.Spec, &engine.Spec); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	_, err = cl.UpdateDatabaseEngine(ctx.Request().Context(), name, engine)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not update database engine in Kubernetes"),
		})
	}

	return ctx.NoContent(http.StatusOK)
}

func (e *EverestServer) parseDBEngineObj(engine *everestv1alpha1.DatabaseEngine) (*client.DatabaseEngineWithName, error) {
	d := &client.DatabaseEngineWithName{
		Name: engine.Name,
	}

	if err := e.assignFieldBetweenStructs(engine.Spec, &d.Spec); err != nil {
		return nil, errors.New("Could not parse database engine spec")
	}
	if err := e.assignFieldBetweenStructs(engine.Status, &d.Status); err != nil {
		return nil, errors.New("Could not parse database engine status")
	}

	return d, nil
}
