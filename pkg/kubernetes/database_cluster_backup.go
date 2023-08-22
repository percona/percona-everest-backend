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

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDatabaseClusterBackup returns database cluster backup by name.
func (k *Kubernetes) GetDatabaseClusterBackup(ctx context.Context, name string) (*everestv1alpha1.DatabaseClusterBackup, error) { //nolint:unparam
	var backup *everestv1alpha1.DatabaseClusterBackup
	err := k.client.GetResource(ctx, name, backup, &metav1.GetOptions{})
	return backup, err
}
