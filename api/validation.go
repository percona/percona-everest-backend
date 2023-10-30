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
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/AlekSi/pointer"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo/v4"
	everestv1alpha1 "github.com/percona/everest-operator/api/v1alpha1"
	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/percona/percona-everest-backend/cmd/config"
)

const (
	pxcDeploymentName   = "percona-xtradb-cluster-operator"
	psmdbDeploymentName = "percona-server-mongodb-operator"
	pgDeploymentName    = "percona-postgresql-operator"
)

var (
	minStorageQuantity = resource.MustParse("1G")   //nolint:gochecknoglobals
	minCPUQuantity     = resource.MustParse("600m") //nolint:gochecknoglobals
	minMemQuantity     = resource.MustParse("512M") //nolint:gochecknoglobals

	errDBCEmptyMetadata      = errors.New("databaseCluster's Metadata should not be empty")
	errDBCNameEmpty          = errors.New("databaseCluster's metadata.name should not be empty")
	errDBCNameWrongFormat    = errors.New("databaseCluster's metadata.name should be a string")
	errNotEnoughMemory       = fmt.Errorf("memory limits should be above %s", minMemQuantity.String())
	errInt64NotSupported     = errors.New("specifying resources using int64 data type is not supported. Please use string format for that")
	errNotEnoughCPU          = fmt.Errorf("CPU limits should be above %s", minCPUQuantity.String())
	errNotEnoughDiskSize     = fmt.Errorf("storage size should be above %s", minStorageQuantity.String())
	errUnsupportedPXCProxy   = errors.New("you can use either HAProxy or Proxy SQL for PXC clusters")
	errUnsupportedPGProxy    = errors.New("you can use only PGBouncer as a proxy type for Postgres clusters")
	errUnsupportedPSMDBProxy = errors.New("you can use only Mongos as a proxy type for MongoDB clusters")
	errNoSchedules           = errors.New("please specify at least one backup schedule")
	errNoNameInSchedule      = errors.New("'name' field for the backup schedules cannot be empty")
	errNoBackupStorageName   = errors.New("'backupStorageName' field cannot be empty when schedule is enabled")
	errNoResourceDefined     = errors.New("please specify resource limits for the cluster")
	//nolint:gochecknoglobals
	operatorEngine = map[everestv1alpha1.EngineType]string{
		everestv1alpha1.DatabaseEnginePXC:        pxcDeploymentName,
		everestv1alpha1.DatabaseEnginePSMDB:      psmdbDeploymentName,
		everestv1alpha1.DatabaseEnginePostgresql: pgDeploymentName,
	}
)

// ErrNameNotRFC1035Compatible when the given fieldName doesn't contain RFC 1035 compatible string.
func ErrNameNotRFC1035Compatible(fieldName string) error {
	return fmt.Errorf(`'%s' is not RFC 1035 compatible. The name should contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric character`,
		fieldName,
	)
}

// ErrNameTooLong when the given fieldName is longer than expected.
func ErrNameTooLong(fieldName string) error {
	return fmt.Errorf("'%s' can be at most 22 characters long", fieldName)
}

// ErrCreateStorageNotSupported appears when trying to create a storage of a type that is not supported.
func ErrCreateStorageNotSupported(storageType string) error {
	return fmt.Errorf("creating storage is not implemented for '%s'", storageType)
}

// ErrUpdateStorageNotSupported appears when trying to update a storage of a type that is not supported.
func ErrUpdateStorageNotSupported(storageType string) error {
	return fmt.Errorf("updating storage is not implemented for '%s'", storageType)
}

// ErrInvalidURL when the given fieldName contains invalid URL.
func ErrInvalidURL(fieldName string) error {
	return fmt.Errorf("'%s' is an invalid URL", fieldName)
}

