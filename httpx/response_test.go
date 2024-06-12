package httpx

import (
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
