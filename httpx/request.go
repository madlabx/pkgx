package httpx

import (
	"github.com/labstack/echo"
	"net/http"
	"reflect"
)

func ValidateMust(input interface{}, keys ...string) error {
	v := reflect.ValueOf(input)
	t := reflect.TypeOf(input)

	for _, key := range keys {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i)
			if field.Name == key && value.IsZero() {
				return MessageResp(http.StatusBadRequest, handleGetECodeBadRequest(), "Need "+key)
			}
		}
	}

	return nil
}

func QueryMustParam(c echo.Context, key string) (string, error) {
	var err error
	value := c.QueryParam(key)
	if len(value) == 0 {
		err = MessageResp(http.StatusBadRequest, handleGetECodeBadRequest(), "Missing "+key)
	}

	return value, err
}