// validates names to be RFC-1035 compatible  https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#rfc-1035-label-names
func validateRFC1035(s, name string) error {
	// We are diverging from the RFC1035 spec in regards to the length of the
	// name because the PXC operator limits the name of the cluster to 22.
	if len(s) > 22 {
		return ErrNameTooLong(name)
	}

	rfc1035Regex := "^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$"
	re := regexp.MustCompile(rfc1035Regex)
	if !re.MatchString(s) {
		return ErrNameNotRFC1035Compatible(name)
	}

	return nil
}

func validateURL(urlStr string) bool {
	_, err := url.ParseRequestURI(urlStr)
	return err == nil
}

func validateStorageAccessByCreate(ctx context.Context, params CreateBackupStorageParams, l *zap.SugaredLogger) error {
	switch params.Type {
	case CreateBackupStorageParamsTypeS3:
		return s3Access(l, params.Url, params.AccessKey, params.SecretKey, params.BucketName, params.Region)
	case CreateBackupStorageParamsTypeAzure:
		return azureAccess(ctx, l, params.AccessKey, params.SecretKey, params.BucketName)
	default:
		return ErrCreateStorageNotSupported(string(params.Type))
	}
}

func s3Access(l *zap.SugaredLogger, endpoint *string, accessKey, secretKey, bucketName, region string) error {
	if config.Debug {
		return nil
	}

	if endpoint != nil && *endpoint == "" {
		endpoint = nil
	}

	// Create a new session with the provided credentials
	sess, err := session.NewSession(&aws.Config{
		Endpoint:    endpoint,
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		l.Error(err)
		return errors.New("could not initialize S3 session")
	}

	// Create a new S3 client with the session
	svc := s3.New(sess)

	_, err = svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		l.Error(err)
		return errors.New("unable to connect to s3. Check your credentials")
	}

	testKey := "everest-write-test"
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Body:   bytes.NewReader([]byte{}),
		Key:    aws.String(testKey),
	})
	if err != nil {
		l.Error(err)
		return errors.New("could not write to S3 bucket")
	}

	_, err = svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(testKey),
	})
	if err != nil {
		l.Error(err)
		return errors.New("could not read from S3 bucket")
	}

	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(testKey),
	})
	if err != nil {
		l.Error(err)
		return errors.New("could not delete an object from S3 bucket")
	}

	return nil
}

func azureAccess(ctx context.Context, l *zap.SugaredLogger, accountName, accountKey, containerName string) error {
	if config.Debug {
		return nil
	}

	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		l.Error(err)
		return errors.New("could not initialize Azure credentials")
	}

	client, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", url.PathEscape(accountName)), cred, nil)
	if err != nil {
		l.Error(err)
		return errors.New("could not initialize Azure client")
	}

	pager := client.NewListBlobsFlatPager(containerName, nil)
	if pager.More() {
		if _, err := pager.NextPage(ctx); err != nil {
			l.Error(err)
			return errors.New("could not list blobs in Azure container")
		}
	}

	blobName := "everest-test-blob"
	if _, err = client.UploadBuffer(ctx, containerName, blobName, []byte{}, nil); err != nil {
		l.Error(err)
		return errors.New("could not write to Azure container")
	}

	if _, err = client.DownloadBuffer(ctx, containerName, blobName, []byte{}, nil); err != nil {
		l.Error(err)
		return errors.New("could not read from Azure container")
	}

	if _, err = client.DeleteBlob(ctx, containerName, blobName, nil); err != nil {
		l.Error(err)
		return errors.New("could not delete a blob from Azure container")
	}

	return nil
}

