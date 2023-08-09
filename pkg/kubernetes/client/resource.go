// percona-everest-backend
// Copyright (C) 2023 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

func (c *Client) objectKind(obj runtime.Object) (schema.GroupVersionKind, error) {
	gvk := obj.GetObjectKind().GroupVersionKind()
	if gvk.Kind != "" {
		if !strings.HasSuffix(gvk.Kind, "List") {
			return gvk, nil
		}
		gvk.Kind = strings.TrimSuffix(gvk.Kind, "List")
		if scheme.Scheme.Recognizes(gvk) {
			return gvk, nil
		}
	}

	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		return schema.GroupVersionKind{}, errors.Wrap(err, "could not retrieve object kinds")
	}

	if len(gvks) != 1 {
		return schema.GroupVersionKind{}, errors.New("multiple group, version, kind options found. Specify kind explicitly")
	}

	return gvks[0], nil
}

// ListResources returns a list of k8s resources.
func (c *Client) ListResources(
	ctx context.Context,
	into runtime.Object, opts *metav1.ListOptions,
) error {
	gvk, err := c.objectKind(into)
	if err != nil {
		return err
	}

	err = c.everestClient.
		Get().
		Namespace(c.namespace).
		Resource(gvk.Kind).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(ctx).
		Into(into)

	return err
}

// GetResource returns a resource by its name.
func (c *Client) GetResource(
	ctx context.Context, name string,
	into runtime.Object, opts *metav1.GetOptions,
) error {
	gvk, err := c.objectKind(into)
	if err != nil {
		return err
	}

	err = c.everestClient.
		Get().
		Namespace(c.namespace).
		Resource(gvk.Kind).
		VersionedParams(opts, scheme.ParameterCodec).
		Name(name).
		Do(ctx).
		Into(into)

	return err
}

// CreateResource creates a k8s resource.
func (c *Client) CreateResource(
	ctx context.Context,
	obj runtime.Object, into runtime.Object, opts *metav1.CreateOptions,
) error {
	gvk, err := c.objectKind(obj)
	if err != nil {
		return err
	}

	err = c.everestClient.
		Post().
		Namespace(c.namespace).
		Resource(gvk.Kind).
		VersionedParams(opts, scheme.ParameterCodec).
		Body(obj).
		Do(ctx).
		Into(into)

	return err
}

// UpdateResource updates a resource by its name.
func (c *Client) UpdateResource(
	ctx context.Context, name string,
	obj runtime.Object, into runtime.Object, opts *metav1.UpdateOptions,
) error {
	gvk, err := c.objectKind(obj)
	if err != nil {
		return err
	}

	err = c.everestClient.
		Put().
		Namespace(c.namespace).
		Resource(gvk.Kind).
		VersionedParams(opts, scheme.ParameterCodec).
		Name(name).
		Body(obj).
		Do(ctx).
		Into(into)

	return err
}

// DeleteResource deletes a resource by its name.
func (c *Client) DeleteResource(
	ctx context.Context, name string,
	obj runtime.Object, opts *metav1.DeleteOptions,
) error {
	gvk, err := c.objectKind(obj)
	if err != nil {
		return err
	}

	r := c.everestClient.
		Delete().
		Namespace(c.namespace).
		Resource(gvk.Kind).
		VersionedParams(opts, scheme.ParameterCodec).
		Name(name).
		Do(ctx)

	return r.Error()
}
