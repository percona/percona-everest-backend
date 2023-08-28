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

	backups, err := kubeClient.ListDatabaseClusterBackups(ctx)
	if err != nil {
		return false, errors.Wrap(err, "could not list database cluster backups in Kubernetes")
	}
	for _, b := range backups.Items {
		if b.Spec.BackupStorageName == name {
			return true, nil
		}
	}

	return false, nil
}

// BackupStorageNamesFromDBCluster returns a map of backup storage names used by a DB cluster.
func BackupStorageNamesFromDBCluster(db *everestv1alpha1.DatabaseCluster) map[string]struct{} {
	names := make(map[string]struct{})
	if db.Spec.DataSource != nil && db.Spec.DataSource.BackupStorageName != "" {
		names[db.Spec.DataSource.BackupStorageName] = struct{}{}
	}

	for _, schedule := range db.Spec.Backup.Schedules {
		if schedule.BackupStorageName != "" {
			names[schedule.BackupStorageName] = struct{}{}
		}
	}

	return names
}
