// Code generated by mockery v1.0.0. DO NOT EDIT.

package client

import (
	context "context"

	v1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	mock "github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	version "k8s.io/apimachinery/pkg/version"
)

// MockKubeClientConnector is an autogenerated mock type for the KubeClientConnector type
type MockKubeClientConnector struct {
	mock.Mock
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

// CreateObjectStorage provides a mock function with given fields: ctx, storage
func (_m *MockKubeClientConnector) CreateObjectStorage(ctx context.Context, storage *v1alpha1.ObjectStorage) error {
	ret := _m.Called(ctx, storage)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.ObjectStorage) error); ok {
		r0 = rf(ctx, storage)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateSecret provides a mock function with given fields: ctx, secret
func (_m *MockKubeClientConnector) CreateSecret(ctx context.Context, secret *v1.Secret) (*v1.Secret, error) {
	ret := _m.Called(ctx, secret)

	var r0 *v1.Secret
	if rf, ok := ret.Get(0).(func(context.Context, *v1.Secret) *v1.Secret); ok {
		r0 = rf(ctx, secret)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *v1.Secret) error); ok {
		r1 = rf(ctx, secret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteObjectStorage provides a mock function with given fields: ctx, name, namespace
func (_m *MockKubeClientConnector) DeleteObjectStorage(ctx context.Context, name string, namespace string) error {
	ret := _m.Called(ctx, name, namespace)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, name, namespace)
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

// GetSecret provides a mock function with given fields: ctx, name, namespace
func (_m *MockKubeClientConnector) GetSecret(ctx context.Context, name string, namespace string) (*v1.Secret, error) {
	ret := _m.Called(ctx, name, namespace)

	var r0 *v1.Secret
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *v1.Secret); ok {
		r0 = rf(ctx, name, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
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
