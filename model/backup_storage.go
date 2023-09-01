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

package model

import (
	"context"
	"fmt"
	"time"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// BackupStorage represents db model for BackupStorage.
type BackupStorage struct {
	Type        string
	Name        string `gorm:"primaryKey"`
	Description string
	BucketName  string
	URL         string
	Region      string
	AccessKeyID string
	SecretKeyID string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// SecretName returns the name of the k8s secret as referenced by the k8s MonitoringConfig resource.
func (b *BackupStorage) SecretName() string {
	return fmt.Sprintf("%s-secret", b.Name)
}

// Secrets returns all monitoring instance secrets from secrets storage.
func (b *BackupStorage) Secrets(ctx context.Context, getSecret func(ctx context.Context, id string) (string, error)) (map[string]string, error) {
	secretKey, err := getSecret(ctx, b.SecretKeyID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get secretKey")
	}
	accessKey, err := getSecret(ctx, b.AccessKeyID)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get accessKey")
	}
	return map[string]string{
		"AWS_SECRET_ACCESS_KEY": secretKey,
		"AWS_ACCESS_KEY_ID":     accessKey,
	}, nil
}

// K8sResource returns a resource which shall be created when storing this struct in Kubernetes.
func (b *BackupStorage) K8sResource(namespace string) (runtime.Object, error) { //nolint:unparam,ireturn
	bs := &everestv1alpha1.BackupStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name,
			Namespace: namespace,
		},
		Spec: everestv1alpha1.BackupStorageSpec{
			Type:                  everestv1alpha1.BackupStorageType(b.Type),
			Bucket:                b.BucketName,
			Region:                b.Region,
			EndpointURL:           b.URL,
			CredentialsSecretName: b.SecretName(),
		},
	}

	return bs, nil
}
