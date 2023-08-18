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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// GetResource returns a resource by its name.
func (c *Client) GetResource(
	ctx context.Context, name string,
	into runtime.Object, opts *metav1.GetOptions,
) error {
	return c.customClientSet.GetResource(ctx, c.namespace, name, into, opts)
}

// CreateResource creates a k8s resource.
func (c *Client) CreateResource(
	ctx context.Context,
	obj runtime.Object, opts *metav1.CreateOptions,
) error {
	return c.customClientSet.CreateResource(ctx, c.namespace, obj, opts)
}
