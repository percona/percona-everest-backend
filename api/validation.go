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
	"net/url"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/percona/percona-everest-backend/cmd/config"
	"github.com/percona/percona-everest-backend/model"
)

// ErrNameNotRFC1123Compatible when the given fieldName doesn't contain RFC 1123 compatible string.
func ErrNameNotRFC1123Compatible(fieldName string) error {
	return errors.Errorf("'%s' is not RFC 1123 compatible", fieldName)
}

// ErrCreateStorageNotSupported appears when trying to create a storage of a type that is not supported.
func ErrCreateStorageNotSupported(storageType string) error {
	return errors.Errorf("Creating storage is not implemented for '%s'", storageType)
}

// ErrUpdateStorageNotSupported appears when trying to update a storage of a type that is not supported.
func ErrUpdateStorageNotSupported(storageType string) error {
	return errors.Errorf("Updating storage is not implemented for '%s'", storageType)
}

// ErrInvalidURL when the given fieldName contains invalid URL.
func ErrInvalidURL(fieldName string) error {
	return errors.Errorf("'%s' is an invalid URL", fieldName)
}

// validates names to be RFC-1123 compatible  https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
func validateRFC1123(s string) bool {
	rfc1123Regex := "^[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$"
	re := regexp.MustCompile(rfc1123Regex)
	return re.MatchString(s)
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

	if *endpoint == "" {
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

	if params.Name != nil {
		if ok := validateRFC1123(*params.Name); !ok {
			err := ErrNameNotRFC1123Compatible("name")
			return nil, err
		}
	}

	if params.Url != nil {
		if ok := validateURL(*params.Url); !ok {
			err := ErrInvalidURL("url")
			return nil, err
		}
	}

	return &params, nil
}

func validateCreateBackupStorageRequest(ctx echo.Context) (*CreateBackupStorageParams, error) {
	var params CreateBackupStorageParams
	if err := ctx.Bind(&params); err != nil {
		return nil, err
	}

	if ok := validateRFC1123(params.Name); !ok {
		err := ErrNameNotRFC1123Compatible("name")
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
		return nil, err
	}

	return &params, nil
}

func validateCreatePMMInstanceRequest(ctx echo.Context) (*CreatePMMInstanceJSONRequestBody, error) {
	var params CreatePMMInstanceJSONRequestBody
	if err := ctx.Bind(&params); err != nil {
		return nil, err
	}

	if ok := validateURL(params.Url); !ok {
		err := ErrInvalidURL("url")
		return nil, err
	}

	return &params, nil
}

func validateUpdatePMMInstanceRequest(ctx echo.Context) (*UpdatePMMInstanceJSONRequestBody, error) {
	var params UpdatePMMInstanceJSONRequestBody
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
