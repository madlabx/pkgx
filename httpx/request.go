package httpx

import (
	"reflect"

	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/errors"
)

func ValidateMust(input interface{}, keys ...string) error {
	v := reflect.ValueOf(input)
	t := reflect.TypeOf(input)

	for _, key := range keys {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i)
			if field.Name == key && value.IsZero() {
				return errors.New("Need " + key)
			}
		}
	}

	return nil
}

func QueryMustParam(c echo.Context, key string) (string, error) {
	var err error
	value := c.QueryParam(key)
	if len(value) == 0 {
		err = errors.New("Missing " + key)
	}

	return value, err
}

func QueryOptionalParam(c echo.Context, key string) (string, bool) {
	value := c.QueryParam(key)
	return value, len(value) != 0
}
