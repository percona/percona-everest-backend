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

// Package configs contains methods common to configs management.
package configs

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/percona/percona-everest-backend/model"
	"github.com/percona/percona-everest-backend/pkg/kubernetes"
)

type (
	getSecretFn func(ctx context.Context, id string) (string, error)
	isInUseFn   func(ctx context.Context, name string, kubeClient *kubernetes.Kubernetes) (bool, error)
)

type initKubeClientFn = func(ctx context.Context, kubernetesID string) (*model.KubernetesCluster, *kubernetes.Kubernetes, int, error)

// DeleteConfigFromK8sClusters deletes the provided config from all provided Kubernetes clusters.
// If an error occurs, it's logged via the provided logger, but no other action is taken.
func DeleteConfigFromK8sClusters(
	ctx context.Context,
	kubernetesClusters []model.KubernetesCluster,
	cfg kubernetes.ConfigK8sResourcer,
	initKubeClient initKubeClientFn,
	isInUse isInUseFn,
	l *zap.SugaredLogger,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// Delete configs in all k8s clusters
	for _, k := range kubernetesClusters {
		k := k
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, kubeClient, _, err := initKubeClient(ctx, k.ID)
			if err != nil {
				l.Error(errors.Wrap(err, "could not init kube client for config"))
				return
			}

			err = kubeClient.DeleteConfig(ctx, cfg, func(ctx context.Context, name string) (bool, error) {
				return isInUse(ctx, name, kubeClient)
			})
			if err != nil && !errors.Is(err, kubernetes.ErrConfigInUse) {
				l.Error(errors.Wrap(err, "could not delete config"))
				return
			}
		}()
	}

	wg.Wait()
}

// UpdateConfigInAllK8sClusters updates config resources in all the provided Kubernetes clusters.
func UpdateConfigInAllK8sClusters(
	ctx context.Context, kubernetesClusters []model.KubernetesCluster, cfg kubernetes.ConfigK8sResourcer,
	getSecret getSecretFn, initKubeClient initKubeClientFn, l *zap.SugaredLogger, wg *sync.WaitGroup,
) {
	defer wg.Done()
	// Update configs in all k8s clusters
	for _, k := range kubernetesClusters {
		k := k
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, kubeClient, _, err := initKubeClient(ctx, k.ID)
			if err != nil {
				l.Error(errors.Wrap(err, "could not init kube client to update config"))
				return
			}

			if err := kubeClient.UpdateConfig(ctx, cfg, getSecret); err != nil {
				l.Error(errors.Wrap(err, "could not update config"))
				return
			}
		}()
	}

	wg.Wait()
}
