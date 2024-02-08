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
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/AlekSi/pointer"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateRFC1035(t *testing.T) {
	t.Parallel()
	type testCase struct {
		value string
		valid bool
	}

	cases := []testCase{
		{
			value: "abc-sdf12",
			valid: true,
		},
		{
			value: "-abc-sdf12",
			valid: false,
		},
		{
			value: "abc-sdf12-",
			valid: false,
		},
		{
			value: "abc-sAAf12",
			valid: false,
		},
		{
			value: "abc-sAAf12",
			valid: false,
		},
		{
			value: "1abc-sf12",
			valid: false,
		},
		{
			value: "aaa123",
			valid: true,
		},
		{
			value: "asldkafaslkdjfalskdfjaslkdjflsakfjdalskfdjaslkfdjaslkfdjsaklfdassksjdfhskdjfskjdfsdfsdflasdkfasdfk",
			valid: false,
		},
		{
			value: "$%",
			valid: false,
		},
		{
			value: "asdf32$%",
			valid: false,
		},
		{
			value: "",
			valid: false,
		},
	}

	for _, tc := range cases {
		c := tc
		t.Run(c.value, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, c.valid, validateRFC1035(c.value, "") == nil)
		})
	}
}

func TestValidateCreateDatabaseClusterRequest(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name  string
		value DatabaseCluster
		err   error
	}

	cases := []testCase{
		{
			name:  "empty metadata",
			value: DatabaseCluster{},
			err:   errDBCEmptyMetadata,
		},
		{
			name:  "no dbCluster name",
			value: DatabaseCluster{Metadata: &map[string]interface{}{}},
			err:   errDBCNameEmpty,
		},
		{
			name: "empty dbCluster name",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": "",
			}},
			err: ErrNameNotRFC1035Compatible("metadata.name"),
		},
		{
			name: "starts with -",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": "-sdfasa",
			}},
			err: ErrNameNotRFC1035Compatible("metadata.name"),
		},
		{
			name: "ends with -",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": "sdfasa-",
			}},
			err: ErrNameNotRFC1035Compatible("metadata.name"),
		},
		{
			name: "contains uppercase",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": "AAsdf",
			}},
			err: ErrNameNotRFC1035Compatible("metadata.name"),
		},
		{
			name: "valid",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": "amsdf-sllla",
			}},
			err: nil,
		},
		{
			name: "dbCluster name wrong format",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": make(map[string]string),
			}},
			err: errDBCNameWrongFormat,
		},
		{
			name: "dbCluster name too long",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": "a123456789a123456789a12",
			}},
			err: ErrNameTooLong("metadata.name"),
		},
	}

	for _, tc := range cases {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := validateCreateDatabaseClusterRequest(c.value)
			if c.err == nil {
				require.NoError(t, err)
				return
			}
			assert.Equal(t, c.err.Error(), err.Error())
		})
	}
}

func TestValidateProxy(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		engineType string
		proxyType  string
		err        error
	}{
		{
			name:       "PXC with mongos",
			engineType: "pxc",
			proxyType:  "mongos",
			err:        errUnsupportedPXCProxy,
		},
		{
			name:       "PXC with pgbouncer",
			engineType: "pxc",
			proxyType:  "pgbouncer",
			err:        errUnsupportedPXCProxy,
		},
		{
			name:       "PXC with haproxy",
			engineType: "pxc",
			proxyType:  "haproxy",
			err:        nil,
		},
		{
			name:       "PXC with proxysql",
			engineType: "pxc",
			proxyType:  "proxysql",
			err:        nil,
		},
		{
			name:       "psmdb with mongos",
			engineType: "psmdb",
			proxyType:  "mongos",
			err:        nil,
		},
		{
			name:       "psmdb with pgbouncer",
			engineType: "psmdb",
			proxyType:  "pgbouncer",
			err:        errUnsupportedPSMDBProxy,
		},
		{
			name:       "psmdb with haproxy",
			engineType: "psmdb",
			proxyType:  "haproxy",
			err:        errUnsupportedPSMDBProxy,
		},
		{
			name:       "psmdb with proxysql",
			engineType: "psmdb",
			proxyType:  "proxysql",
			err:        errUnsupportedPSMDBProxy,
		},
		{
			name:       "postgresql with mongos",
			engineType: "postgresql",
			proxyType:  "mongos",
			err:        errUnsupportedPGProxy,
		},
		{
			name:       "postgresql with pgbouncer",
			engineType: "postgresql",
			proxyType:  "pgbouncer",
			err:        nil,
		},
		{
			name:       "postgresql with haproxy",
			engineType: "postgresql",
			proxyType:  "haproxy",
			err:        errUnsupportedPGProxy,
		},
		{
			name:       "postgresql with proxysql",
			engineType: "postgresql",
			proxyType:  "proxysql",
			err:        errUnsupportedPGProxy,
		},
	}
	for _, tc := range cases {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := validateProxy(DatabaseClusterSpecEngineType(c.engineType), c.proxyType)
			if c.err == nil {
				require.NoError(t, err)
				return
			}
			assert.Equal(t, c.err.Error(), err.Error())
		})
	}
}

