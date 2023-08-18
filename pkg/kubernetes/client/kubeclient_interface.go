// Code generated by ifacemaker; DO NOT EDIT.

package client

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// KubeClientConnector ...
type KubeClientConnector interface {
	// ClusterName returns the name of the k8s cluster.
	ClusterName() string
	// GetServerVersion returns server version.
	GetServerVersion() (*version.Info, error)
	// ApplyObject applies object.
	ApplyObject(obj runtime.Object) error
	// DeleteObject deletes object from the k8s cluster.
	DeleteObject(obj runtime.Object) error
	// ListObjects lists objects by provided group, version, kind.
	ListObjects(gvk schema.GroupVersionKind, into runtime.Object) error
	// GetObject retrieves an object by provided group, version, kind and name.
	GetObject(gvk schema.GroupVersionKind, name string, into runtime.Object) error
	// ListDatabaseClusters returns list of managed database clusters.
	ListDatabaseClusters(ctx context.Context) (*everestv1alpha1.DatabaseClusterList, error)
	// GetDatabaseCluster returns database clusters by provided name.
	GetDatabaseCluster(ctx context.Context, name string) (*everestv1alpha1.DatabaseCluster, error)
	// CreateMonitoringConfig creates an MonitoringConfig.
	CreateMonitoringConfig(ctx context.Context, mc *everestv1alpha1.MonitoringConfig) error
	// GetMonitoringConfig returns the MonitoringConfig.
	GetMonitoringConfig(ctx context.Context, name string) (*everestv1alpha1.MonitoringConfig, error)
	// DeleteMonitoringConfig deletes the MonitoringConfig.
	DeleteMonitoringConfig(ctx context.Context, name string) error
	// ListMonitoringConfigs returns list of MonitoringConfig.
	ListMonitoringConfigs(ctx context.Context) (*everestv1alpha1.MonitoringConfigList, error)
	// GetNodes returns list of nodes.
	GetNodes(ctx context.Context) (*corev1.NodeList, error)
	// GetPods returns list of pods.
	GetPods(ctx context.Context, namespace string, labelSelector *metav1.LabelSelector) (*corev1.PodList, error)
	// GetResource returns a resource by its name.
	GetResource(ctx context.Context, name string, into runtime.Object, opts *metav1.GetOptions) error
	// CreateResource creates a k8s resource.
	CreateResource(ctx context.Context, obj runtime.Object, opts *metav1.CreateOptions) error
	// GetSecret returns secret by name.
	GetSecret(ctx context.Context, name, namespace string) (*corev1.Secret, error)
	// UpdateSecret updates k8s Secret.
	UpdateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error)
	// CreateSecret creates k8s Secret.
	CreateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error)
	// DeleteSecret deletes the k8s Secret.
	DeleteSecret(ctx context.Context, name, namespace string) error
	// GetStorageClasses returns all storage classes available in the cluster.
	GetStorageClasses(ctx context.Context) (*storagev1.StorageClassList, error)
	// GetPersistentVolumes returns Persistent Volumes available in the cluster.
	GetPersistentVolumes(ctx context.Context) (*corev1.PersistentVolumeList, error)
	// CreateBackupStorage creates an backupStorage.
	CreateBackupStorage(ctx context.Context, storage *everestv1alpha1.BackupStorage) error
	// UpdateBackupStorage updates an backupStorage.
	UpdateBackupStorage(ctx context.Context, storage *everestv1alpha1.BackupStorage) error
	// GetBackupStorage returns the backupStorage.
	GetBackupStorage(ctx context.Context, name, namespace string) (*everestv1alpha1.BackupStorage, error)
	// DeleteBackupStorage deletes the backupStorage.
	DeleteBackupStorage(ctx context.Context, name, namespace string) error
}
