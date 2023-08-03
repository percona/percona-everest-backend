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
package api

import (
	"context"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/go-logr/zapr"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/percona/percona-everest-backend/model"
	"github.com/percona/percona-everest-backend/pkg/kubernetes"
	"github.com/percona/percona-everest-backend/pkg/logger"
)

// ListKubernetesClusters returns list of k8s clusters.
func (e *EverestServer) ListKubernetesClusters(ctx echo.Context) error {
	list, err := e.storage.ListKubernetesClusters(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("Could not list Kubernetes clusters")})
	}

	result := make([]KubernetesCluster, 0, len(list))
	for _, k := range list {
		result = append(result, KubernetesCluster{
			Id:        k.ID,
			Name:      k.Name,
			Namespace: k.Namespace,
		})
	}

	return ctx.JSON(http.StatusOK, result)
}

// RegisterKubernetesCluster registers a k8s cluster in Everest server.
func (e *EverestServer) RegisterKubernetesCluster(ctx echo.Context) error {
	var params CreateKubernetesClusterParams
	if err := ctx.Bind(&params); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	c := ctx.Request().Context()

	_, err := clientcmd.BuildConfigFromKubeconfigGetter("", newConfigGetter(params.Kubeconfig).loadFromString)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString("Could not build kubeconfig")})
	}

	k, err := e.storage.CreateKubernetesCluster(c, model.CreateKubernetesClusterParams{
		Name:      params.Name,
		Namespace: params.Namespace,
	})
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code.Name() == pgErrUniqueViolation {
				return ctx.JSON(http.StatusBadRequest, Error{
					Message: pointer.ToString("Kubernetes cluster with the same name already exists. " + pgErr.Detail),
				})
			}
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("Could not create Kubernetes cluster")})
	}

	err = e.secretsStorage.CreateSecret(c, k.ID, params.Kubeconfig)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("Could not store kubeconfig in secrets storage")})
	}

	result := KubernetesCluster{
		Id:   k.ID,
		Name: k.Name,
	}
	return ctx.JSON(http.StatusOK, result)
}

// GetKubernetesCluster Get the specified Kubernetes cluster.
func (e *EverestServer) GetKubernetesCluster(ctx echo.Context, kubernetesID string) error {
	k, err := e.storage.GetKubernetesCluster(ctx.Request().Context(), kubernetesID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("Could not find Kubernetes cluster")})
	}
	result := KubernetesCluster{
		Id:        k.ID,
		Name:      k.Name,
		Namespace: k.Namespace,
	}
	return ctx.JSON(http.StatusOK, result)
}

// UnregisterKubernetesCluster removes a Kubernetes cluster from Everest.
func (e *EverestServer) UnregisterKubernetesCluster(ctx echo.Context, kubernetesID string) error {
	var params UnregisterKubernetesClusterParams
	if err := ctx.Bind(&params); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	k, err := e.storage.GetKubernetesCluster(ctx.Request().Context(), kubernetesID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("Could not find Kubernetes cluster")})
	}

	l := logger.MustInitLogger()
	client, err := kubernetes.NewFromSecretsStorage(
		ctx.Request().Context(), e.secretsStorage, k.ID,
		k.Namespace, zapr.NewLogger(l),
	)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not create Kubernetes client from kubeconfig"),
		})
	}

	if params.Force == nil || !*params.Force {
		clusters, err := client.ListDatabaseClusters(ctx.Request().Context())
		if err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{
				Message: pointer.ToString("Could not list database clusters"),
			})
		}

		if len(clusters.Items) != 0 {
			return ctx.JSON(http.StatusBadRequest, Error{
				Message: pointer.ToString("Remove all database clusters before unregistering a Kubernetes cluster or use \"Force\" field to ignore this message"),
			})
		}
	}

	if err := e.removeK8sCluster(ctx.Request().Context(), kubernetesID); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not remove Kubernetes cluster"),
		})
	}

	return ctx.NoContent(http.StatusOK)
}

func (e *EverestServer) removeK8sCluster(ctx context.Context, kubernetesID string) error {
	if _, err := e.secretsStorage.DeleteSecret(ctx, kubernetesID); err != nil {
		return errors.Wrap(err, "could not delete kubeconfig from secrets storage")
	}

	if err := e.storage.DeleteKubernetesCluster(ctx, kubernetesID); err != nil {
		return errors.Wrap(err, "could not delete Kubernetes cluster from db")
	}

	return nil
}