func validateUpdateBackupStorageRequest(ctx echo.Context, bs *everestv1alpha1.BackupStorage, l *zap.SugaredLogger) (*UpdateBackupStorageParams, error) {
	var params UpdateBackupStorageParams
	if err := ctx.Bind(&params); err != nil {
		return nil, err
	}

	if params.Url != nil {
		if ok := validateURL(*params.Url); !ok {
			err := ErrInvalidURL("url")
			return nil, err
		}
	}
	if params.AccessKey != nil && params.SecretKey == nil {
		return nil, errors.New("cannot update access key without secret key")
	}
	if params.SecretKey != nil && params.AccessKey == nil {
		return nil, errors.New("cannot update secret key without access key")
	}

	bucketName := bs.Spec.Bucket
	if params.BucketName != nil {
		bucketName = *params.BucketName
	}
	switch string(bs.Spec.Type) {
	case string(BackupStorageTypeS3):
		if params.Region != nil && *params.Region == "" {
			return nil, errors.New("region is required when using S3 storage type")
		}
		if err := s3Access(l, &bs.Spec.EndpointURL, *params.AccessKey, *params.SecretKey, bucketName, bs.Spec.Region); err != nil {
			return nil, err
		}
	case string(BackupStorageTypeAzure):
		if err := azureAccess(ctx.Request().Context(), l, *params.AccessKey, *params.SecretKey, bucketName); err != nil {
			return nil, err
		}
	default:
		return nil, ErrUpdateStorageNotSupported(string(bs.Spec.Type))
	}

	return &params, nil
}

func validateCreateBackupStorageRequest(ctx echo.Context, l *zap.SugaredLogger) (*CreateBackupStorageParams, error) {
	var params CreateBackupStorageParams
	if err := ctx.Bind(&params); err != nil {
		return nil, err
	}

	if err := validateRFC1035(params.Name, "name"); err != nil {
		return nil, err
	}

	if params.Url != nil {
		if ok := validateURL(*params.Url); !ok {
			err := ErrInvalidURL("url")
			return nil, err
		}
	}

	if params.Type == CreateBackupStorageParamsTypeS3 {
		if params.Region == "" {
			return nil, errors.New("region is required when using S3 storage type")
		}
	}

	// check data access
	if err := validateStorageAccessByCreate(ctx.Request().Context(), params, l); err != nil {
		l.Error(err)
		return nil, err
	}

	return &params, nil
}

func validateCreateMonitoringInstanceRequest(ctx echo.Context) (*CreateMonitoringInstanceJSONRequestBody, error) {
	var params CreateMonitoringInstanceJSONRequestBody
	if err := ctx.Bind(&params); err != nil {
		return nil, err
	}

	if err := validateRFC1035(params.Name, "name"); err != nil {
		return nil, err
	}

	if ok := validateURL(params.Url); !ok {
		return nil, ErrInvalidURL("url")
	}

	switch params.Type {
	case MonitoringInstanceCreateParamsTypePmm:
		if params.Pmm == nil {
			return nil, fmt.Errorf("pmm key is required for type %s", params.Type)
		}

		if params.Pmm.ApiKey == "" && params.Pmm.User == "" && params.Pmm.Password == "" {
			return nil, errors.New("one of pmm.apiKey, pmm.user or pmm.password fields is required")
		}
	default:
		return nil, fmt.Errorf("monitoring type %s is not supported", params.Type)
	}

	return &params, nil
}

func validateUpdateMonitoringInstanceRequest(ctx echo.Context) (*UpdateMonitoringInstanceJSONRequestBody, error) {
	var params UpdateMonitoringInstanceJSONRequestBody
	if err := ctx.Bind(&params); err != nil {
		return nil, err
	}

	if params.Url != "" {
		if ok := validateURL(params.Url); !ok {
			err := ErrInvalidURL("url")
			return nil, err
		}
	}

	if err := validateUpdateMonitoringInstanceType(params); err != nil {
		return nil, err
	}

	if params.Pmm != nil && params.Pmm.ApiKey == "" && params.Pmm.User == "" && params.Pmm.Password == "" {
		return nil, errors.New("one of pmm.apiKey, pmm.user or pmm.password fields is required")
	}

	return &params, nil
}

func validateUpdateMonitoringInstanceType(params UpdateMonitoringInstanceJSONRequestBody) error {
	switch params.Type {
	case "":
		return nil
	case MonitoringInstanceUpdateParamsTypePmm:
		if params.Pmm == nil {
			return fmt.Errorf("pmm key is required for type %s", params.Type)
		}
	default:
		return errors.New("this monitoring type is not supported")
	}

	return nil
}

