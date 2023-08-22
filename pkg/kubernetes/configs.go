package kubernetes

import (
	"context"
	"fmt"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type isInUseFn func(ctx context.Context, name string) (bool, error)

// ConfigK8sResourcer defines interface for config structs which support storage in Kubernetes.
// The struct is representeed in Kubernetes by:
//   - The structure itself as a resource
//   - Related secret identified by its name in the structure
type ConfigK8sResourcer interface {
	// K8sResource returns a resource which shall be created when storing this struct in Kubernetes.
	K8sResource(namespace string) (runtime.Object, error)
	// Secrets returns all monitoring instance secrets from secrets storage.
	Secrets(ctx context.Context, getSecret func(ctx context.Context, id string) (string, error)) (map[string]string, error)
	// SecretName returns the name of the k8s secret as referenced by the k8s config resource.
	SecretName() string
}

// ErrConfigInUse is returned when a config is in use.
var ErrConfigInUse error = errors.New("config is in use")

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
		return errors.Wrap(err, "could not get config from Kubernetes")
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

// UpdateConfig updates the config resources based on the provided config object.
func (k *Kubernetes) UpdateConfig(
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
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}

		return errors.Wrap(err, "could not get config resource from Kubernetes")
	}

	cfgSecrets, err := cfg.Secrets(ctx, getSecret)
	if err != nil {
		return errors.Wrap(err, "could not get config secrets from secrets storage")
	}

	err = k.updateConfigWithSecret(ctx, cfg.SecretName(), config, cfgSecrets)
	if err != nil {
		return errors.Wrap(err, "could not update config with secrets in Kubernetes")
	}

	return nil
}

// DeleteConfig deletes the config and secret resources from k8s
// for the provided config object.
// If the config is in use, ErrConfigInUse is returned.
func (k *Kubernetes) DeleteConfig(
	ctx context.Context, cfg ConfigK8sResourcer, isInUse isInUseFn,
) error {
	k.l.Debugf("Starting to delete config")

	config, err := cfg.K8sResource(k.namespace)
	if err != nil {
		return errors.Wrap(err, "could not get Kubernetes resource object")
	}

	acc := meta.NewAccessor()
	name, err := acc.Name(config)
	if err != nil {
		return errors.Wrap(err, "could not get name from a config object")
	}

	k.l.Debugf("Checking if config %s is in use", name)
	used, err := isInUse(ctx, name)
	if err != nil {
		return errors.Wrap(err, "could not check if config is in use")
	}
	if used {
		return errors.Wrapf(ErrConfigInUse, "config %s in use", name)
	}

	k.l.Debugf("Deleting config %s", name)

	if err := k.client.DeleteResource(ctx, config, &metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, "could not delete Kubernetes config object")
	}

	go func() {
		ctx := context.Background()
		secretName := cfg.SecretName()
		if secretName != "" {
			if err := k.DeleteSecret(ctx, secretName, k.namespace); err != nil {
				k.l.Error(errors.Wrapf(err, "could not delete secret %s for config %s", secretName, name))
			}
		}
	}()

	return nil
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
func (k *Kubernetes) updateConfigWithSecret(
	ctx context.Context, secretName string, obj runtime.Object, secretData map[string]string,
) error {
	oldSecret, err := k.GetSecret(ctx, secretName, k.namespace)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to read secret %s", secretName))
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: k.namespace,
		},
		StringData: secretData,
		Type:       corev1.SecretTypeOpaque,
	}
	_, err = k.UpdateSecret(ctx, secret)
	if err != nil {
		return err
	}

	if err := k.client.UpdateResource(ctx, obj, &metav1.UpdateOptions{}); err != nil {
		// rollback the changes
		_, err := k.UpdateSecret(ctx, oldSecret)
		if err != nil {
			k.l.Error(errors.Wrapf(err, "could not revert back secret %s", oldSecret.Name))
		}

		return errors.Wrap(err, "could not update config in Kubernetes")
	}

	return nil
}
