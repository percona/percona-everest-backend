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

// Package api ...
package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/AlekSi/pointer"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/percona/percona-everest-backend/cmd/config"
	"github.com/percona/percona-everest-backend/model"
)

const (
	pxcDeploymentName   = "percona-xtradb-cluster-operator"
	psmdbDeploymentName = "percona-server-mongodb-operator"
	pgDeploymentName    = "percona-postgresql-operator"
	engineTypePXC       = "pxc"
	engineTypePSMDB     = "psmdb"
	engineTypePG        = "postgresql"
)

var (
	errDBCEmptyMetadata   = errors.New("DatabaseCluster's Metadata should not be empty")
	errDBCNameEmpty       = errors.New("DatabaseCluster's metadata.name should not be empty")
	errDBCNameWrongFormat = errors.New("DatabaseCluster's metadata.name should be a string")
	operatorEngine        = map[string]string{
		engineTypePXC:   pxcDeploymentName,
		engineTypePSMDB: psmdbDeploymentName,
		engineTypePG:    pgDeploymentName,
	}
	minStorageQuantity = resource.MustParse("1G")
	minCPUQuantity     = resource.MustParse("600m")
	minMemQuantity     = resource.MustParse("512M")
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
	return fmt.Errorf("Creating storage is not implemented for '%s'", storageType)
}

// ErrUpdateStorageNotSupported appears when trying to update a storage of a type that is not supported.
func ErrUpdateStorageNotSupported(storageType string) error {
	return fmt.Errorf("Updating storage is not implemented for '%s'", storageType)
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

func validateStorageAccessByCreate(params CreateBackupStorageParams) error {
	switch params.Type { //nolint:exhaustive
	case CreateBackupStorageParamsTypeS3:
		return s3Access(params.Url, params.AccessKey, params.SecretKey, params.BucketName, params.Region)
	default:
		return ErrCreateStorageNotSupported(string(params.Type))
	}
}

func validateStorageAccessByUpdate(oldData *storageData, params UpdateBackupStorageParams) error {
	endpoint := &oldData.storage.URL
	if params.Url != nil {
		endpoint = params.Url
	}

	accessKey := oldData.accessKey
	if params.AccessKey != nil {
		accessKey = *params.AccessKey
	}

	secretKey := oldData.secretKey
	if params.SecretKey != nil {
		secretKey = *params.SecretKey
	}

	bucketName := oldData.storage.BucketName
	if params.BucketName != nil {
		bucketName = *params.BucketName
	}

	region := oldData.storage.Region
	if params.Region != nil {
		region = *params.Region
	}

	switch oldData.storage.Type {
	case string(BackupStorageTypeS3):
		return s3Access(endpoint, accessKey, secretKey, bucketName, region)
	default:
		return ErrUpdateStorageNotSupported(oldData.storage.Type)
	}
}

type storageData struct {
	accessKey string
	secretKey string
	storage   model.BackupStorage
}

func s3Access(endpoint *string, accessKey, secretKey, bucketName, region string) error {
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
		return err
	}

	// Create a new S3 client with the session
	svc := s3.New(sess)

	_, err = svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return err
	}

	return nil
}

func validateUpdateBackupStorageRequest(ctx echo.Context) (*UpdateBackupStorageParams, error) {
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

	// check data access
	if err := validateStorageAccessByCreate(params); err != nil {
		l.Error(err)
		return nil, errors.New("Could not connect to the backup storage, please check the new credentials are correct")
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

func (e *EverestServer) validateDBClusterAccess(ctx echo.Context, kubernetesID, dbClusterName string) error {
	_, kubeClient, code, err := e.initKubeClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return ctx.JSON(code, Error{Message: pointer.ToString(err.Error())})
	}

	_, err = kubeClient.GetDatabaseCluster(ctx.Request().Context(), dbClusterName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return ctx.JSON(http.StatusBadRequest, Error{Message: pointer.ToString(fmt.Sprintf("DatabaseCluster '%s' is not found", dbClusterName))})
		}
		e.l.Error(err)
		return ctx.JSON(http.StatusInternalServerError, Error{Message: pointer.ToString(err.Error())})
	}

	return nil
}