func validateCreateDatabaseClusterRequest(dbc DatabaseCluster) error {
	if dbc.Metadata == nil {
		return errDBCEmptyMetadata
	}

	md := *dbc.Metadata
	name, ok := md["name"]
	if !ok {
		return errDBCNameEmpty
	}

	strName, ok := name.(string)
	if !ok {
		return errDBCNameWrongFormat
	}

	return validateRFC1035(strName, "metadata.name")
}

func (e *EverestServer) validateDBClusterAccess(ctx echo.Context, dbClusterName string) error {
	_, err := e.kubeClient.GetDatabaseCluster(ctx.Request().Context(), dbClusterName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(fmt.Sprintf("DatabaseCluster '%s' is not found", dbClusterName))})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	return nil
}

func (e *EverestServer) validateDatabaseClusterCR(ctx echo.Context, databaseCluster *DatabaseCluster) error {
	if err := validateCreateDatabaseClusterRequest(*databaseCluster); err != nil {
		return err
	}

	engineName, ok := operatorEngine[everestv1alpha1.EngineType(databaseCluster.Spec.Engine.Type)]
	if !ok {
		return errors.New("unsupported database engine")
	}
	engine, err := e.kubeClient.GetDatabaseEngine(ctx.Request().Context(), engineName)
	if err != nil {
		return err
	}
	if err := validateVersion(databaseCluster.Spec.Engine.Version, engine); err != nil {
		return err
	}
	if databaseCluster.Spec.Proxy != nil && databaseCluster.Spec.Proxy.Type != nil {
		if err := validateProxy(databaseCluster.Spec.Engine.Type, string(*databaseCluster.Spec.Proxy.Type)); err != nil {
			return err
		}
	}
	if err := validateBackupSpec(databaseCluster); err != nil {
		return err
	}
	return validateResourceLimits(databaseCluster)
}

func validateVersion(version *string, engine *everestv1alpha1.DatabaseEngine) error {
	if version != nil {
		if len(engine.Spec.AllowedVersions) != 0 {
			if !containsVersion(*version, engine.Spec.AllowedVersions) {
				return fmt.Errorf("using %s version for %s is not allowed", *version, engine.Spec.Type)
			}
			return nil
		}
		if _, ok := engine.Status.AvailableVersions.Engine[*version]; !ok {
			return fmt.Errorf("%s is not in available versions list", *version)
		}
	}
	return nil
}

func containsVersion(version string, versions []string) bool {
	if version == "" {
		return true
	}
	for _, allowedVersion := range versions {
		if version == allowedVersion {
			return true
		}
	}
	return false
}

func validateProxy(engineType, proxyType string) error {
	if engineType == string(everestv1alpha1.DatabaseEnginePXC) {
		if proxyType != string(everestv1alpha1.ProxyTypeProxySQL) && proxyType != string(everestv1alpha1.ProxyTypeHAProxy) {
			return errUnsupportedPXCProxy
		}
	}

	if engineType == string(everestv1alpha1.DatabaseEnginePostgresql) && proxyType != string(everestv1alpha1.ProxyTypePGBouncer) {
		return errUnsupportedPGProxy
	}
	if engineType == string(everestv1alpha1.DatabaseEnginePSMDB) && proxyType != string(everestv1alpha1.ProxyTypeMongos) {
		return errUnsupportedPSMDBProxy
	}
	return nil
}

func validateBackupSpec(cluster *DatabaseCluster) error {
	if cluster.Spec.Backup == nil {
		return nil
	}
	if !cluster.Spec.Backup.Enabled {
		return nil
	}
	if cluster.Spec.Backup.Schedules == nil {
		return errNoSchedules
	}

	for _, schedule := range *cluster.Spec.Backup.Schedules {
		if schedule.Name == "" {
			return errNoNameInSchedule
		}
		if schedule.Enabled && schedule.BackupStorageName == "" {
			return errNoBackupStorageName
		}
	}
	return nil
}