func TestContainsVersion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		version  string
		versions []string
		result   bool
	}{
		{
			version:  "1",
			versions: []string{},
			result:   false,
		},
		{
			version:  "1",
			versions: []string{"1", "2"},
			result:   true,
		},
		{
			version:  "1",
			versions: []string{"1"},
			result:   true,
		},
		{
			version:  "1",
			versions: []string{"12", "23"},
			result:   false,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.version, func(t *testing.T) {
			t.Parallel()
			res := containsVersion(tc.version, tc.versions)
			assert.Equal(t, res, tc.result)
		})
	}
}

func TestValidateVersion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		version *string
		engine  *everestv1alpha1.DatabaseEngine
		err     error
	}{
		{
			name:    "empty version is allowed",
			version: nil,
			engine:  nil,
			err:     nil,
		},
		{
			name:    "shall exist in availableVersions",
			version: pointer.ToString("8.0.32"),
			engine: &everestv1alpha1.DatabaseEngine{
				Status: everestv1alpha1.DatabaseEngineStatus{
					AvailableVersions: everestv1alpha1.Versions{
						Engine: everestv1alpha1.ComponentsMap{
							"8.0.32": &everestv1alpha1.Component{},
						},
					},
				},
			},
			err: nil,
		},
		{
			name:    "shall not exist in availableVersions",
			version: pointer.ToString("8.0.32"),
			engine: &everestv1alpha1.DatabaseEngine{
				Status: everestv1alpha1.DatabaseEngineStatus{
					AvailableVersions: everestv1alpha1.Versions{
						Engine: everestv1alpha1.ComponentsMap{
							"8.0.31": &everestv1alpha1.Component{},
						},
					},
				},
			},
			err: errors.New("8.0.32 is not in available versions list"),
		},
		{
			name:    "shall exist in allowedVersions",
			version: pointer.ToString("8.0.32"),
			engine: &everestv1alpha1.DatabaseEngine{
				Spec: everestv1alpha1.DatabaseEngineSpec{
					Type:            "pxc",
					AllowedVersions: []string{"8.0.32"},
				},
			},
			err: nil,
		},
		{
			name:    "shall not exist in allowedVersions",
			version: pointer.ToString("8.0.32"),
			engine: &everestv1alpha1.DatabaseEngine{
				Spec: everestv1alpha1.DatabaseEngineSpec{
					Type:            "pxc",
					AllowedVersions: []string{"8.0.31"},
				},
			},
			err: errors.New("using 8.0.32 version for pxc is not allowed"),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateVersion(tc.version, tc.engine)
			if tc.err == nil {
				require.NoError(t, err)
				return
			}
			assert.Equal(t, err.Error(), tc.err.Error())
		})
	}
}

