package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStorageClasses(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		storagesList *storagev1.StorageClassList
		result       []string
	}{
		{
			name: "no-default",
			storagesList: &storagev1.StorageClassList{
				Items: []storagev1.StorageClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "local-storage",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "cool-storage",
						},
					},
				},
			},
			result: []string{"local-storage", "cool-storage"},
		},
		{
			name: "default is the first item",
			storagesList: &storagev1.StorageClassList{
				Items: []storagev1.StorageClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "local-storage",
							Annotations: map[string]string{
								annotationStorageClassDefault: "true",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "cool-storage",
						},
					},
				},
			},
			result: []string{"local-storage", "cool-storage"},
		},
		{
			name: "default is the middle item",
			storagesList: &storagev1.StorageClassList{
				Items: []storagev1.StorageClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "cool-storage",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "local-storage",
							Annotations: map[string]string{
								annotationStorageClassDefault: "true",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "another-storage",
						},
					},
				},
			},
			result: []string{"local-storage", "cool-storage", "another-storage"},
		},
		{
			name: "default is the last item",
			storagesList: &storagev1.StorageClassList{
				Items: []storagev1.StorageClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "cool-storage",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "another-storage",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "local-storage",
							Annotations: map[string]string{
								annotationStorageClassDefault: "true",
							},
						},
					},
				},
			},
			result: []string{"local-storage", "another-storage", "cool-storage"},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, storageClasses(tc.storagesList), tc.result)
		})
	}
}
