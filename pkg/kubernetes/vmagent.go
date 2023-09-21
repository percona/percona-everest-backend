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

package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// VMAgentResourceName is the name of the VMAgent resource in k8s.
	VMAgentResourceName       = "everest-cluster-monitoring"
	vmAgentUsernameSecretName = "everest-cluster-vmagent-username"
)

// DeployVMAgent deploys a default VMAgent used by Everest.
func (k *Kubernetes) DeployVMAgent(ctx context.Context, secretName, monitoringURL string) error {
	k.l.Debug("Creating VMAgent username secret")
	_, err := k.CreateSecret(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vmAgentUsernameSecretName,
			Namespace: k.namespace,
		},
		StringData: map[string]string{
			"username": "api_key",
		},
	})

	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return errors.Join(err, errors.New("could not create VMAgent username secret"))
	}

	k.l.Debug("Applying VMAgent spec")
	vmagent, err := vmAgentSpec(k.namespace, secretName, monitoringURL)
	if err != nil {
		return errors.Join(err, errors.New("cannot generate VMAgent spec"))
	}

	err = k.client.ApplyObject(vmagent)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return errors.Join(err, errors.New("cannot apply VMAgent spec"))
	}
	k.l.Debug("VMAgent spec has been applied")

	return nil
}

// DeleteVMAgent deletes the default VMAgent as installed by Everest.
func (k *Kubernetes) DeleteVMAgent() error {
	vmagent, err := vmAgentSpec(k.namespace, "", "")
	if err != nil {
		return errors.Join(err, errors.New("cannot generate VMAgent spec"))
	}

	err = k.client.DeleteObject(vmagent)
	if err != nil {
		return errors.Join(err, errors.New("cannot delete VMAgent"))
	}

	return nil
}

// ListVMAgents returns list of VMAgents.
func (k *Kubernetes) ListVMAgents() (*unstructured.UnstructuredList, error) {
	vmAgents := &unstructured.UnstructuredList{}
	err := k.client.ListObjects(schema.FromAPIVersionAndKind("operator.victoriametrics.com/v1beta1", "VMAgent"), vmAgents)
	return vmAgents, err
}

// GetVMAgent returns VMAgent by name.
func (k *Kubernetes) GetVMAgent(name string) (*unstructured.Unstructured, error) {
	vmAgent := &unstructured.Unstructured{}
	err := k.client.GetObject(
		schema.FromAPIVersionAndKind("operator.victoriametrics.com/v1beta1", "VMAgent"), name, vmAgent,
	)
	return vmAgent, err
}

const specVMAgent = `
{
	"kind": "VMAgent",
	"apiVersion": "operator.victoriametrics.com/v1beta1",
	"metadata": {
		"name": %[4]s,
		"namespace": %[3]s,
		"creationTimestamp": null,
		"labels": {
			"app.kubernetes.io/managed-by": "everest",
			"everest.percona.com/type": "monitoring"
		}
	},
	"spec": {
		"image": {},
		"replicaCount": 1,
		"resources": {
			"limits": {
				"cpu": "500m",
				"memory": "850Mi"
			},
			"requests": {
				"cpu": "250m",
				"memory": "350Mi"
			}
		},
		"remoteWrite": [
			{
				"url": %[2]s,
				"basicAuth": {
					"username": {
						"name": %[5]s,
						"key": "username"
					},
					"password": {
						"name": %[1]s,
						"key": "apiKey"
					}
				},
				"tlsConfig": {
					"ca": {},
					"cert": {},
					"insecureSkipVerify": true
				}
			}
		],
		"selectAllByDefault": true,
		"serviceScrapeSelector": {},
		"serviceScrapeNamespaceSelector": {},
		"podScrapeSelector": {},
		"podScrapeNamespaceSelector": {},
		"probeSelector": {},
		"probeNamespaceSelector": {},
		"staticScrapeSelector": {},
		"staticScrapeNamespaceSelector": {},
		"arbitraryFSAccessThroughSMs": {},
		"extraArgs": {
			"memory.allowedPercent": "40"
		}
	}
}`

func vmAgentSpec(namespace, secretName, address string) (runtime.Object, error) { //nolint:ireturn
	jName, err := json.Marshal(VMAgentResourceName)
	if err != nil {
		return nil, err
	}

	jSecret, err := json.Marshal(secretName)
	if err != nil {
		return nil, err
	}

	jAddress, err := json.Marshal(address + "/victoriametrics/api/v1/write")
	if err != nil {
		return nil, err
	}

	jNamespace, err := json.Marshal(namespace)
	if err != nil {
		return nil, err
	}

	jUser, err := json.Marshal(vmAgentUsernameSecretName)
	if err != nil {
		return nil, err
	}

	manifest := fmt.Sprintf(specVMAgent, jSecret, jAddress, jNamespace, jName, jUser)

	o, _, err := unstructured.UnstructuredJSONScheme.Decode([]byte(manifest), nil, nil)
	if err != nil {
		return nil, err
	}

	return o, nil
}