func TestValidateBackupSpec(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		cluster []byte
		err     error
	}{
		{
			name:    "empty backup is allowed",
			cluster: []byte(`{"spec": {"backup": null}}`),
			err:     nil,
		},
		{
			name:    "disabled backup is allowed",
			cluster: []byte(`{"spec": {"backup": {"enabled": false}}}`),
			err:     nil,
		},
		{
			name:    "errNoSchedules",
			cluster: []byte(`{"spec": {"backup": {"enabled": true}}}`),
			err:     errNoSchedules,
		},
		{
			name:    "errNoNameInSchedule",
			cluster: []byte(`{"spec": {"backup": {"enabled": true, "schedules": [{"enabled": true}]}}}`),
			err:     errNoNameInSchedule,
		},
		{
			name:    "errNoBackupStorageName",
			cluster: []byte(`{"spec": {"backup": {"enabled": true, "schedules": [{"enabled": true, "name": "name"}]}}}`),
			err:     errScheduleNoBackupStorageName,
		},
		{
			name:    "valid spec",
			cluster: []byte(`{"spec": {"backup": {"enabled": true, "schedules": [{"enabled": true, "name": "name", "backupStorageName": "some"}]}}}`),
			err:     nil,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cluster := &DatabaseCluster{}
			err := json.Unmarshal(tc.cluster, cluster)
			require.NoError(t, err)
			err = validateBackupSpec(cluster)
			if tc.err == nil {
				require.NoError(t, err)
				return
			}
			assert.Equal(t, err.Error(), tc.err.Error())
		})
	}
}

func TestValidateBackupStoragesFor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		cluster []byte
		storage []byte
		err     error
	}{
		{
			name:    "errPSMDBMultipleStorages",
			cluster: []byte(`{"spec": {"backup": {"enabled": true, "schedules": [{"enabled": true, "name": "name", "backupStorageName": "storage1"}, {"enabled": true, "name": "name2", "backupStorageName": "storage2"}]}, "engine": {"type": "psmdb"}}}`),
			storage: []byte(`{"spec": {"type": "s3"}}`),
			err:     errPSMDBMultipleStorages,
		},
		{
			name:    "errPSMDBViolateActiveStorage",
			cluster: []byte(`{"status": {"activeStorage": "storage1"}, "spec": {"backup": {"enabled": true, "schedules": [{"enabled": true, "name": "otherName", "backupStorageName": "storage2"}]}, "engine": {"type": "psmdb"}}}`),
			storage: []byte(`{"spec": {"type": "s3"}}`),
			err:     errPSMDBViolateActiveStorage,
		},
		{
			name:    "no errPSMDBViolateActiveStorage",
			cluster: []byte(`{"status": {"activeStorage": ""}, "spec": {"backup": {"enabled": true, "schedules": [{"enabled": true, "name": "otherName", "backupStorageName": "storage2"}]}, "engine": {"type": "psmdb"}}}`),
			storage: []byte(`{"spec": {"type": "s3"}}`),
			err:     nil,
		},
		{
			name:    "errPXCPitrS3Only",
			cluster: []byte(`{"status":{},"spec": {"backup": {"enabled": true, "pitr": {"enabled": true, "backupStorageName": "storage"}, "schedules": [{"enabled": true, "name": "otherName", "backupStorageName": "storage"}]}, "engine": {"type": "pxc"}}}`),
			storage: []byte(`{"spec": {"type": "azure"}}`),
			err:     errPXCPitrS3Only,
		},
		{
			name:    "errPitrNoBackupStorageName",
			cluster: []byte(`{"status":{},"spec": {"backup": {"enabled": true, "pitr": {"enabled": true}, "schedules": [{"enabled": true, "name": "otherName", "backupStorageName": "storage"}]}, "engine": {"type": "pxc"}}}`),
			storage: []byte(`{"spec": {"type": "s3"}}`),
			err:     errPitrNoBackupStorageName,
		},
		{
			name:    "valid",
			cluster: []byte(`{"status":{},"spec": {"backup": {"enabled": true, "pitr": {"enabled": true, "backupStorageName": "storage2"}, "schedules": [{"enabled": true, "name": "otherName", "backupStorageName": "storage"}]}, "engine": {"type": "pxc"}}}`),
			storage: []byte(`{"spec": {"type": "s3"}}`),
			err:     nil,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cluster := &DatabaseCluster{}
			err := json.Unmarshal(tc.cluster, cluster)
			require.NoError(t, err)

			storage := &everestv1alpha1.BackupStorage{}
			err = json.Unmarshal(tc.storage, storage)
			require.NoError(t, err)

			err = validateBackupStoragesFor(
				context.Background(),
				cluster,
				func(ctx context.Context, s string) (*everestv1alpha1.BackupStorage, error) { return storage, nil },
			)
			if tc.err == nil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Equal(t, err.Error(), tc.err.Error())
		})
	}
}

