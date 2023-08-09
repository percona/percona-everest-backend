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

	dbCluster := &everestv1alpha1.DatabaseCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: params.Name,
		},
	}

	return e.createResource(ctx, kubernetesID, params.Spec, &dbCluster.Spec, dbCluster)
}

// ListDatabaseClusters List of the created database clusters on the specified kubernetes cluster.
func (e *EverestServer) ListDatabaseClusters(ctx echo.Context, kubernetesID string) error {
	cl, statusCode, err := e.getK8sClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	dbClusterList, err := cl.ListDatabaseClusters(ctx.Request().Context())
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
	cl, statusCode, err := e.getK8sClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	if err := cl.DeleteDatabaseCluster(ctx.Request().Context(), name); err != nil {
		e.l.Error(err)
		return ctx.JSON(statusCode, Error{
			Message: pointer.ToString("Could not delete database cluster"),
		})
	}

	return ctx.NoContent(http.StatusOK)
}

// GetDatabaseCluster Get the specified database cluster on the specified kubernetes cluster.
func (e *EverestServer) GetDatabaseCluster(ctx echo.Context, kubernetesID string, name string) error {
	cl, statusCode, err := e.getK8sClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(statusCode, Error{Message: pointer.ToString(err.Error())})
	}

	dbCluster, err := cl.GetDatabaseCluster(ctx.Request().Context(), name)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(statusCode, Error{
			Message: pointer.ToString("Could not get database cluster"),
		})
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
	var params UpdateDatabaseClusterJSONRequestBody
	if err := ctx.Bind(&params); err != nil {
		return err
	}

	dbCluster := &everestv1alpha1.DatabaseCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	return e.updateResource(ctx, kubernetesID, name, params.Spec, &dbCluster.Spec, dbCluster)
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
