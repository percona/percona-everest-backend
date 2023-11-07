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
	"errors"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	corev1 "k8s.io/api/core/v1"

	"github.com/percona/percona-everest-backend/pkg/kubernetes"
)

// ListKubernetesClusters returns list of k8s clusters.
func (e *EverestServer) ListKubernetesClusters(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, []KubernetesCluster{
		{
			Id:        "id",
			Name:      "name",
			Namespace: "namespace",
			Uid:       "uid",
		},
	})
}

// RegisterKubernetesCluster registers a k8s cluster in Everest server.
func (e *EverestServer) RegisterKubernetesCluster(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, nil)
}

// GetKubernetesCluster Get the specified Kubernetes cluster.
func (e *EverestServer) GetKubernetesCluster(ctx echo.Context, _ string) error {
	result := KubernetesCluster{
		Id:        "id",
		Name:      "name",
		Namespace: "namespace",
		Uid:       "uid",
	}
	return ctx.JSON(http.StatusOK, result)
}

// UnregisterKubernetesCluster removes a Kubernetes cluster from Everest.
func (e *EverestServer) UnregisterKubernetesCluster(ctx echo.Context, _ string) error {
	var params UnregisterKubernetesClusterParams
	if err := ctx.Bind(&params); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(err.Error())})
	}

	if !params.Force {
		clusters, err := e.kubeClient.ListDatabaseClusters(ctx.Request().Context())
		if err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{
				Message: pointer.ToString("Could not list database clusters"),
			})
		}

		if len(clusters.Items) > 0 {
			return ctx.JSON(http.StatusBadRequest, Error{
				Message: pointer.ToString("Remove all database clusters before unregistering a Kubernetes cluster or use \"Force\" field to ignore this message"),
			})
		}
	}

	return ctx.NoContent(http.StatusOK)
}

// GetKubernetesClusterResources returns all and available resources of a Kubernetes cluster.
func (e *EverestServer) GetKubernetesClusterResources(ctx echo.Context, _ string) error {
	// Get cluster type
	clusterType, err := e.kubeClient.GetClusterType(ctx.Request().Context())
	if err != nil {
		e.l.Error(err)
		// Instead of failing we switch to a generic cluster type.
		clusterType = kubernetes.ClusterTypeGeneric
	}

	var volumes *corev1.PersistentVolumeList
	if clusterType == kubernetes.ClusterTypeEKS {
		volumes, err = e.kubeClient.GetPersistentVolumes(ctx.Request().Context())
		if err != nil {
			e.l.Error(err)
			return ctx.JSON(http.StatusInternalServerError, Error{
				Message: pointer.ToString("Could not get persistent volumes"),
			})
		}
	}

	res, err := e.calculateClusterResources(ctx, e.kubeClient, clusterType, volumes)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	return ctx.JSON(http.StatusOK, res)
}

// SetKubernetesClusterMonitoring enables or disables Kubernetes cluster monitoring.
func (e *EverestServer) SetKubernetesClusterMonitoring(ctx echo.Context, _ string) error {
	var params KubernetesClusterMonitoring
	if err := ctx.Bind(&params); err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusBadRequest, Error{
			Message: pointer.ToString("Could not parse request body"),
		})
	}

	if params.Enable {
		return e.enableK8sClusterMonitoring(ctx, params)
	}

	return e.disableK8sClusterMonitoring(ctx)
}

func (e *EverestServer) disableK8sClusterMonitoring(ctx echo.Context) error {
	if err := e.kubeClient.DeleteVMAgent(); err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not delete VMAgent"),
		})
	}

	return ctx.NoContent(http.StatusOK)
}

func (e *EverestServer) enableK8sClusterMonitoring(ctx echo.Context, params KubernetesClusterMonitoring) error {
	mc, err := e.kubeClient.GetMonitoringConfig(ctx.Request().Context(), params.MonitoringInstanceName)
	if err != nil {
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{
			Message: pointer.ToString("Could not create VMAgent in Kubernetes"),
		})
	}
	if err := e.kubeClient.DeployVMAgent(ctx.Request().Context(), mc.Spec.CredentialsSecretName, mc.Spec.PMM.URL); err != nil {
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