func TestValidatePitrSpec(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		cluster []byte
		err     error
	}{
		{
			name:    "valid spec pitr enabled",
			cluster: []byte(`{"spec": {"backup": {"enabled": true, "pitr": {"enabled": true, "backupStorageName": "name"}}}}`),
			err:     nil,
		},
		{
			name:    "valid spec pitr disabled",
			cluster: []byte(`{"spec": {"backup": {"enabled": true, "pitr": {"enabled": false}}}}`),
			err:     nil,
		},
		{
			name:    "valid spec no pitr",
			cluster: []byte(`{"spec": {"backup": {"enabled": true}}}`),
			err:     nil,
		},
		{
			name:    "no backup storage pxc",
			cluster: []byte(`{"spec": {"backup": {"enabled": true, "pitr": {"enabled": true}}, "engine": {"type": "pxc"}}}`),
			err:     errPitrNoBackupStorageName,
		},
		{
			name:    "no backup storage psmdb",
			cluster: []byte(`{"spec": {"backup": {"enabled": true, "pitr": {"enabled": true}}, "engine": {"type": "psmdb"}}}`),
			err:     nil,
		},
		{
			name:    "no backup storage pg",
			cluster: []byte(`{"spec": {"backup": {"enabled": true, "pitr": {"enabled": true}}, "engine": {"type": "postgresql"}}}`),
			err:     nil,
		},
		{
			name:    "zero upload interval",
			cluster: []byte(`{"spec": {"backup": {"enabled": true, "pitr": {"enabled": true, "backupStorageName": "name", "uploadIntervalSec": 0}}}}`),
			err:     errPitrUploadInterval,
		},
		{
			name:    "negative upload interval",
			cluster: []byte(`{"spec": {"backup": {"enabled": true, "pitr": {"enabled": true, "backupStorageName": "name", "uploadIntervalSec": -100}}}}`),
			err:     errPitrUploadInterval,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cluster := &DatabaseCluster{}
			err := json.Unmarshal(tc.cluster, cluster)
			require.NoError(t, err)
			err = validatePitrSpec(cluster)
			if tc.err == nil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Equal(t, err.Error(), tc.err.Error())
		})
	}
}

func TestValidateResourceLimits(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		cluster []byte
		err     error
	}{
		{
			name:    "success",
			cluster: []byte(`{"spec": {"engine": {"resources": {"cpu": "600m", "memory":"1G"}, "storage": {"size": "2G"}}}}`),
			err:     nil,
		},
		{
			name:    "errNoResourceDefined",
			cluster: []byte(`{"spec": {"engine": {"resources":null, "storage": {"size": "2G"}}}}`),
			err:     errNoResourceDefined,
		},
		{
			name:    "Not enough CPU",
			cluster: []byte(`{"spec": {"engine": {"resources": {"cpu": null, "memory":"1G"}, "storage": {"size": "2G"}}}}`),
			err:     errNotEnoughCPU,
		},
		{
			name:    "Not enough memory",
			cluster: []byte(`{"spec": {"engine": {"resources": {"cpu": "600m", "memory":null}, "storage": {"size": "2G"}}}}`),
			err:     errNotEnoughMemory,
		},
		{
			name:    "No int64 for CPU",
			cluster: []byte(`{"spec": {"engine": {"resources": {"cpu": 6000, "memory": "1G"}, "storage": {"size": "2G"}}}}`),
			err:     errInt64NotSupported,
		},
		{
			name:    "No int64 for Memory",
			cluster: []byte(`{"spec": {"engine": {"resources": {"cpu": "600m", "memory": 1000000}, "storage": {"size": "2G"}}}}`),
			err:     errInt64NotSupported,
		},
		{
			name:    "No int64 for storage",
			cluster: []byte(`{"spec": {"engine": {"resources": {"cpu": "600m", "memory": "1G"}, "storage": {"size": 20000}}}}`),
			err:     errInt64NotSupported,
		},
		{
			name:    "not enough disk size",
			cluster: []byte(`{"spec": {"engine": {"resources": {"cpu": "600m", "memory": "1G"}, "storage": {"size": "512M"}}}}`),
			err:     errNotEnoughDiskSize,
		},
		{
			name:    "not enough CPU",
			cluster: []byte(`{"spec": {"engine": {"resources": {"cpu": "200m", "memory": "1G"}, "storage": {"size": "2G"}}}}`),
			err:     errNotEnoughCPU,
		},
		{
			name:    "not enough Mem",
			cluster: []byte(`{"spec": {"engine": {"resources": {"cpu": "600m", "memory": "400M"}, "storage": {"size": "2G"}}}}`),
			err:     errNotEnoughMemory,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cluster := &DatabaseCluster{}
			err := json.Unmarshal(tc.cluster, cluster)
			require.NoError(t, err)
			err = validateResourceLimits(cluster)
			if tc.err == nil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Equal(t, err.Error(), tc.err.Error())
		})
	}
}

