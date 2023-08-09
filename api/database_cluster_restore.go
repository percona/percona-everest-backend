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

// ListDatabaseClusterRestores List of the created database cluster restores on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusterRestores(ctx echo.Context, kubernetesID string) error {
	cl, statusCode, err := e.getK8sClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	restores, err := cl.ListDatabaseClusterRestores(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	items := make([]client.DatabaseClusterRestoreWithName, 0, len(restores.Items))
	res := &client.DatabaseClusterRestoreList{Items: &items}
	for _, i := range restores.Items {
		i := i
		d, err := e.parseDBClusterRestoreObj(&i)
		if err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
		}
		*res.Items = append(*res.Items, *d)
	}

	return ctx.JSON(http.StatusOK, res)
}

// CreateDatabaseClusterRestore Create a database cluster restore on the specified kubernetes cluster.
func (e *EverestServer) CreateDatabaseClusterRestore(ctx echo.Context, kubernetesID string) error {
	var params CreateDatabaseClusterRestoreJSONRequestBody
	if err := ctx.Bind(&params); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	restore := &everestv1alpha1.DatabaseClusterRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name: params.Name,
		},
	}

	return e.createResource(ctx, kubernetesID, params.Spec, &restore.Spec, restore)
}

// DeleteDatabaseClusterRestore Delete the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) DeleteDatabaseClusterRestore(ctx echo.Context, kubernetesID string, name string) error {
	cl, statusCode, err := e.getK8sClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	if err := cl.DeleteDatabaseClusterRestore(ctx.Request().Context(), name); err != nil {
		e.l.Error(err)
		return ctx.JSON(statusCode, Error{
			Message: pointer.ToString("Could not delete database cluster restore"),
		})
	}

	return ctx.NoContent(http.StatusOK)
}

// GetDatabaseClusterRestore Returns the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseClusterRestore(ctx echo.Context, kubernetesID string, name string) error {
	cl, statusCode, err := e.getK8sClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	restore, err := cl.GetDatabaseClusterRestore(ctx.Request().Context(), name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(statusCode, Error{
			Message: pointer.ToString("Could not get database cluster restore"),
		})
	}

	d, err := e.parseDBClusterRestoreObj(restore)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, d)
}

// UpdateDatabaseClusterRestore Replace the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseClusterRestore(ctx echo.Context, kubernetesID string, name string) error {
	var params UpdateDatabaseClusterRestoreJSONRequestBody
	if err := ctx.Bind(&params); err != nil {
		return err
	}

	restore := &everestv1alpha1.DatabaseClusterRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	return e.updateResource(ctx, kubernetesID, name, params.Spec, &restore.Spec, restore)
}

func (e *EverestServer) parseDBClusterRestoreObj(restore *everestv1alpha1.DatabaseClusterRestore) (*client.DatabaseClusterRestoreWithName, error) {
	d := &client.DatabaseClusterRestoreWithName{
		Name: restore.Name,
	}

	if err := e.assignFieldBetweenStructs(restore.Spec, &d.Spec); err != nil {
		return nil, errors.New("Could not parse database cluster restore spec")
	}
	if err := e.assignFieldBetweenStructs(restore.Status, &d.Status); err != nil {
		return nil, errors.New("Could not parse database cluster restore status")
	}

	return d, nil
}
