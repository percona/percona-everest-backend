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
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/go-logr/zapr"
	"github.com/labstack/echo/v4"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/percona/percona-everest-backend/client"
	"github.com/percona/percona-everest-backend/pkg/kubernetes"
	"github.com/percona/percona-everest-backend/pkg/logger"
)

// CreateDatabaseCluster creates a new db cluster inside the given k8s cluster.
func (e *EverestServer) CreateDatabaseCluster(ctx echo.Context, kubernetesID string) error {
	var params CreateDatabaseClusterJSONRequestBody
	if err := ctx.Bind(&params); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	return e.createOrUpdateDBCluster(ctx, kubernetesID, params.Name, params.Spec, "")
}

// ListDatabaseClusters List of the created database clusters on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusters(ctx echo.Context, kubernetesID string) error {
	dbClusterList := &everestv1alpha1.DatabaseClusterList{}
	statusCode, err := e.doK8sRequest(
		ctx.Request().Context(), ctx.Request().URL, ctx.Request().Method,
		kubernetesID, "", struct{}{}, dbClusterList,
	)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	items := make([]client.DatabaseClusterWithName, 0, len(dbClusterList.Items))
	res := &client.DatabaseClusterList{Items: &items}
	for _, i := range dbClusterList.Items {
		i := i
		d, err := e.parseDBClusterObj(&i)
		if err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
		}
		*res.Items = append(*res.Items, *d)
	}

	return ctx.JSON(http.StatusOK, res)
}

// DeleteDatabaseCluster Create a database cluster on the specified kubernetes cluster.
func (e *EverestServer) DeleteDatabaseCluster(ctx echo.Context, kubernetesID string, name string) error {
	statusCode, err := e.doK8sRequest(
		ctx.Request().Context(), ctx.Request().URL, ctx.Request().Method,
		kubernetesID, name, struct{}{}, nil,
	)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.NoContent(http.StatusOK)
}

// GetDatabaseCluster Get the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseCluster(ctx echo.Context, kubernetesID string, name string) error {
	dbCluster := &everestv1alpha1.DatabaseCluster{}
	statusCode, err := e.doK8sRequest(
		ctx.Request().Context(), ctx.Request().URL, ctx.Request().Method,
		kubernetesID, name, struct{}{}, dbCluster,
	)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	d, err := e.parseDBClusterObj(dbCluster)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, d)
}

// UpdateDatabaseCluster Replace the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) UpdateDatabaseCluster(ctx echo.Context, kubernetesID string, name string) error {
	dbCluster := &everestv1alpha1.DatabaseCluster{}
	statusCode, err := e.doK8sRequest(
		ctx.Request().Context(), ctx.Request().URL, http.MethodGet,
		kubernetesID, name, struct{}{}, dbCluster,
	)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	var params UpdateDatabaseClusterJSONRequestBody
	if err := ctx.Bind(&params); err != nil {
		return err
	}

	return e.createOrUpdateDBCluster(ctx, kubernetesID, name, params.Spec, dbCluster.ResourceVersion)
}

func (e *EverestServer) parseDBClusterObj(dbCluster *everestv1alpha1.DatabaseCluster) (*client.DatabaseClusterWithName, error) {
	d := &client.DatabaseClusterWithName{
		Name: dbCluster.Name,
	}

	if err := e.assignFieldBetweenStructs(dbCluster.Spec, &d.Spec); err != nil {
		return nil, errors.New("Could not parse database cluster spec")
	}
	if err := e.assignFieldBetweenStructs(dbCluster.Status, &d.Status); err != nil {
		return nil, errors.New("Could not parse database cluster status")
	}

	return d, nil
}

func (e *EverestServer) createOrUpdateDBCluster(
	ctx echo.Context, kubernetesID, resourceName string, spec any, resourceVersion string,
) error {
	dbCluster := &everestv1alpha1.DatabaseCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "everest.percona.com/v1alpha1",
			Kind:       "DatabaseCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            resourceName,
			ResourceVersion: resourceVersion,
		},
	}

	if err := e.assignFieldBetweenStructs(spec, &dbCluster.Spec); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	name := ""
	if resourceVersion != "" {
		name = resourceName
	}

	statusCode, err := e.doK8sRequest(
		ctx.Request().Context(), ctx.Request().URL, ctx.Request().Method,
		kubernetesID, name, dbCluster, nil,
	)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.NoContent(http.StatusOK)
}

// GetDatabaseClusterCredentials returns credentials for the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseClusterCredentials(ctx echo.Context, kubernetesID string, name string) error {
	cluster, err := e.storage.GetKubernetesCluster(ctx.Request().Context(), kubernetesID)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	l := logger.MustInitLogger()
	kubeClient, err := kubernetes.NewFromSecretsStorage(ctx.Request().Context(), e.secretsStorage, cluster.ID, cluster.Namespace, zapr.NewLogger(l))
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	databaseCluster, err := kubeClient.GetDatabaseCluster(ctx.Request().Context(), name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	secret, err := kubeClient.GetSecret(ctx.Request().Context(), databaseCluster.Spec.Engine.UserSecretsName, cluster.Namespace)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}
	response := &DatabaseClusterCredential{}
	switch databaseCluster.Spec.Engine.Type {
	case everestv1alpha1.DatabaseEnginePXC:
		response.Username = pointer.ToString("root")
		response.Password = pointer.ToString(string(secret.Data["root"]))
	case everestv1alpha1.DatabaseEnginePSMDB:
		response.Username = pointer.ToString(string(secret.Data["MONGODB_USER_ADMIN_USER"]))
		response.Password = pointer.ToString(string(secret.Data["MONGODB_USER_ADMIN_PASSWORD"]))
	case everestv1alpha1.DatabaseEnginePostgresql:
		response.Username = pointer.ToString("postgres")
		response.Password = pointer.ToString(string(secret.Data["password"]))
	default:
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("Unsupported database engine")})
	}

	return ctx.JSON(http.StatusOK, response)
}
