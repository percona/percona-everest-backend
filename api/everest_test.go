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
package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildProxiedUrl(t *testing.T) {
	t.Parallel()
	type tCase struct {
		url          string
		resourceName string
		expected     string
	}

	cases := []tCase{
		{
			url:          "/v1/database-clusters",
			resourceName: "",
			expected:     "/apis/everest.percona.com/v1alpha1/namespaces/percona-everest/databaseclusters",
		},
		{
			url:          "/v1/database-clusters/snake_case_name",
			resourceName: "snake_case_name",
			expected:     "/apis/everest.percona.com/v1alpha1/namespaces/percona-everest/databaseclusters/snake_case_name",
		},
		{
			url:          "/v1/database-clusters/kebab-case-name",
			resourceName: "kebab-case-name",
			expected:     "/apis/everest.percona.com/v1alpha1/namespaces/percona-everest/databaseclusters/kebab-case-name",
		},
		{
			url:          "/v1/database-cluster-restores/kebab-case-name",
			resourceName: "kebab-case-name",
			expected:     "/apis/everest.percona.com/v1alpha1/namespaces/percona-everest/databaseclusterrestores/kebab-case-name",
		},
	}

	for _, testCase := range cases {
		tc := testCase
		t.Run(tc.url, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, buildProxiedURL(tc.url, tc.resourceName, "percona-everest"))
		})
	}
}
