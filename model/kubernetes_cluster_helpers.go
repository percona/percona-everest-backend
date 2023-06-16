package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const defaultK8sNamespace = "percona-everest"

// CreateKubernetesCluster creates a KubernetesCluster record.
func (db *Database) CreateKubernetesCluster(_ echo.Context, params CreateKubernetesClusterParams) (*KubernetesCluster, error) {
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
func (db *Database) ListKubernetesClusters(_ echo.Context) ([]KubernetesCluster, error) {
	var clusters []KubernetesCluster
	err := db.gormDB.Find(&clusters).Error
	if err != nil {
		return nil, err
	}
	return clusters, nil
}

// GetKubernetesCluster returns KubernetesCluster record by its ID.
func (db *Database) GetKubernetesCluster(_ echo.Context, id string) (*KubernetesCluster, error) {
	var cluster KubernetesCluster
	err := db.gormDB.First(&cluster, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}
