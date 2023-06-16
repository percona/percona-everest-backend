package api

import (
	"net/url"
	"regexp"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// ErrNameNotRFC1123Compatible when the given fieldName doesn't contain RFC 1123 compatible string.
func ErrNameNotRFC1123Compatible(fieldName string) error {
	return errors.Errorf("'%s' is not RFC 1123 compatible", fieldName)
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

	if ok := validateURL(params.Url); !ok {
		err := ErrInvalidURL("url")
		return nil, err
	}

	return &params, nil
}
