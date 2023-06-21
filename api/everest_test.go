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
			url:          "/v1/kubernetes/123/database-clusters",
			kubernetesID: "123",
			resourceName: "",
			expected:     "/apis/dbaas.percona.com/v1/namespaces/percona-everest/databaseclusters",
		},
		{
			url:          "/v1/kubernetes/123/database-clusters/snake_case_name",
			kubernetesID: "123",
			resourceName: "snake_case_name",
			expected:     "/apis/dbaas.percona.com/v1/namespaces/percona-everest/databaseclusters/snake_case_name",
		},
		{
			url:          "/v1/kubernetes/123/database-clusters/kebab-case-name",
			kubernetesID: "123",
			resourceName: "kebab-case-name",
			expected:     "/apis/dbaas.percona.com/v1/namespaces/percona-everest/databaseclusters/kebab-case-name",
		},
		{
			url:          "/v1/kubernetes/123/database-cluster-restores/kebab-case-name",
			kubernetesID: "123",
			resourceName: "kebab-case-name",
			expected:     "/apis/dbaas.percona.com/v1/namespaces/percona-everest/databaseclusterrestores/kebab-case-name",
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
