// Code generated by mockery v2.40.1. DO NOT EDIT.

package client

import (
	context "context"

	v1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	mock "github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	version "k8s.io/apimachinery/pkg/version"
	rest "k8s.io/client-go/rest"
)

// MockKubeClientConnector is an autogenerated mock type for the KubeClientConnector type
type MockKubeClientConnector struct {
	mock.Mock
}

// ApplyObject provides a mock function with given fields: obj
func (_m *MockKubeClientConnector) ApplyObject(obj runtime.Object) error {
	ret := _m.Called(obj)

	if len(ret) == 0 {
		panic("no return value specified for ApplyObject")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(runtime.Object) error); ok {
		r0 = rf(obj)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ClusterName provides a mock function with given fields:
func (_m *MockKubeClientConnector) ClusterName() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ClusterName")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Config provides a mock function with given fields:
func (_m *MockKubeClientConnector) Config() *rest.Config {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Config")
	}

	var r0 *rest.Config
	if rf, ok := ret.Get(0).(func() *rest.Config); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*rest.Config)
		}
	}

	return r0
}

// CreateBackupStorage provides a mock function with given fields: ctx, storage
func (_m *MockKubeClientConnector) CreateBackupStorage(ctx context.Context, storage *v1alpha1.BackupStorage) error {
	ret := _m.Called(ctx, storage)

	if len(ret) == 0 {
		panic("no return value specified for CreateBackupStorage")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.BackupStorage) error); ok {
		r0 = rf(ctx, storage)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateMonitoringConfig provides a mock function with given fields: ctx, config
func (_m *MockKubeClientConnector) CreateMonitoringConfig(ctx context.Context, config *v1alpha1.MonitoringConfig) error {
	ret := _m.Called(ctx, config)

	if len(ret) == 0 {
		panic("no return value specified for CreateMonitoringConfig")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.MonitoringConfig) error); ok {
		r0 = rf(ctx, config)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateSecret provides a mock function with given fields: ctx, secret
func (_m *MockKubeClientConnector) CreateSecret(ctx context.Context, secret *v1.Secret) (*v1.Secret, error) {
	ret := _m.Called(ctx, secret)

	if len(ret) == 0 {
		panic("no return value specified for CreateSecret")
	}

	var r0 *v1.Secret
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.Secret) (*v1.Secret, error)); ok {
		return rf(ctx, secret)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.Secret) *v1.Secret); ok {
		r0 = rf(ctx, secret)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.Secret) error); ok {
		r1 = rf(ctx, secret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteBackupStorage provides a mock function with given fields: ctx, name
func (_m *MockKubeClientConnector) DeleteBackupStorage(ctx context.Context, name string) error {
	ret := _m.Called(ctx, name)

	if len(ret) == 0 {
		panic("no return value specified for DeleteBackupStorage")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteMonitoringConfig provides a mock function with given fields: ctx, namespace, name
func (_m *MockKubeClientConnector) DeleteMonitoringConfig(ctx context.Context, namespace string, name string) error {
	ret := _m.Called(ctx, namespace, name)

	if len(ret) == 0 {
		panic("no return value specified for DeleteMonitoringConfig")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, namespace, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteObject provides a mock function with given fields: obj
func (_m *MockKubeClientConnector) DeleteObject(obj runtime.Object) error {
	ret := _m.Called(obj)

	if len(ret) == 0 {
		panic("no return value specified for DeleteObject")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(runtime.Object) error); ok {
		r0 = rf(obj)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteSecret provides a mock function with given fields: ctx, namespace, name
func (_m *MockKubeClientConnector) DeleteSecret(ctx context.Context, namespace string, name string) error {
	ret := _m.Called(ctx, namespace, name)

	if len(ret) == 0 {
		panic("no return value specified for DeleteSecret")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, namespace, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetBackupStorage provides a mock function with given fields: ctx, name
func (_m *MockKubeClientConnector) GetBackupStorage(ctx context.Context, name string) (*v1alpha1.BackupStorage, error) {
	ret := _m.Called(ctx, name)

	if len(ret) == 0 {
		panic("no return value specified for GetBackupStorage")
	}

	var r0 *v1alpha1.BackupStorage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*v1alpha1.BackupStorage, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *v1alpha1.BackupStorage); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.BackupStorage)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetConfigMap provides a mock function with given fields: ctx, namespace, name
func (_m *MockKubeClientConnector) GetConfigMap(ctx context.Context, namespace string, name string) (*v1.ConfigMap, error) {
	ret := _m.Called(ctx, namespace, name)

	if len(ret) == 0 {
		panic("no return value specified for GetConfigMap")
	}

	var r0 *v1.ConfigMap
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*v1.ConfigMap, error)); ok {
		return rf(ctx, namespace, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *v1.ConfigMap); ok {
		r0 = rf(ctx, namespace, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ConfigMap)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, namespace, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDatabaseCluster provides a mock function with given fields: ctx, namespace, name
func (_m *MockKubeClientConnector) GetDatabaseCluster(ctx context.Context, namespace string, name string) (*v1alpha1.DatabaseCluster, error) {
	ret := _m.Called(ctx, namespace, name)

	if len(ret) == 0 {
		panic("no return value specified for GetDatabaseCluster")
	}

	var r0 *v1alpha1.DatabaseCluster
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*v1alpha1.DatabaseCluster, error)); ok {
		return rf(ctx, namespace, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *v1alpha1.DatabaseCluster); ok {
		r0 = rf(ctx, namespace, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.DatabaseCluster)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, namespace, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDatabaseClusterBackup provides a mock function with given fields: ctx, name
func (_m *MockKubeClientConnector) GetDatabaseClusterBackup(ctx context.Context, name string) (*v1alpha1.DatabaseClusterBackup, error) {
	ret := _m.Called(ctx, name)

	if len(ret) == 0 {
		panic("no return value specified for GetDatabaseClusterBackup")
	}

	var r0 *v1alpha1.DatabaseClusterBackup
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*v1alpha1.DatabaseClusterBackup, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *v1alpha1.DatabaseClusterBackup); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.DatabaseClusterBackup)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDatabaseClusterRestore provides a mock function with given fields: ctx, name
func (_m *MockKubeClientConnector) GetDatabaseClusterRestore(ctx context.Context, name string) (*v1alpha1.DatabaseClusterRestore, error) {
	ret := _m.Called(ctx, name)

	if len(ret) == 0 {
		panic("no return value specified for GetDatabaseClusterRestore")
	}

	var r0 *v1alpha1.DatabaseClusterRestore
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*v1alpha1.DatabaseClusterRestore, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *v1alpha1.DatabaseClusterRestore); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.DatabaseClusterRestore)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDatabaseEngine provides a mock function with given fields: ctx, name
func (_m *MockKubeClientConnector) GetDatabaseEngine(ctx context.Context, name string) (*v1alpha1.DatabaseEngine, error) {
	ret := _m.Called(ctx, name)

	if len(ret) == 0 {
		panic("no return value specified for GetDatabaseEngine")
	}

	var r0 *v1alpha1.DatabaseEngine
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*v1alpha1.DatabaseEngine, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *v1alpha1.DatabaseEngine); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.DatabaseEngine)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDeployment provides a mock function with given fields: ctx, name, namespace
func (_m *MockKubeClientConnector) GetDeployment(ctx context.Context, name string, namespace string) (*appsv1.Deployment, error) {
	ret := _m.Called(ctx, name, namespace)

	if len(ret) == 0 {
		panic("no return value specified for GetDeployment")
	}

	var r0 *appsv1.Deployment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*appsv1.Deployment, error)); ok {
		return rf(ctx, name, namespace)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *appsv1.Deployment); ok {
		r0 = rf(ctx, name, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*appsv1.Deployment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, name, namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMonitoringConfig provides a mock function with given fields: ctx, namespace, name
func (_m *MockKubeClientConnector) GetMonitoringConfig(ctx context.Context, namespace string, name string) (*v1alpha1.MonitoringConfig, error) {
	ret := _m.Called(ctx, namespace, name)

	if len(ret) == 0 {
		panic("no return value specified for GetMonitoringConfig")
	}

	var r0 *v1alpha1.MonitoringConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*v1alpha1.MonitoringConfig, error)); ok {
		return rf(ctx, namespace, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *v1alpha1.MonitoringConfig); ok {
		r0 = rf(ctx, namespace, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.MonitoringConfig)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, namespace, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNamespace provides a mock function with given fields: ctx, name
func (_m *MockKubeClientConnector) GetNamespace(ctx context.Context, name string) (*v1.Namespace, error) {
	ret := _m.Called(ctx, name)

	if len(ret) == 0 {
		panic("no return value specified for GetNamespace")
	}

	var r0 *v1.Namespace
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*v1.Namespace, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *v1.Namespace); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Namespace)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNodes provides a mock function with given fields: ctx
func (_m *MockKubeClientConnector) GetNodes(ctx context.Context) (*v1.NodeList, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetNodes")
	}

	var r0 *v1.NodeList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*v1.NodeList, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *v1.NodeList); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.NodeList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetObject provides a mock function with given fields: gvk, name, into
func (_m *MockKubeClientConnector) GetObject(gvk schema.GroupVersionKind, name string, into runtime.Object) error {
	ret := _m.Called(gvk, name, into)

	if len(ret) == 0 {
		panic("no return value specified for GetObject")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(schema.GroupVersionKind, string, runtime.Object) error); ok {
		r0 = rf(gvk, name, into)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetPersistentVolumes provides a mock function with given fields: ctx
func (_m *MockKubeClientConnector) GetPersistentVolumes(ctx context.Context) (*v1.PersistentVolumeList, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetPersistentVolumes")
	}

	var r0 *v1.PersistentVolumeList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*v1.PersistentVolumeList, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *v1.PersistentVolumeList); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.PersistentVolumeList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPods provides a mock function with given fields: ctx, namespace, labelSelector
func (_m *MockKubeClientConnector) GetPods(ctx context.Context, namespace string, labelSelector *metav1.LabelSelector) (*v1.PodList, error) {
	ret := _m.Called(ctx, namespace, labelSelector)

	if len(ret) == 0 {
		panic("no return value specified for GetPods")
	}

	var r0 *v1.PodList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *metav1.LabelSelector) (*v1.PodList, error)); ok {
		return rf(ctx, namespace, labelSelector)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *metav1.LabelSelector) *v1.PodList); ok {
		r0 = rf(ctx, namespace, labelSelector)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.PodList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *metav1.LabelSelector) error); ok {
		r1 = rf(ctx, namespace, labelSelector)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSecret provides a mock function with given fields: ctx, namespace, name
func (_m *MockKubeClientConnector) GetSecret(ctx context.Context, namespace string, name string) (*v1.Secret, error) {
	ret := _m.Called(ctx, namespace, name)

	if len(ret) == 0 {
		panic("no return value specified for GetSecret")
	}

	var r0 *v1.Secret
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*v1.Secret, error)); ok {
		return rf(ctx, namespace, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *v1.Secret); ok {
		r0 = rf(ctx, namespace, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, namespace, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetServerVersion provides a mock function with given fields:
func (_m *MockKubeClientConnector) GetServerVersion() (*version.Info, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetServerVersion")
	}

	var r0 *version.Info
	var r1 error
	if rf, ok := ret.Get(0).(func() (*version.Info, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *version.Info); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*version.Info)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetStorageClasses provides a mock function with given fields: ctx
func (_m *MockKubeClientConnector) GetStorageClasses(ctx context.Context) (*storagev1.StorageClassList, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetStorageClasses")
	}

	var r0 *storagev1.StorageClassList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*storagev1.StorageClassList, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *storagev1.StorageClassList); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storagev1.StorageClassList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListBackupStorages provides a mock function with given fields: ctx, options
func (_m *MockKubeClientConnector) ListBackupStorages(ctx context.Context, options metav1.ListOptions) (*v1alpha1.BackupStorageList, error) {
	ret := _m.Called(ctx, options)

	if len(ret) == 0 {
		panic("no return value specified for ListBackupStorages")
	}

	var r0 *v1alpha1.BackupStorageList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) (*v1alpha1.BackupStorageList, error)); ok {
		return rf(ctx, options)
	}
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) *v1alpha1.BackupStorageList); ok {
		r0 = rf(ctx, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.BackupStorageList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, metav1.ListOptions) error); ok {
		r1 = rf(ctx, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListDatabaseClusterBackups provides a mock function with given fields: ctx, options
func (_m *MockKubeClientConnector) ListDatabaseClusterBackups(ctx context.Context, options metav1.ListOptions) (*v1alpha1.DatabaseClusterBackupList, error) {
	ret := _m.Called(ctx, options)

	if len(ret) == 0 {
		panic("no return value specified for ListDatabaseClusterBackups")
	}

	var r0 *v1alpha1.DatabaseClusterBackupList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) (*v1alpha1.DatabaseClusterBackupList, error)); ok {
		return rf(ctx, options)
	}
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) *v1alpha1.DatabaseClusterBackupList); ok {
		r0 = rf(ctx, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.DatabaseClusterBackupList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, metav1.ListOptions) error); ok {
		r1 = rf(ctx, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListDatabaseClusterRestores provides a mock function with given fields: ctx, options
func (_m *MockKubeClientConnector) ListDatabaseClusterRestores(ctx context.Context, options metav1.ListOptions) (*v1alpha1.DatabaseClusterRestoreList, error) {
	ret := _m.Called(ctx, options)

	if len(ret) == 0 {
		panic("no return value specified for ListDatabaseClusterRestores")
	}

	var r0 *v1alpha1.DatabaseClusterRestoreList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) (*v1alpha1.DatabaseClusterRestoreList, error)); ok {
		return rf(ctx, options)
	}
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) *v1alpha1.DatabaseClusterRestoreList); ok {
		r0 = rf(ctx, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.DatabaseClusterRestoreList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, metav1.ListOptions) error); ok {
		r1 = rf(ctx, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListDatabaseClusters provides a mock function with given fields: ctx, namespace, options
func (_m *MockKubeClientConnector) ListDatabaseClusters(ctx context.Context, namespace string, options metav1.ListOptions) (*v1alpha1.DatabaseClusterList, error) {
	ret := _m.Called(ctx, namespace, options)

	if len(ret) == 0 {
		panic("no return value specified for ListDatabaseClusters")
	}

	var r0 *v1alpha1.DatabaseClusterList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, metav1.ListOptions) (*v1alpha1.DatabaseClusterList, error)); ok {
		return rf(ctx, namespace, options)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, metav1.ListOptions) *v1alpha1.DatabaseClusterList); ok {
		r0 = rf(ctx, namespace, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.DatabaseClusterList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, metav1.ListOptions) error); ok {
		r1 = rf(ctx, namespace, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListDatabaseEngines provides a mock function with given fields: ctx
func (_m *MockKubeClientConnector) ListDatabaseEngines(ctx context.Context) (*v1alpha1.DatabaseEngineList, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for ListDatabaseEngines")
	}

	var r0 *v1alpha1.DatabaseEngineList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*v1alpha1.DatabaseEngineList, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *v1alpha1.DatabaseEngineList); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.DatabaseEngineList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListMonitoringConfigs provides a mock function with given fields: ctx, namespace
func (_m *MockKubeClientConnector) ListMonitoringConfigs(ctx context.Context, namespace string) (*v1alpha1.MonitoringConfigList, error) {
	ret := _m.Called(ctx, namespace)

	if len(ret) == 0 {
		panic("no return value specified for ListMonitoringConfigs")
	}

	var r0 *v1alpha1.MonitoringConfigList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*v1alpha1.MonitoringConfigList, error)); ok {
		return rf(ctx, namespace)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *v1alpha1.MonitoringConfigList); ok {
		r0 = rf(ctx, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.MonitoringConfigList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListObjects provides a mock function with given fields: gvk, into
func (_m *MockKubeClientConnector) ListObjects(gvk schema.GroupVersionKind, into runtime.Object) error {
	ret := _m.Called(gvk, into)

	if len(ret) == 0 {
		panic("no return value specified for ListObjects")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(schema.GroupVersionKind, runtime.Object) error); ok {
		r0 = rf(gvk, into)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Namespace provides a mock function with given fields:
func (_m *MockKubeClientConnector) Namespace() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Namespace")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// UpdateBackupStorage provides a mock function with given fields: ctx, storage
func (_m *MockKubeClientConnector) UpdateBackupStorage(ctx context.Context, storage *v1alpha1.BackupStorage) error {
	ret := _m.Called(ctx, storage)

	if len(ret) == 0 {
		panic("no return value specified for UpdateBackupStorage")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.BackupStorage) error); ok {
		r0 = rf(ctx, storage)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateMonitoringConfig provides a mock function with given fields: ctx, config
func (_m *MockKubeClientConnector) UpdateMonitoringConfig(ctx context.Context, config *v1alpha1.MonitoringConfig) error {
	ret := _m.Called(ctx, config)

	if len(ret) == 0 {
		panic("no return value specified for UpdateMonitoringConfig")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.MonitoringConfig) error); ok {
		r0 = rf(ctx, config)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateSecret provides a mock function with given fields: ctx, secret
func (_m *MockKubeClientConnector) UpdateSecret(ctx context.Context, secret *v1.Secret) (*v1.Secret, error) {
	ret := _m.Called(ctx, secret)

	if len(ret) == 0 {
		panic("no return value specified for UpdateSecret")
	}

	var r0 *v1.Secret
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.Secret) (*v1.Secret, error)); ok {
		return rf(ctx, secret)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.Secret) *v1.Secret); ok {
		r0 = rf(ctx, secret)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.Secret) error); ok {
		r1 = rf(ctx, secret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMockKubeClientConnector creates a new instance of MockKubeClientConnector. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockKubeClientConnector(t interface {
	mock.TestingT
	Cleanup(func())
},
) *MockKubeClientConnector {
	mock := &MockKubeClientConnector{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