func TestValidateDataSource(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		cluster []byte
		err     error
	}{
		{
			name:    "err none of the data source specified",
			cluster: []byte(`{}`),
			err:     errDataSourceConfig,
		},
		{
			name:    "err both of the data source specified",
			cluster: []byte(`{"dbClusterBackupName":"some-backup", "backupSource": {"backupStorageName":"some-name","path":"some-path"}}`),
			err:     errDataSourceConfig,
		},
		{
			name:    "err no date in pitr",
			cluster: []byte(`{"dbClusterBackupName":"some-backup","pitr":{}}`),
			err:     errDataSourceNoPitrDateSpecified,
		},
		{
			name:    "wrong pitr date format",
			cluster: []byte(`{"dbClusterBackupName":"some-backup","pitr":{"date":"2006-06-07 14:06:07"}}`),
			err:     errDataSourceWrongDateFormat,
		},
		{
			name:    "wrong pitr date format",
			cluster: []byte(`{"dbClusterBackupName":"some-backup","pitr":{"date":""}}`),
			err:     errDataSourceWrongDateFormat,
		},
		{
			name:    "correct minimal",
			cluster: []byte(`{"dbClusterBackupName":"some-backup","pitr":{"date":"2006-06-07T14:06:07Z"}}`),
			err:     nil,
		},
		{
			name:    "correct with pitr type",
			cluster: []byte(`{"dbClusterBackupName":"some-backup","pitr":{"type":"date","date":"2006-06-07T14:06:07Z"}}`),
			err:     nil,
		},
		{
			name:    "unsupported pitr type",
			cluster: []byte(`{"backupSource":{"backupStorageName":"some-name","path":"some-path"},"pitr":{"type":"latest"}}`),
			err:     errUnsupportedPitrType,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dsDB := &dataSourceStruct{}
			err := json.Unmarshal(tc.cluster, dsDB)
			require.NoError(t, err)
			err = validateDataSource(*dsDB)
			if tc.err == nil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Equal(t, err.Error(), tc.err.Error())
		})
	}
}

