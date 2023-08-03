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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/version"
	fake "k8s.io/client-go/kubernetes/fake"
)

func TestGetServerVersion(t *testing.T) {
	t.Parallel()
	clientset := fake.NewSimpleClientset()
	client := &Client{clientset: clientset, namespace: "default"}
	ver, err := client.GetServerVersion()
	expectedVersion := &version.Info{}
	require.NoError(t, err)
	assert.Equal(t, expectedVersion.Minor, ver.Minor)
}
