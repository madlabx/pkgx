package httpx

import (
	"fmt"
	"testing"

	"github.com/madlabx/pkgx/errors"
	"github.com/stretchr/testify/require"
)

func TestErrorIs(t *testing.T) {

	rawErr := errors.New("rawErrorString")

	newErr := Wrap(rawErr)

	isSameError := errors.Is(newErr, rawErr)
	require.True(t, isSameError)
	isSameError = errors.Is(rawErr, newErr)
	require.True(t, isSameError)
}

func TestErrorCopy(t *testing.T) {
	err := Wrap(errors.New("rawErrorString"))

	err2 := err.Copy()

	err.Code = "NEW"
	require.NotEqual(t, err.Code, err2.Code)
	fmt.Printf("%s, %s\n", err.Code, err2.Code)

	fmt.Printf("%p, %p\n", err, err2)
}
