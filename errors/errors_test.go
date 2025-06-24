package errors_test

import (
	"fmt"
	"testing"

	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/httpx"
)

func callTest() error {
	return httpx.Wrap(httpx.SuccessResp(nil))
}

func TestStackPrint(t *testing.T) {

	fmt.Printf("no wrap: %+v\n", callTest())
	fmt.Printf("with wrap: %+v\n", errors.Wrap(callTest()))
}

func TestWrap(t *testing.T) {
	jr := httpx.SuccessResp("")

	err := errors.Wrapf(jr, "test")
	fmt.Printf("no wrap: %+v\n", err)
	fmt.Printf("no wrap: %v\n", err)
}
