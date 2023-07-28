// Copyright (C) 2017 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
