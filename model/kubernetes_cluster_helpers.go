package model

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CreateKubernetesClusterParams parameters for KubernetesCluster record creation.
type CreateKubernetesClusterParams struct {
	Name string
}

// KubernetesCluster represents db model for KubernetesCluster.
type KubernetesCluster struct {
	ID   string
	Name string

	CreatedAt time.Time
	UpdatedAt time.Time
}

const defaultK8sNamespace = "percona-everest"

// CreateKubernetesCluster creates a KubernetesCluster record.
func (db *Database) CreateKubernetesCluster(_ context.Context, params CreateKubernetesClusterParams) (*KubernetesCluster, error) {
	namespace := defaultK8sNamespace
	if params.Namespace != nil {
		namespace = *params.Namespace
	}

	k := &KubernetesCluster{
		ID:        uuid.NewString(),
		Name:      params.Name,
		Namespace: namespace,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	err := db.gormDB.Create(k).Error
	if err != nil {
		return nil, err
	}

	return k, nil
}

// ListKubernetesClusters returns all available KubernetesCluster records.
func (db *Database) ListKubernetesClusters(_ context.Context) ([]KubernetesCluster, error) {
	var clusters []KubernetesCluster
	err := db.gormDB.Find(&clusters).Error
	if err != nil {
		return nil, err
	}
	return clusters, nil
}

// GetKubernetesCluster returns KubernetesCluster record by its ID.
func (db *Database) GetKubernetesCluster(_ context.Context, id string) (*KubernetesCluster, error) {
	var cluster KubernetesCluster
	err := db.gormDB.First(&cluster, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}
