// Code generated by mockery v1.0.0. DO NOT EDIT.

package client

import (
	context "context"

	v1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	mock "github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	version "k8s.io/apimachinery/pkg/version"
)

// MockKubeClientConnector is an autogenerated mock type for the KubeClientConnector type
type MockKubeClientConnector struct {
	mock.Mock
}

// ApplyObject provides a mock function with given fields: obj
func (_m *MockKubeClientConnector) ApplyObject(obj runtime.Object) error {
	ret := _m.Called(obj)

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

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// CreateBackupStorage provides a mock function with given fields: ctx, storage
func (_m *MockKubeClientConnector) CreateBackupStorage(ctx context.Context, storage *v1alpha1.BackupStorage) error {
	ret := _m.Called(ctx, storage)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.BackupStorage) error); ok {
		r0 = rf(ctx, storage)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateMonitoringConfig provides a mock function with given fields: ctx, mc
func (_m *MockKubeClientConnector) CreateMonitoringConfig(ctx context.Context, mc *v1alpha1.MonitoringConfig) error {
	ret := _m.Called(ctx, mc)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.MonitoringConfig) error); ok {
		r0 = rf(ctx, mc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateResource provides a mock function with given fields: ctx, obj, opts
func (_m *MockKubeClientConnector) CreateResource(ctx context.Context, obj runtime.Object, opts *v1.CreateOptions) error {
	ret := _m.Called(ctx, obj, opts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, runtime.Object, *v1.CreateOptions) error); ok {
		r0 = rf(ctx, obj, opts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateSecret provides a mock function with given fields: ctx, secret
func (_m *MockKubeClientConnector) CreateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	ret := _m.Called(ctx, secret)

	var r0 *corev1.Secret
	if rf, ok := ret.Get(0).(func(context.Context, *corev1.Secret) *corev1.Secret); ok {
		r0 = rf(ctx, secret)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *corev1.Secret) error); ok {
		r1 = rf(ctx, secret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteBackupStorage provides a mock function with given fields: ctx, name, namespace
func (_m *MockKubeClientConnector) DeleteBackupStorage(ctx context.Context, name string, namespace string) error {
	ret := _m.Called(ctx, name, namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, name, namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteMonitoringConfig provides a mock function with given fields: ctx, name
func (_m *MockKubeClientConnector) DeleteMonitoringConfig(ctx context.Context, name string) error {
	ret := _m.Called(ctx, name)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteObject provides a mock function with given fields: obj
func (_m *MockKubeClientConnector) DeleteObject(obj runtime.Object) error {
	ret := _m.Called(obj)

	var r0 error
	if rf, ok := ret.Get(0).(func(runtime.Object) error); ok {
		r0 = rf(obj)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteSecret provides a mock function with given fields: ctx, name, namespace
func (_m *MockKubeClientConnector) DeleteSecret(ctx context.Context, name string, namespace string) error {
	ret := _m.Called(ctx, name, namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, name, namespace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetBackupStorage provides a mock function with given fields: ctx, name, namespace
func (_m *MockKubeClientConnector) GetBackupStorage(ctx context.Context, name string, namespace string) (*v1alpha1.BackupStorage, error) {
	ret := _m.Called(ctx, name, namespace)

	var r0 *v1alpha1.BackupStorage
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *v1alpha1.BackupStorage); ok {
		r0 = rf(ctx, name, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.BackupStorage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, name, namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDatabaseCluster provides a mock function with given fields: ctx, name
func (_m *MockKubeClientConnector) GetDatabaseCluster(ctx context.Context, name string) (*v1alpha1.DatabaseCluster, error) {
	ret := _m.Called(ctx, name)

	var r0 *v1alpha1.DatabaseCluster
	if rf, ok := ret.Get(0).(func(context.Context, string) *v1alpha1.DatabaseCluster); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.DatabaseCluster)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMonitoringConfig provides a mock function with given fields: ctx, name
func (_m *MockKubeClientConnector) GetMonitoringConfig(ctx context.Context, name string) (*v1alpha1.MonitoringConfig, error) {
	ret := _m.Called(ctx, name)

	var r0 *v1alpha1.MonitoringConfig
	if rf, ok := ret.Get(0).(func(context.Context, string) *v1alpha1.MonitoringConfig); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.MonitoringConfig)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNodes provides a mock function with given fields: ctx
func (_m *MockKubeClientConnector) GetNodes(ctx context.Context) (*corev1.NodeList, error) {
	ret := _m.Called(ctx)

	var r0 *corev1.NodeList
	if rf, ok := ret.Get(0).(func(context.Context) *corev1.NodeList); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.NodeList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPersistentVolumes provides a mock function with given fields: ctx
func (_m *MockKubeClientConnector) GetPersistentVolumes(ctx context.Context) (*corev1.PersistentVolumeList, error) {
	ret := _m.Called(ctx)

	var r0 *corev1.PersistentVolumeList
	if rf, ok := ret.Get(0).(func(context.Context) *corev1.PersistentVolumeList); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.PersistentVolumeList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPods provides a mock function with given fields: ctx, namespace, labelSelector
func (_m *MockKubeClientConnector) GetPods(ctx context.Context, namespace string, labelSelector *v1.LabelSelector) (*corev1.PodList, error) {
	ret := _m.Called(ctx, namespace, labelSelector)

	var r0 *corev1.PodList
	if rf, ok := ret.Get(0).(func(context.Context, string, *v1.LabelSelector) *corev1.PodList); ok {
		r0 = rf(ctx, namespace, labelSelector)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.PodList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, *v1.LabelSelector) error); ok {
		r1 = rf(ctx, namespace, labelSelector)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetResource provides a mock function with given fields: ctx, name, into, opts
func (_m *MockKubeClientConnector) GetResource(ctx context.Context, name string, into runtime.Object, opts *v1.GetOptions) error {
	ret := _m.Called(ctx, name, into, opts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, runtime.Object, *v1.GetOptions) error); ok {
		r0 = rf(ctx, name, into, opts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetSecret provides a mock function with given fields: ctx, name, namespace
func (_m *MockKubeClientConnector) GetSecret(ctx context.Context, name string, namespace string) (*corev1.Secret, error) {
	ret := _m.Called(ctx, name, namespace)

	var r0 *corev1.Secret
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *corev1.Secret); ok {
		r0 = rf(ctx, name, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, name, namespace)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetServerVersion provides a mock function with given fields:
func (_m *MockKubeClientConnector) GetServerVersion() (*version.Info, error) {
	ret := _m.Called()

	var r0 *version.Info
	if rf, ok := ret.Get(0).(func() *version.Info); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*version.Info)
		}
	}

	var r1 error
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

	var r0 *storagev1.StorageClassList
	if rf, ok := ret.Get(0).(func(context.Context) *storagev1.StorageClassList); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storagev1.StorageClassList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListDatabaseClusters provides a mock function with given fields: ctx
func (_m *MockKubeClientConnector) ListDatabaseClusters(ctx context.Context) (*v1alpha1.DatabaseClusterList, error) {
	ret := _m.Called(ctx)

	var r0 *v1alpha1.DatabaseClusterList
	if rf, ok := ret.Get(0).(func(context.Context) *v1alpha1.DatabaseClusterList); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.DatabaseClusterList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateBackupStorage provides a mock function with given fields: ctx, storage
func (_m *MockKubeClientConnector) UpdateBackupStorage(ctx context.Context, storage *v1alpha1.BackupStorage) error {
	ret := _m.Called(ctx, storage)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.BackupStorage) error); ok {
		r0 = rf(ctx, storage)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateSecret provides a mock function with given fields: ctx, secret
func (_m *MockKubeClientConnector) UpdateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	ret := _m.Called(ctx, secret)

	var r0 *corev1.Secret
	if rf, ok := ret.Get(0).(func(context.Context, *corev1.Secret) *corev1.Secret); ok {
		r0 = rf(ctx, secret)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *corev1.Secret) error); ok {
		r1 = rf(ctx, secret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
