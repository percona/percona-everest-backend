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

// ErrCreteStorageNotSupported appears when trying to create a storage of a type that is not supported.
func ErrCreteStorageNotSupported(storageType string) error {
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
	_, err := url.Parse(urlStr)
	return err == nil
}

func validateStorageAccessByCreate(params CreateBackupStorageParams) error {
	switch params.Type { //nolint:exhaustive
	case CreateBackupStorageParamsTypeS3:
		return s3Access(params.AccessKey, params.SecretKey, params.BucketName, params.Region)
	default:
		return ErrCreteStorageNotSupported(string(params.Type))
	}
}

func validateStorageAccessByUpdate(oldData *storageData, params UpdateBackupStorageParams) error {
	accessKey := oldData.accessKey
	if params.AccessKey != nil {
		accessKey = *params.AccessKey
	}

	secretKey := oldData.secretKey
	if params.SecretKey != nil {
		secretKey = *params.SecretKey
	}

	bucketName := oldData.s.BucketName
	if params.BucketName != nil {
		bucketName = *params.BucketName
	}

	region := oldData.s.Region
	if params.Region != nil {
		region = *params.Region
	}

	switch oldData.s.Type {
	case string(BackupStorageTypeS3):
		return s3Access(accessKey, secretKey, bucketName, region)
	default:
		return ErrUpdateStorageNotSupported(oldData.s.Type)
	}
}

type storageData struct {
	accessKey string
	secretKey string
	s         model.BackupStorage
}

func s3Access(accessKey, secretKey, bucketName, region string) error {
	if config.Debug {
		return nil
	}

	// Create a new session with the provided credentials
	sess, err := session.NewSession(&aws.Config{ //nolint:exhaustruct
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
