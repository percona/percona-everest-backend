package kubernetes

import (
	"context"
	"fmt"
	"strings"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/percona/percona-everest-backend/model"
)

type applyFunc func(secretName, namespace string) error

// ConfigK8sResourcer defines interface for config structs which support storage in Kubernetes.
// The struct is representeed in Kubernetes by:
//   - The structure itself as a resource
//   - Related secret identified by its name in the structure
type ConfigK8sResourcer interface {
	// K8sResource returns a resource which shall be created when storing this struct in Kubernetes.
	K8sResource(namespace string) (runtime.Object, error)
	// Secrets returns all monitoring instance secrets from secrets storage.
	Secrets(ctx context.Context, getSecret func(ctx context.Context, id string) (string, error)) (map[string]string, error)
	// SecretName returns the name of the k8s secret as referenced by the k8s MonitoringConfig resource.
	SecretName() string
}

// EnsureConfigExists makes sure a config resource for the provided object
// exists in Kubernetes. If it does not, it is created.
func (k *Kubernetes) EnsureConfigExists(
	ctx context.Context, cfg ConfigK8sResourcer,
	getSecret func(ctx context.Context, id string) (string, error),
) error {
	config, err := cfg.K8sResource(k.namespace)
	if err != nil {
		return errors.Wrap(err, "could not get Kubernetes resource object")
	}

	acc := meta.NewAccessor()
	name, err := acc.Name(config)
	if err != nil {
		return errors.Wrap(err, "could not get name from a config object")
	}

	r, err := cfg.K8sResource(k.namespace)
	if err != nil {
		return errors.Wrap(err, "could not get Kubernetes resource object")
	}

	err = k.client.GetResource(ctx, name, r, &metav1.GetOptions{})
	if err == nil {
		return nil
	}

	if !k8serrors.IsNotFound(err) {
		return errors.Wrap(err, "could not get config from kubernetes")
	}

	cfgSecrets, err := cfg.Secrets(ctx, getSecret)
	if err != nil {
		return errors.Wrap(err, "could not get config secrets from secrets storage")
	}

	err = k.createConfigWithSecret(ctx, cfg.SecretName(), config, cfgSecrets)
	if err != nil {
		return errors.Wrap(err, "could not create a config with secret")
	}

	return nil
}

// UpdateBackupStorage creates a BackupStorage.
func (k *Kubernetes) UpdateBackupStorage(ctx context.Context, namespace string, bs model.BackupStorage, secretData map[string]string) error {
	storage, err := k.client.GetBackupStorage(ctx, bs.Name, namespace)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "Failed to get BackupStorage %s", bs.Name)
	}

	return k.updateConfigWithSecret(ctx, bs.SecretName(), namespace, secretData, func(secretName, namespace string) error {
		storage.Spec.Type = everestv1alpha1.BackupStorageType(bs.Type)
		storage.Spec.Bucket = bs.BucketName
		storage.Spec.Region = bs.Region
		storage.Spec.EndpointURL = bs.URL

		return k.client.UpdateBackupStorage(ctx, storage)
	})
}

// DeleteBackupStorage deletes an BackupStorage.
func (k *Kubernetes) DeleteBackupStorage(ctx context.Context, name, secretName string, parentDBCluster string) error {
	dbClusters, err := k.getDBClustersByBackupStorage(ctx, name, parentDBCluster)
	if err != nil {
		return err
	}

	if err = buildBackupStorageInUseError(dbClusters, name); err != nil {
		return err
	}

	err = k.client.DeleteBackupStorage(ctx, name, k.namespace)
	if err != nil {
		return err
	}

	return k.DeleteSecret(ctx, secretName, k.namespace)
}

// GetBackupStorage returns the BackupStorage.
func (k *Kubernetes) GetBackupStorage(ctx context.Context, name, namespace string) (*everestv1alpha1.BackupStorage, error) {
	return k.client.GetBackupStorage(ctx, name, namespace)
}

// CreateConfigWithSecret creates a resource and the linked secret.
func (k *Kubernetes) createConfigWithSecret(ctx context.Context, secretName string, cfg runtime.Object, secretData map[string]string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: k.namespace,
		},
		StringData: secretData,
		Type:       corev1.SecretTypeOpaque,
	}
	_, err := k.CreateSecret(ctx, secret)
	// if such Secret is already present in k8s - consider it as created and do nothing (fixme)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}

	err = k.client.CreateResource(ctx, cfg, &metav1.CreateOptions{})
	// if such config is already present in k8s - consider it as created and do nothing (fixme)
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			// rollback the changes
			_ = k.DeleteSecret(ctx, secret.Name, secret.Namespace)
			return err
		}
	}

	return nil
}

// updateConfigWithSecret creates a resource and the linked secret.
func (k *Kubernetes) updateConfigWithSecret(ctx context.Context, secretName, namespace string, secretData map[string]string, apply applyFunc) error {
	oldSecret, err := k.GetSecret(ctx, secretName, namespace)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to read secret %s", secretName))
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		StringData: secretData,
		Type:       corev1.SecretTypeOpaque,
	}
	_, err = k.UpdateSecret(ctx, secret)
	if err != nil {
		return err
	}

	err = apply(secretName, namespace)
	if err != nil {
		// rollback the changes
		_, _ = k.UpdateSecret(ctx, oldSecret)
		return err
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