func (e *EverestServer) validateDatabaseClusterCR(ctx echo.Context, kubernetesID string, databaseCluster *DatabaseCluster) error {
	if err := validateCreateDatabaseClusterRequest(*databaseCluster); err != nil {
		return err
	}

	_, kubeClient, _, err := e.initKubeClient(ctx.Request().Context(), kubernetesID)
	if err != nil {
		return err
	}
	engineName, ok := operatorEngine[databaseCluster.Spec.Engine.Type]
	if !ok {
		return errors.New("Unsupported database engine")
	}
	engine, err := kubeClient.GetDatabaseEngine(ctx.Request().Context(), engineName)
	if err != nil {
		return err
	}
	if databaseCluster.Spec.Engine.Version != nil {
		if len(engine.Spec.AllowedVersions) != 0 && !containsVersion(*databaseCluster.Spec.Engine.Version, engine.Spec.AllowedVersions) {
			return fmt.Errorf("Using %s version for %s is not allowed", databaseCluster.Spec.Engine.Version, databaseCluster.Spec.Engine.Type)
		}
		if _, ok := engine.Status.AvailableVersions.Engine[*databaseCluster.Spec.Engine.Version]; !ok {
			return fmt.Errorf("%s is not in available versions list", *databaseCluster.Spec.Engine.Version)
		}
	}
	if databaseCluster.Spec.Proxy.Type != nil {
		if err := validateProxy(databaseCluster.Spec.Engine.Type, string(*databaseCluster.Spec.Proxy.Type)); err != nil {
			return err
		}
	}
	if err := validateBackupSpec(databaseCluster); err != nil {
		return err
	}
	if err := validateResourceLimits(databaseCluster); err != nil {
		return err
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
	if engineType == engineTypePXC && (proxyType != "proxysql" || proxyType != "haproxy") {
		return errors.New("You can use only either HAProxy or Proxy SQL for PXC clusters")
	}

	if engineType == engineTypePG && proxyType != "pgbouncer" {
		return errors.New("You can use only PGBouncer as a proxy type for Postgres clusters")
	}
	if engineType == engineTypePSMDB && proxyType != "mongos" {
		return errors.New("You can use only Mongos as a proxy type for MongoDB clusters")
	}
	return nil
}

func validateBackupSpec(cluster *DatabaseCluster) error {
	return nil
}
func validateResourceLimits(cluster *DatabaseCluster) error {
	if cluster.Spec.Engine.Resources == nil {
		return errors.New("Please specify resource limits for the cluster")
	}
	if cluster.Spec.Engine.Resources.Cpu == nil {
		return errors.New("CPU limits should be above 600m and cannot be empty")
	}
	if cluster.Spec.Engine.Resources.Memory == nil {
		return errors.New("Memory limits should be above 512M and cannot be empty")
	}
	cpuStr, err := cluster.Spec.Engine.Resources.Cpu.AsDatabaseClusterSpecEngineResourcesCpu1()
	if err == nil {
		cpu, err := resource.ParseQuantity(cpuStr)
		if err != nil {
			return err
		}
		if cpu.Cmp(minCPUQuantity) == -1 {
			return errors.New("CPU limits should be above 600m")
		}
	}
	_, err = cluster.Spec.Engine.Resources.Cpu.AsDatabaseClusterSpecEngineResourcesCpu0()
	if err == nil {
		return errors.New("Specifying resources using int64 data type is not supported. Please use string format for that")
	}
	_, err = cluster.Spec.Engine.Resources.Memory.AsDatabaseClusterSpecEngineResourcesMemory0()
	if err == nil {
		return errors.New("Specifying resources using int64 data type is not supported. Please use string format for that")
	}
	memStr, err := cluster.Spec.Engine.Resources.Memory.AsDatabaseClusterSpecEngineResourcesMemory1()
	if err == nil {
		mem, err := resource.ParseQuantity(memStr)
		if err != nil {
			return err
		}
		if mem.Cmp(minMemQuantity) == -1 {
			return errors.New("Memory limits should be above 512M")
		}
	}
	_, err = cluster.Spec.Engine.Storage.Size.AsDatabaseClusterSpecEngineStorageSize0()
	if err == nil {
		return errors.New("Specifying resources using int64 data type is not supported. Please use string format for that")
	}
	sizeStr, err := cluster.Spec.Engine.Storage.Size.AsDatabaseClusterSpecEngineStorageSize1()
	if err == nil {

		size, err := resource.ParseQuantity(sizeStr)
		if err != nil {
			return err
		}
		if size.Cmp(minStorageQuantity) == -1 {
			return errors.New("Storage size should be above 1G")
		}
	}
	return nil
}
