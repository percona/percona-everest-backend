package kubernetes

import (
	"context"
	"strings"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/percona/percona-everest-backend/model"
)

type applyFunc func(secretName, namespace string) error

// CreateObjectStorage creates an ObjectStorage.
func (k *Kubernetes) CreateObjectStorage(ctx context.Context, namespace string, bs model.BackupStorage, secretData map[string]string) error {
	return k.createConfigWithSecret(ctx, bs.Name, namespace, secretData, func(secretName, namespace string) error {
		backupStorage := &everestv1alpha1.ObjectStorage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bs.Name,
				Namespace: namespace,
			},
			Spec: everestv1alpha1.ObjectStorageSpec{
				Type:                  everestv1alpha1.ObjectStorageType(bs.Type),
				Bucket:                bs.BucketName,
				Region:                bs.Region,
				EndpointURL:           bs.URL,
				CredentialsSecretName: secretName,
			},
		}

		return k.client.CreateObjectStorage(ctx, backupStorage)
	})
}

// DeleteObjectStorage deletes an ObjectStorage.
func (k *Kubernetes) DeleteObjectStorage(ctx context.Context, name, namespace string) error {
	dbClusters, err := k.getDBClustersByObjectStorage(ctx, name)
	if err != nil {
		return err
	}

	if err = buildObjectStorageInUseError(dbClusters, name); err != nil {
		return err
	}

	err = k.client.DeleteObjectStorage(ctx, name, namespace)
	if err != nil {
		return err
	}

	secretName := buildSecretName(name)
	return k.DeleteSecret(ctx, secretName, namespace)
}

// CreateConfigWithSecret creates a resource and the linked secret.
func (k *Kubernetes) createConfigWithSecret(ctx context.Context, configName, namespace string, secretData map[string]string, apply applyFunc) error {
	secretName := buildSecretName(configName)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		StringData: secretData,
		Type:       corev1.SecretTypeOpaque,
	}
	_, err := k.CreateSecret(ctx, secret)
	// if such Secret is already present in k8s - consider it as created and do nothing (fixme)
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return err
	}

	err = apply(secretName, namespace)
	// if such config is already present in k8s - consider it as created and do nothing (fixme)
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (k *Kubernetes) getDBClustersByObjectStorage(ctx context.Context, storageName string) ([]everestv1alpha1.DatabaseCluster, error) {
	list, err := k.ListDatabaseClusters(ctx)
	if err != nil {
		return nil, err
	}

	dbClusters := make([]everestv1alpha1.DatabaseCluster, 0, len(list.Items))
	for _, dbCluster := range list.Items {
		for _, schedule := range dbCluster.Spec.Backup.Schedules {
			if schedule.ObjectStorageName == storageName {
				dbClusters = append(dbClusters, dbCluster)
				break
			}
		}
	}

	return dbClusters, nil
}

func buildObjectStorageInUseError(dbClusters []everestv1alpha1.DatabaseCluster, storageName string) error {
	if len(dbClusters) == 0 {
		return nil
	}
	names := make([]string, 0, len(dbClusters))
	for _, cluster := range dbClusters {
		names = append(names, cluster.Name)
	}

	return errors.Errorf("the ObjectStorage '%s' is used in following DatabaseClusters: %s. Please update the DatabaseClusters configuration first", storageName, strings.Join(names, ","))
}

func buildSecretName(crName string) string {
	return crName + "-secret"
}