func validateResourceLimits(cluster *DatabaseCluster) error {
	if err := ensureNonEmptyResources(cluster); err != nil {
		return err
	}
	if err := validateCPU(cluster); err != nil {
		return err
	}
	if err := validateMemory(cluster); err != nil {
		return err
	}
	return validateStorageSize(cluster)
}

func ensureNonEmptyResources(cluster *DatabaseCluster) error {
	if cluster.Spec.Engine.Resources == nil {
		return errNoResourceDefined
	}
	if cluster.Spec.Engine.Resources.Cpu == nil {
		return errNotEnoughCPU
	}
	if cluster.Spec.Engine.Resources.Memory == nil {
		return errNotEnoughMemory
	}
	return nil
}

func validateCPU(cluster *DatabaseCluster) error {
	cpuStr, err := cluster.Spec.Engine.Resources.Cpu.AsDatabaseClusterSpecEngineResourcesCpu1()
	if err == nil {
		cpu, err := resource.ParseQuantity(cpuStr)
		if err != nil {
			return err
		}
		if cpu.Cmp(minCPUQuantity) == -1 {
			return errNotEnoughCPU
		}
	}
	_, err = cluster.Spec.Engine.Resources.Cpu.AsDatabaseClusterSpecEngineResourcesCpu0()
	if err == nil {
		return errInt64NotSupported
	}
	return nil
}

func validateMemory(cluster *DatabaseCluster) error {
	_, err := cluster.Spec.Engine.Resources.Memory.AsDatabaseClusterSpecEngineResourcesMemory0()
	if err == nil {
		return errInt64NotSupported
	}
	memStr, err := cluster.Spec.Engine.Resources.Memory.AsDatabaseClusterSpecEngineResourcesMemory1()
	if err == nil {
		mem, err := resource.ParseQuantity(memStr)
		if err != nil {
			return err
		}
		if mem.Cmp(minMemQuantity) == -1 {
			return errNotEnoughMemory
		}
	}
	return nil
}

func validateStorageSize(cluster *DatabaseCluster) error {
	_, err := cluster.Spec.Engine.Storage.Size.AsDatabaseClusterSpecEngineStorageSize0()
	if err == nil {
		return errInt64NotSupported
	}
	sizeStr, err := cluster.Spec.Engine.Storage.Size.AsDatabaseClusterSpecEngineStorageSize1()

	if err == nil {
		size, err := resource.ParseQuantity(sizeStr)
		if err != nil {
			return err
		}
		if size.Cmp(minStorageQuantity) == -1 {
			return errNotEnoughDiskSize
		}
	}
	return nil
}

func validateDatabaseClusterOnUpdate(dbc *DatabaseCluster, oldDB *everestv1alpha1.DatabaseCluster) error {
	if dbc.Spec.Engine.Version != nil {
		// XXX: Right now we do not support upgrading of versions
		// because it varies across different engines. Also, we should
		// prohibit downgrades. Hence, if versions are not equal we just return an error
		if oldDB.Spec.Engine.Version != *dbc.Spec.Engine.Version {
			return errors.New("changing version is not allowed")
		}
	}
	if *dbc.Spec.Engine.Replicas < oldDB.Spec.Engine.Replicas && *dbc.Spec.Engine.Replicas == 1 {
		// XXX: We can scale down multiple node clusters to a single node but we need to set
		// `allowUnsafeConfigurations` to `true`. Having this configuration is not recommended
		// and makes a database cluster unsafe. Once allowUnsafeConfigurations set to true you
		// can't set it to false for all operators and psmdb operator does not support it.
		//
		// Once it is supported by all operators we can revert this.
		return fmt.Errorf("cannot scale down %d node cluster to 1. The operation is not supported", oldDB.Spec.Engine.Replicas)
	}
	return nil
}
