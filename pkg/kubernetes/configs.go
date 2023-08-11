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

// CreateBackupStorage creates an BackupStorage.
func (k *Kubernetes) CreateBackupStorage(ctx context.Context, namespace string, bs model.BackupStorage, secretData map[string]string) error {
	return k.createConfigWithSecret(ctx, bs.Name, namespace, secretData, func(secretName, namespace string) error {
		backupStorage := &everestv1alpha1.BackupStorage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bs.Name,
				Namespace: namespace,
			},
			Spec: everestv1alpha1.BackupStorageSpec{
				Type:                  everestv1alpha1.BackupStorageType(bs.Type),
				Bucket:                bs.BucketName,
				Region:                bs.Region,
				EndpointURL:           bs.URL,
				CredentialsSecretName: secretName,
			},
		}

		return k.client.CreateBackupStorage(ctx, backupStorage)
	})
}

// DeleteBackupStorage deletes an BackupStorage.
func (k *Kubernetes) DeleteBackupStorage(ctx context.Context, name, namespace string, parentDBCluster string) error {
	dbClusters, err := k.getDBClustersByBackupStorage(ctx, name, parentDBCluster)
	if err != nil {
		return err
	}

	if err = buildBackupStorageInUseError(dbClusters, name); err != nil {
		return err
	}

	err = k.client.DeleteBackupStorage(ctx, name, namespace)
	if err != nil {
		return err
	}

	secretName := buildSecretName(name)
	return k.DeleteSecret(ctx, secretName, namespace)
}

// GetBackupStorage returns the BackupStorage.
func (k *Kubernetes) GetBackupStorage(ctx context.Context, name, namespace string) (*everestv1alpha1.BackupStorage, error) {
	return k.client.GetBackupStorage(ctx, name, namespace)
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
	if err != nil {
		if !k8serr.IsAlreadyExists(err) {
			// rollback the changes
			_ = k.DeleteSecret(ctx, secret.Name, secret.Namespace)
			return err
		}
	}

	return nil
}

func (k *Kubernetes) getDBClustersByBackupStorage(ctx context.Context, storageName, exceptCluster string) ([]everestv1alpha1.DatabaseCluster, error) {
	list, err := k.ListDatabaseClusters(ctx)
	if err != nil {
		return nil, err
	}

	dbClusters := make([]everestv1alpha1.DatabaseCluster, 0, len(list.Items))
	for _, dbCluster := range list.Items {
		for _, schedule := range dbCluster.Spec.Backup.Schedules {
			if schedule.BackupStorageName == storageName && dbCluster.Name != exceptCluster {
				dbClusters = append(dbClusters, dbCluster)
				break
			}
		}
	}

	return dbClusters, nil
}

func buildBackupStorageInUseError(dbClusters []everestv1alpha1.DatabaseCluster, storageName string) error {
	if len(dbClusters) == 0 {
		return nil
	}
	names := make([]string, 0, len(dbClusters))
	for _, cluster := range dbClusters {
		names = append(names, cluster.Name)
	}

	return errors.Errorf("the BackupStorage '%s' is used in following DatabaseClusters: %s. Please update the DatabaseClusters configuration first", storageName, strings.Join(names, ","))
}

func buildSecretName(crName string) string {
	return crName + "-secret"
}
