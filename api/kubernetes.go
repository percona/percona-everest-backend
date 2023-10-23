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
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/percona/percona-everest-backend/model"
	"github.com/percona/percona-everest-backend/pkg/kubernetes"
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
			Uid:       k.UID,
		})
	}

	return ctx.JSON(http.StatusOK, result)
}

// RegisterKubernetesCluster registers a k8s cluster in Everest server.
func (e *EverestServer) RegisterKubernetesCluster(ctx echo.Context) error {
	list, err := e.storage.ListKubernetesClusters(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("Could not list Kubernetes clusters")})
	}
	if len(list) != 0 {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString("Everest does not support multiple kubernetes clusters right now. Please delete the existing cluster before registering a new one")})
	}
	var params CreateKubernetesClusterParams
	if err := ctx.Bind(&params); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}
	c := ctx.Request().Context()

	_, err = clientcmd.BuildConfigFromKubeconfigGetter("", newConfigGetter(params.Kubeconfig).loadFromString)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not build kubeconfig"),
		})
	}

	ns, err := e.getNamespace(ctx.Request().Context(), params)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString(err.Error()),
		})
	}

	k, err := e.storage.CreateKubernetesCluster(c, model.CreateKubernetesClusterParams{
		Name:      params.Name,
		Namespace: params.Namespace,
		UID:       string(ns.UID),
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

	err = e.secretsStorage.PutSecret(c, k.ID, params.Kubeconfig)
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
		Uid:       k.UID,
	}
	return ctx.JSON(http.StatusOK, result)
}

// UnregisterKubernetesCluster removes a Kubernetes cluster from Everest.
func (e *EverestServer) UnregisterKubernetesCluster(ctx echo.Context, kubernetesID string) error {
	var params UnregisterKubernetesClusterParams
	if err := ctx.Bind(&params); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	var kubeClient *kubernetes.Kubernetes
	var code int
	var err error
	if !params.Force {
		_, kubeClient, code, err = e.initKubeClient(ctx.Request().Context(), kubernetesID)
		if err != nil && !params.IgnoreKubernetesUnavailable {
			return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
		}
	}

	if kubeClient != nil && !params.Force {
		clusters, err := kubeClient.ListDatabaseClusters(ctx.Request().Context())
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
	if err := e.secretsStorage.DeleteSecret(ctx, kubernetesID); err != nil {
		return errors.Join(err, errors.New("could not delete kubeconfig from secrets storage"))
	}

	if err := e.storage.DeleteKubernetesCluster(ctx, kubernetesID); err != nil {
		return errors.Join(err, errors.New("could not delete Kubernetes cluster from db"))
	}

	return nil
}

// GetKubernetesClusterResources returns all and available resources of a Kubernetes cluster.
func (e *EverestServer) GetKubernetesClusterResources(ctx echo.Context, kubernetesID string) error {
	_, kubeClient, code, err := e.initKubeClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}

	// Get cluster type
	clusterType, err := kubeClient.GetClusterType(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		// Instead of failing we switch to a generic cluster type.
		clusterType = kubernetes.ClusterTypeGeneric
	}

	var volumes *corev1.PersistentVolumeList
	if clusterType == kubernetes.ClusterTypeEKS {
		volumes, err = kubeClient.GetPersistentVolumes(ctx.Request().Context())
		if err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{
				Message: pointer.ToString("Could not get persistent volumes"),
			})
		}
	}

	res, err := e.calculateClusterResources(ctx, kubeClient, clusterType, volumes)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, res)
}

// SetKubernetesClusterMonitoring enables or disables Kubernetes cluster monitoring.
func (e *EverestServer) SetKubernetesClusterMonitoring(ctx echo.Context, kubernetesID string) error {
	var params KubernetesClusterMonitoring
	if err := ctx.Bind(&params); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not parse request body"),
		})
	}

	_, kubeClient, code, err := e.initKubeClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}

	if params.Enable {
		return e.enableK8sClusterMonitoring(ctx, params, kubeClient)
	}

	return e.disableK8sClusterMonitoring(ctx, kubeClient)
}

