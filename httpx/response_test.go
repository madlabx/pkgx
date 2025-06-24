package httpx

import (
	"net/http"
	"runtime"
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
	require.False(t, isSameError)

	isSameError = errors.Is(newErr, newErr)
	require.True(t, isSameError)
}

func TestErrorCopy(t *testing.T) {
	err := Wrap(errors.New("rawErrorString"))

	err2 := err.Copy()

	err.Code = "NEW"
	require.NotEqual(t, err.Code, err2.Code)
	require.Equal(t, TrimHttpStatusText(http.StatusInternalServerError), err.Code)
}

func currentStackTrackDeep() int {
	const depth = 32
	var pcs [depth]uintptr
	return runtime.Callers(2, pcs[:])
}
func TestRawStruct(t *testing.T) {
	n := currentStackTrackDeep()
	msgString := "this is message"
	jr := JsonResponse{
		Result:  nil,
		Message: msgString,
		err:     nil,
		Code:    "OK",
		Errno:   200,
	}

	require.Equal(t, msgString, jr.Message)
	require.Equal(t, 0, len(jr.StackTrace()))
	require.Equal(t, 0, len(jr.Frames()))
	require.Equal(t, true, jr.IsOK())
	require.Equal(t, nil, jr.Unwrap())
	require.Equal(t, msgString, jr.CompleteMessage().Message)
	require.Equal(t, "", jr.Cause())

	errString := "this is error"
	_ = jr.WithErrorf(errString)
	require.Equal(t, "", jr.Message)
	require.Equal(t, n, len(jr.StackTrace()))
	require.Equal(t, n, len(jr.Frames()))
	require.Equal(t, false, jr.IsOK())
	require.Equal(t, errString, jr.Unwrap().Error())
	require.Equal(t, errString, jr.CompleteMessage().Message)
	require.Equal(t, errString, jr.Cause())

	_ = jr.WithMessagef("")
	require.Equal(t, "", jr.Message)
	require.Equal(t, n, len(jr.StackTrace()))
	require.Equal(t, n, len(jr.Frames()))
	require.Equal(t, false, jr.IsOK())
	require.Equal(t, errString, jr.Unwrap().Error())
	require.Equal(t, errString, jr.CompleteMessage().Message)
	require.Equal(t, errString, jr.Cause())
	require.Nil(t, jr.Result)

	_ = jr.WithMessagef("")
	resultString := "this is result"
	_ = jr.WithResult(resultString)
	require.Equal(t, "", jr.Message)
	require.Equal(t, n, len(jr.StackTrace()))
	require.Equal(t, n, len(jr.Frames()))
	require.Equal(t, false, jr.IsOK())
	require.Equal(t, errString, jr.Unwrap().Error())
	require.Equal(t, "", jr.CompleteMessage().Message)
	require.Equal(t, errString, jr.Cause())
	require.Equal(t, resultString, jr.Result)
}
