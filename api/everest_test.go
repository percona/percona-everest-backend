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
		kubernetesID string
		expected     string
	}

	cases := []tCase{
		{
			url:          "/kubernetes/123/database-clusters",
			kubernetesID: "123",
			resourceName: "",
			expected:     "/apis/dbaas.percona.com/v1/namespaces/default/databaseclusters",
		},
		{
			url:          "/kubernetes/123/database-clusters/snake_case_name",
			kubernetesID: "123",
			resourceName: "snake_case_name",
			expected:     "/apis/dbaas.percona.com/v1/namespaces/default/databaseclusters/snake_case_name",
		},
		{
			url:          "/kubernetes/123/database-clusters/kebab-case-name",
			kubernetesID: "123",
			resourceName: "kebab-case-name",
			expected:     "/apis/dbaas.percona.com/v1/namespaces/default/databaseclusters/kebab-case-name",
		},
		{
			url:          "/kubernetes/123/database-cluster-restores/kebab-case-name",
			kubernetesID: "123",
			resourceName: "kebab-case-name",
			expected:     "/apis/dbaas.percona.com/v1/namespaces/default/databaseclusterrestores/kebab-case-name",
		},
	}

	for _, testCase := range cases {
		tc := testCase
		t.Run(tc.url, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, buildProxiedURL(tc.url, tc.kubernetesID, tc.resourceName))
		})
	}
}
