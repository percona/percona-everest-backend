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
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var restoreStatuses = map[string]struct{}{ //nolint:gochecknoglobals
	"":                         {},
	"running":                  {},
	"waiting":                  {},
	"requested":                {},
	"Starting":                 {},
	"Stopping Cluster":         {},
	"Restoring":                {},
	"Running":                  {},
	"Starting Cluster":         {},
	"Point-in-time recovering": {},
}

// ListDatabaseClusterRestores List of the created database cluster restores on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusterRestores(ctx echo.Context, name string) error {
	req := ctx.Request()
	if err := validateRFC1035(name, "name"); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	val := url.Values{}
	val.Add("labelSelector", fmt.Sprintf("clusterName=%s", name))
	req.URL.RawQuery = val.Encode()
	path := req.URL.Path
	// trim restores
	path = strings.TrimSuffix(path, "/restores")
	// trim name
	path = strings.TrimSuffix(path, name)
	path = strings.ReplaceAll(path, "database-clusters", "database-cluster-restores")
	req.URL.Path = path
	return e.proxyKubernetes(ctx, "")
}

// CreateDatabaseClusterRestore Create a database cluster restore on the specified kubernetes cluster.
func (e *EverestServer) CreateDatabaseClusterRestore(ctx echo.Context) error {
	restore := &DatabaseClusterRestore{}
	if err := e.getBodyFromContext(ctx, restore); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not get DatabaseClusterRestore from the request body"),
		})
	}
	if err := validateDatabaseClusterRestore(ctx.Request().Context(), restore, e.kubeClient); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString(err.Error()),
		})
	}
	backups, err := e.kubeClient.ListDatabaseClusterRestores(ctx.Request().Context(), metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{
				"clusterName": restore.Spec.DbClusterName,
			},
		}),
	})
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString(err.Error()),
		})
	}
	for i := range backups.Items {
		backup := backups.Items[i]
		if _, ok := restoreStatuses[string(backup.Status.State)]; ok {
			return ctx.JSON(http.StatusBadRequest, Error{
				Message: pointer.ToString("Another restore is in progress"),
			})
		}
	}

	return e.proxyKubernetes(ctx, "")
}

// DeleteDatabaseClusterRestore Delete the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) DeleteDatabaseClusterRestore(ctx echo.Context, name string) error {
	return e.proxyKubernetes(ctx, name)
}

// GetDatabaseClusterRestore Returns the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseClusterRestore(ctx echo.Context, name string) error {
	return e.proxyKubernetes(ctx, name)
}

// UpdateDatabaseClusterRestore Replace the specified cluster restore on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseClusterRestore(ctx echo.Context, name string) error {
	restore := &DatabaseClusterRestore{}
	if err := e.getBodyFromContext(ctx, restore); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not get DatabaseClusterRestore from the request body"),
		})
	}
	if err := validateDatabaseClusterRestore(ctx.Request().Context(), restore, e.kubeClient); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString(err.Error()),
		})
	}
	return e.proxyKubernetes(ctx, name)
}