func (e *EverestServer) disableK8sClusterMonitoring(ctx echo.Context, kubeClient *kubernetes.Kubernetes) error {
	vmAgent, err := kubeClient.GetVMAgent(kubernetes.VMAgentResourceName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Nothing to disable
			return ctx.NoContent(http.StatusOK)
		}

		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get VMAgent from Kubernetes"),
		})
	}

	if err := kubeClient.DeleteVMAgent(); err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not delete VMAgent"),
		})
	}

	for _, s := range kubeClient.SecretNamesFromVMAgent(vmAgent) {
		mcs, err := kubeClient.GetMonitoringConfigsBySecretName(ctx.Request().Context(), s)
		if err != nil {
			err = errors.Join(err, fmt.Errorf("could not list monitoring configs by secret name %s", s))
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
		}

		for _, mc := range mcs {
			err = kubeClient.DeleteMonitoringConfig(ctx.Request().Context(), mc.Name, mc.Spec.CredentialsSecretName)
			if err != nil && !errors.Is(err, kubernetes.ErrMonitoringConfigInUse) {
				err = errors.Join(err, fmt.Errorf("could not delete monitoring config %s from Kubernetes", mc.Name))
				e.l.Error(err)
				return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
			}
		}
	}

	return ctx.NoContent(http.StatusOK)
}

func (e *EverestServer) enableK8sClusterMonitoring(ctx echo.Context, params KubernetesClusterMonitoring, kubeClient *kubernetes.Kubernetes) error {
	mi, err := e.storage.GetMonitoringInstance(params.MonitoringInstanceName)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not get monitoring instance"),
		})
	}

	if mi == nil {
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not find the provided monitoring instance by name"),
		})
	}

	if err := kubeClient.EnsureConfigExists(ctx.Request().Context(), mi, e.secretsStorage.GetSecret); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not make sure monitoring config exists in Kubernetes"),
		})
	}

	if err := kubeClient.DeployVMAgent(ctx.Request().Context(), mi.SecretName(), mi.URL); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not create VMAgent in Kubernetes"),
		})
	}

	return ctx.NoContent(http.StatusOK)
}

func (e *EverestServer) calculateClusterResources(
	ctx echo.Context, kubeClient *kubernetes.Kubernetes, clusterType kubernetes.ClusterType,
	volumes *corev1.PersistentVolumeList,
) (*KubernetesClusterResources, error) {
	allCPUMillis, allMemoryBytes, allDiskBytes, err := kubeClient.GetAllClusterResources(
		ctx.Request().Context(), clusterType, volumes,
	)
	if err != nil {
		e.l.Error(err)
		return nil, errors.New("could not get cluster resources")
	}

	consumedCPUMillis, consumedMemoryBytes, err := kubeClient.GetConsumedCPUAndMemory(ctx.Request().Context(), "")
	if err != nil {
		e.l.Error(err)
		return nil, errors.New("could not get consumed cpu and memory")
	}

	consumedDiskBytes, err := kubeClient.GetConsumedDiskBytes(ctx.Request().Context(), clusterType, volumes)
	if err != nil {
		e.l.Error(err)
		return nil, errors.New("could not get consumed disk bytes")
	}

	availableCPUMillis := allCPUMillis - consumedCPUMillis
	// handle underflow
	if availableCPUMillis > allCPUMillis {
		availableCPUMillis = 0
	}
	availableMemoryBytes := allMemoryBytes - consumedMemoryBytes
	// handle underflow
	if availableMemoryBytes > allMemoryBytes {
		availableMemoryBytes = 0
	}
	availableDiskBytes := allDiskBytes - consumedDiskBytes
	// handle underflow
	if availableDiskBytes > allDiskBytes {
		availableDiskBytes = 0
	}

	res := &KubernetesClusterResources{
		Capacity: ResourcesCapacity{
			CpuMillis:   pointer.ToUint64OrNil(allCPUMillis),
			MemoryBytes: pointer.ToUint64OrNil(allMemoryBytes),
			DiskSize:    pointer.ToUint64OrNil(allDiskBytes),
		},
		Available: ResourcesAvailable{
			CpuMillis:   pointer.ToUint64OrNil(availableCPUMillis),
			MemoryBytes: pointer.ToUint64OrNil(availableMemoryBytes),
			DiskSize:    pointer.ToUint64OrNil(availableDiskBytes),
		},
	}

	return res, nil
}

func (e *EverestServer) getNamespace(ctx context.Context, params CreateKubernetesClusterParams) (*corev1.Namespace, error) {
	kubeconfig, err := base64.StdEncoding.DecodeString(params.Kubeconfig)
	if err != nil {
		e.l.Error(err)
		return nil, errors.New("could not decode kubeconfig")
	}

	kubeClient, err := kubernetes.New(kubeconfig, *params.Namespace, e.l)
	if err != nil {
		e.l.Error(err)
		return nil, errors.New("could not create kube client")
	}

	ns, err := kubeClient.GetNamespace(ctx, *params.Namespace)
	if err != nil {
		e.l.Error(err)
		return nil, errors.New("could not get namespace from Kubernetes")
	}

	return ns, nil
}
