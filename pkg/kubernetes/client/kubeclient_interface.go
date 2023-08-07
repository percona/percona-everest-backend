// Code generated by ifacemaker; DO NOT EDIT.

package client

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// KubeClientConnector ...
type KubeClientConnector interface {
	// ClusterName returns the name of the k8s cluster.
	ClusterName() string
	// GetServerVersion returns server version.
	GetServerVersion() (*version.Info, error)
	// ListResources returns a list of k8s resources.
	ListResources(ctx context.Context, kind APIKind, into runtime.Object, opts *metav1.ListOptions) error
	// GetResource returns a resource by its name.
	GetResource(ctx context.Context, kind APIKind, name string, into runtime.Object, opts *metav1.GetOptions) error
	// CreateResource creates a k8s resource.
	CreateResource(ctx context.Context, kind APIKind, obj runtime.Object, into runtime.Object, opts *metav1.CreateOptions) error
	// UpdateResource updates a resource by its name.
	UpdateResource(ctx context.Context, kind APIKind, name string, obj runtime.Object, into runtime.Object, opts *metav1.UpdateOptions) error
	// DeleteResource deletes a resource by its name.
	DeleteResource(ctx context.Context, kind APIKind, name string, opts *metav1.DeleteOptions) error
	// GetSecret returns secret by name.
	GetSecret(ctx context.Context, name, namespace string) (*corev1.Secret, error)
}
