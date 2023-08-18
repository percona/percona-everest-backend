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

type secretNameGetter interface {
	// SecretName returns the name of the k8s secret as referenced by the k8s config resource.
	SecretName() string
}

type k8sClusterLister interface {
	ListKubernetesClusters(ctx context.Context) ([]model.KubernetesCluster, error)
}

type initKubeClientFn = func(ctx context.Context, kubernetesID string) (*model.KubernetesCluster, *kubernetes.Kubernetes, int, error)

// DeleteConfigFromAllK8sClusters deletes the provided config from all Kubernetes clusters.
// If an error occurs, it's logged via the provided logger, but no other action is taken.
func DeleteConfigFromAllK8sClusters(
	ctx context.Context,
	cfg kubernetes.ConfigK8sResourcer,
	getSecret getSecretFn,
	k8s k8sClusterLister,
	initKubeClient initKubeClientFn,
	isInUse isInUseFn,
	l *zap.SugaredLogger,
) error {
	ks, err := k8s.ListKubernetesClusters(ctx)
	if err != nil {
		return errors.Wrap(err, "could not list Kubernetes clusters")
	}

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg := &sync.WaitGroup{}
		errs := make(chan error)

		// Delete configs in all k8s clusters
		for _, k := range ks {
			k := k
			wg.Add(1)
			go func() {
				defer wg.Done()

				_, kubeClient, _, err := initKubeClient(ctx, k.ID)
				if err != nil {
					errs <- errors.Wrap(err, "could not init kube client for config")
					return
				}

				err = kubeClient.DeleteConfig(ctx, cfg, func(ctx context.Context, name string) (bool, error) {
					return isInUse(ctx, name, kubeClient)
				})
				if err != nil {
					errs <- errors.Wrap(err, "could not delete config")
					return
				}
			}()
		}

		// Log all errors
		go func() {
			for {
				select {
				case err := <-errs:
					l.Error(err)
				case <-ctx.Done():
					return
				}
			}
		}()
		wg.Wait()
	}()

	return nil
}

func UpdateConfigInAllK8sClusters(
	ctx context.Context, cfg kubernetes.ConfigK8sResourcer, getSecret getSecretFn, k8s k8sClusterLister,
	initKubeClient initKubeClientFn, l *zap.SugaredLogger,
) error {
	ks, err := k8s.ListKubernetesClusters(ctx)
	if err != nil {
		return errors.Wrap(err, "could not list Kubernetes clusters")
	}

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg := &sync.WaitGroup{}
		errs := make(chan error)

		// Update configs in all k8s clusters
		for _, k := range ks {
			k := k
			wg.Add(1)
			go func() {
				defer wg.Done()

				_, kubeClient, _, err := initKubeClient(ctx, k.ID)
				if err != nil {
					errs <- errors.Wrap(err, "could not init kube client to update config")
					return
				}

				if err := kubeClient.UpdateConfig(ctx, cfg, getSecret); err != nil {
					errs <- errors.Wrap(err, "could not update config")
					return
				}
			}()
		}

		// Log all errors
		go func() {
			for {
				select {
				case err := <-errs:
					l.Error(err)
				case <-ctx.Done():
					return
				}
			}
		}()
		wg.Wait()
	}()

	return nil
}
