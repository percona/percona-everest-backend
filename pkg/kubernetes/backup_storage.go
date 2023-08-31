package kubernetes

import (
	"context"

	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/pkg/errors"
)

// IsBackupStorageConfigInUse returns true if the backup storage is in use
// by the provided Kubernetes cluster.
func IsBackupStorageConfigInUse(ctx context.Context, name string, kubeClient *Kubernetes) (bool, error) {
	dbs, err := kubeClient.ListDatabaseClusters(ctx)
	if err != nil {
		return false, errors.Wrap(err, "could not list database clusters in Kubernetes")
	}

	for _, db := range dbs.Items {
		db := db
		names := BackupStorageNamesFromDBCluster(&db)
		if _, ok := names[name]; ok {
			return true, nil
		}
	}

	restores, err := kubeClient.ListDatabaseClusterRestores(ctx)
	if err != nil {
		return false, errors.Wrap(err, "could not list database cluster restores in Kubernetes")
	}

	for _, restore := range restores.Items {
		if restore.Spec.DataSource.BackupSource != nil && restore.Spec.DataSource.BackupSource.BackupStorageName == name {
			return true, nil
		}
	}

	return false, nil
}

// BackupStorageNamesFromDBCluster returns a map of backup storage names used by a DB cluster.
func BackupStorageNamesFromDBCluster(db *everestv1alpha1.DatabaseCluster) map[string]struct{} {
	names := make(map[string]struct{})
	if db.Spec.DataSource != nil && db.Spec.DataSource.BackupSource != nil && db.Spec.DataSource.BackupSource.BackupStorageName != "" {
		names[db.Spec.DataSource.BackupSource.BackupStorageName] = struct{}{}
	}

	for _, schedule := range db.Spec.Backup.Schedules {
		if schedule.BackupStorageName != "" {
			names[schedule.BackupStorageName] = struct{}{}
		}
	}

	return names
}
