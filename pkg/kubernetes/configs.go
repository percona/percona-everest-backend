package kubernetes

import (
	"context"
	"strings"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

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

	err = k.client.GetResource(ctx, name, &unstructured.Unstructured{}, &metav1.GetOptions{})
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

	return k.createConfigWithSecret(ctx, cfg.SecretName(), config, cfgSecrets)
}

// DeleteObjectStorage deletes an ObjectStorage.
func (k *Kubernetes) DeleteObjectStorage(ctx context.Context, name, secretName string, parentDBCluster string) error {
	dbClusters, err := k.getDBClustersByObjectStorage(ctx, name, parentDBCluster)
	if err != nil {
		return err
	}

	if err = buildObjectStorageInUseError(dbClusters, name); err != nil {
		return err
	}

	err = k.client.DeleteObjectStorage(ctx, name, k.namespace)
	if err != nil {
		return err
	}

	return k.DeleteSecret(ctx, secretName, k.namespace)
}

// GetObjectStorage returns the ObjectStorage.
func (k *Kubernetes) GetObjectStorage(ctx context.Context, name, namespace string) (*everestv1alpha1.ObjectStorage, error) {
	return k.client.GetObjectStorage(ctx, name, namespace)
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
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return err
	}

	err = k.client.CreateResource(ctx, cfg, &unstructured.Unstructured{}, &metav1.CreateOptions{})
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

func (k *Kubernetes) getDBClustersByObjectStorage(ctx context.Context, storageName, exceptCluster string) ([]everestv1alpha1.DatabaseCluster, error) {
	list, err := k.ListDatabaseClusters(ctx)
	if err != nil {
		return nil, err
	}

	dbClusters := make([]everestv1alpha1.DatabaseCluster, 0, len(list.Items))
	for _, dbCluster := range list.Items {
		for _, schedule := range dbCluster.Spec.Backup.Schedules {
			if schedule.ObjectStorageName == storageName && dbCluster.Name != exceptCluster {
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