func TestValidatePGReposForAPIDB(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name           string
		cluster        []byte
		getBackupsFunc func(ctx context.Context, options metav1.ListOptions) (*everestv1alpha1.DatabaseClusterBackupList, error)
		err            error
	}{
		{
			name:    "ok: no schedules no backups",
			cluster: []byte(`{"metaData":{"name":"some"}}`),
			getBackupsFunc: func(ctx context.Context, options metav1.ListOptions) (*everestv1alpha1.DatabaseClusterBackupList, error) {
				return &everestv1alpha1.DatabaseClusterBackupList{
					Items: []everestv1alpha1.DatabaseClusterBackup{},
				}, nil
			},
			err: nil,
		},
		{
			name:    "ok: 2 schedules 2 backups with the same storages",
			cluster: []byte(`{"metaData":{"name":"some"},"spec":{"backup":{"schedules":[{"backupStorageName":"bs1"},{"backupStorageName":"bs2"}]}}}`),
			getBackupsFunc: func(ctx context.Context, options metav1.ListOptions) (*everestv1alpha1.DatabaseClusterBackupList, error) {
				return &everestv1alpha1.DatabaseClusterBackupList{
					Items: []everestv1alpha1.DatabaseClusterBackup{
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs1"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs2"}},
					},
				}, nil
			},
			err: nil,
		},
		{
			name:    "ok: 3 schedules",
			cluster: []byte(`{"metaData":{"name":"some"},"spec":{"backup":{"schedules":[{"backupStorageName":"bs1"},{"backupStorageName":"bs2"},{"backupStorageName":"bs3"}]}}}`),
			getBackupsFunc: func(ctx context.Context, options metav1.ListOptions) (*everestv1alpha1.DatabaseClusterBackupList, error) {
				return &everestv1alpha1.DatabaseClusterBackupList{
					Items: []everestv1alpha1.DatabaseClusterBackup{},
				}, nil
			},
			err: nil,
		},
		{
			name:    "ok: 3 backups with different storages",
			cluster: []byte(`{"metaData":{"name":"some"}}`),
			getBackupsFunc: func(ctx context.Context, options metav1.ListOptions) (*everestv1alpha1.DatabaseClusterBackupList, error) {
				return &everestv1alpha1.DatabaseClusterBackupList{
					Items: []everestv1alpha1.DatabaseClusterBackup{
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs1"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs2"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs3"}},
					},
				}, nil
			},
			err: nil,
		},
		{
			name:    "ok: 5 backups with repeating storages",
			cluster: []byte(`{"metaData":{"name":"some"}}`),
			getBackupsFunc: func(ctx context.Context, options metav1.ListOptions) (*everestv1alpha1.DatabaseClusterBackupList, error) {
				return &everestv1alpha1.DatabaseClusterBackupList{
					Items: []everestv1alpha1.DatabaseClusterBackup{
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs1"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs2"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs3"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs1"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs2"}},
					},
				}, nil
			},
			err: nil,
		},
		{
			name:    "error: 4 backups with different storages",
			cluster: []byte(`{"metaData":{"name":"some"}}`),
			getBackupsFunc: func(ctx context.Context, options metav1.ListOptions) (*everestv1alpha1.DatabaseClusterBackupList, error) {
				return &everestv1alpha1.DatabaseClusterBackupList{
					Items: []everestv1alpha1.DatabaseClusterBackup{
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs1"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs2"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs3"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs4"}},
					},
				}, nil
			},
			err: errTooManyPGStorages,
		},
		{
			name:    "ok: 4 backups with same storages",
			cluster: []byte(`{"metaData":{"name":"some"}}`),
			getBackupsFunc: func(ctx context.Context, options metav1.ListOptions) (*everestv1alpha1.DatabaseClusterBackupList, error) {
				return &everestv1alpha1.DatabaseClusterBackupList{
					Items: []everestv1alpha1.DatabaseClusterBackup{
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs1"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs2"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs2"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs1"}},
					},
				}, nil
			},
			err: nil,
		},
		{
			name:    "error: 4 schedules",
			cluster: []byte(`{"metaData":{"name":"some"},"spec":{"backup":{"schedules":[{"backupStorageName":"bs1"},{"backupStorageName":"bs2"},{"backupStorageName":"bs3"},{"backupStorageName":"bs4"}]}}}`),
			getBackupsFunc: func(ctx context.Context, options metav1.ListOptions) (*everestv1alpha1.DatabaseClusterBackupList, error) {
				return &everestv1alpha1.DatabaseClusterBackupList{
					Items: []everestv1alpha1.DatabaseClusterBackup{},
				}, nil
			},
			err: errTooManyPGSchedules,
		},
		{
			name:    "error: 2 schedules 2 backups with different storages",
			cluster: []byte(`{"metaData":{"name":"some"},"spec":{"backup":{"schedules":[{"backupStorageName":"bs1"},{"backupStorageName":"bs2"}]}}}`),
			getBackupsFunc: func(ctx context.Context, options metav1.ListOptions) (*everestv1alpha1.DatabaseClusterBackupList, error) {
				return &everestv1alpha1.DatabaseClusterBackupList{
					Items: []everestv1alpha1.DatabaseClusterBackup{
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs3"}},
						{Spec: everestv1alpha1.DatabaseClusterBackupSpec{BackupStorageName: "bs4"}},
					},
				}, nil
			},
			err: errTooManyPGStorages,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := &DatabaseCluster{}
			err := json.Unmarshal(tc.cluster, db)
			require.NoError(t, err)
			err = validatePGReposForAPIDB(context.Background(), db, tc.getBackupsFunc)
			if tc.err == nil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Equal(t, tc.err.Error(), err.Error())
		})
	}
}
