package errcode

import (
	"fmt"
	"net/http"
	"runtime"
	"testing"

	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/httpx"
	"github.com/stretchr/testify/require"
)

func TestWrapItself(t *testing.T) {
	err := ErrBadRequest(ErrBadRequest())
	require.NotNil(t, err)
}

func TestIsItself(t *testing.T) {
	err := errors.Wrap(ErrObjectNotExist())

	ok := errors.Is(err, ErrNotFound())
	require.False(t, ok)
}

func TestWrap(t *testing.T) {
	err := errors.New("test")
	err1 := ErrBadRequest(err)
	require.Equal(t, err.Error(), err1.Error())
	require.Equal(t, ErrBadRequest().Code, err1.Code)

	jrErr := httpx.Wrap(err1)
	require.Equal(t, err.Error(), jrErr.Error())
	require.Equal(t, ErrBadRequest().Code, jrErr.Code)
}

func TestWithErrorf2(t *testing.T) {
	err := errors.New("test")
	err1 := ErrBadRequest().WithError(err)
	require.Equal(t, err.Error(), err1.Error())
	require.Equal(t, ErrBadRequest().Code, err1.Code)

	jrErr := httpx.Wrap(err1)
	require.Equal(t, err.Error(), jrErr.Error())
	require.Equal(t, ErrBadRequest().Code, jrErr.Code)
}

func TestWithErrorf(t *testing.T) {
	e := ErrObjectExist().WithErrorf("dir existing").WithMsgf("dir existing1")
	require.Equal(t, "dir existing", e.Error())
	require.Equal(t, "dir existing1", e.Message)

	e1 := ErrObjectExist()
	//should not impact e
	require.NotEqual(t, e, e1)
	require.Equal(t, "Code:ObjectExist, Errno:400", e1.Error())
	require.Equal(t, "", e1.Message)
	require.Equal(t, "Code:ObjectExist, Errno:400", e1.Cause())

	e2 := ErrObjectExist().WithErrorf("dir existing")
	require.Equal(t, e.Error(), e2.Error())
	require.NotEqual(t, e1, e2)
	require.NotEqual(t, e1.Error(), e2.Error())
	require.NotEqual(t, e.Message, e2.Message)

	e3 := e2.WithErrorf("dir existing")
	require.Equal(t, "dir existing", e2.Error())
	require.Equal(t, "dir existing", e3.Error())
	require.Equal(t, "", e3.Message)
	require.Equal(t, e3, e2)

	e4 := ErrObjectExist().WithErrorf("dir existing1")
	require.NotEqual(t, e3.Error(), e4.Error())
}

func currentStackTrackDeep() int {
	const depth = 32
	var pcs [depth]uintptr
	return runtime.Callers(2, pcs[:])
}

func newError(aaa string) error {
	return errors.New(aaa)
}

func TestWithMessagef(t *testing.T) {
	e := ErrObjectExist(errors.New("dir existing"))
	e1 := ErrObjectExist()
	//should not impact e
	require.NotEqual(t, e, e1)

	e2 := ErrObjectExist(errors.New("dir existing"))
	require.Equal(t, e.Error(), e2.Error())
	require.NotEqual(t, e1.Error(), e2.Error())

	e3 := e2.WithMsgf("dir existing")
	require.Equal(t, e3.Error(), e2.Error())
	require.Equal(t, &e2.JsonResponse, e3)
	require.Equal(t, e2.IsOK(), e3.IsOK())

	e4 := ErrSuccess()
	msgString := "i am test"
	var e5 *httpx.JsonResponse
	errors.As(ErrSuccess().WithMsgf(msgString), &e5)
	require.NotEqual(t, e4, e5)
	require.True(t, e5.IsOK())

	require.Equal(t, msgString, e5.Message)
	require.NotEqual(t, msgString, e4.Error())

	require.Equal(t, "", e5.Error())
	require.NotEqual(t, e5.JsonString(), e4.JsonString())

}

func TestRawWrap(t *testing.T) {
	n := currentStackTrackDeep()

	rawErr := newError("dir existing")
	e := ErrObjectExist(rawErr)
	require.Equal(t, rawErr, e.Unwrap())

	e1 := ErrObjectExist()
	//should not impact e
	require.NotEqual(t, e, e1)

	e2 := ErrObjectExist(rawErr)
	require.Equal(t, e.Error(), e2.Error())
	require.NotEqual(t, e1.Error(), e2.Error())
	require.Equal(t, n+1, len(e2.Frames()))

	e3 := e2.JsonResponse.WithErrorf("dir existing")
	require.Equal(t, e3.Error(), e2.Error())

	e4 := ErrObjectExist(errors.New("dir existing1"))
	require.NotEqual(t, e3.Error(), e4.Error())

	require.Equal(t, n+1, len(e.Frames()))
	require.Equal(t, n, len(e1.Frames()))
	require.Equal(t, n, len(e2.Frames()))
	require.Equal(t, n, len(e3.Frames()))
	require.Equal(t, n, len(e4.Frames()))
}

func TestStack(t *testing.T) {
	n := currentStackTrackDeep()

	err := fmt.Errorf("testerror")
	ec := ErrHttpStatus(http.StatusTooManyRequests).WithError(err)

	require.Equal(t, n, len(ec.Frames()))
	require.Equal(t, n, len(errors.Frames(ec.Unwrap())))

	ec2 := ErrHttpStatus(http.StatusTooManyRequests).WithErrorf("err from WithErrorf, %v", err)

	require.Equal(t, n, len(ec2.Frames()))

	ec1 := ErrBadRequest().WithErrorf("err from WithErrorf")

	require.Equal(t, n, len(ec1.Frames()))

	wrapErr := errors.Wrapf(ec1, "newError from errors.Wrapf")

	require.Equal(t, n, len(errors.Frames(wrapErr)))
}

func TestStack2(t *testing.T) {
	n := currentStackTrackDeep()

	rawErr := newError("dir existing")
	e := ErrObjectExist(rawErr)
	require.Equal(t, rawErr, e.Unwrap())

	e1 := ErrObjectExist()
	//should not impact e
	require.NotEqual(t, e, e1)

	e2 := ErrObjectExist(rawErr)
	require.Equal(t, e.Error(), e2.Error())
	require.NotEqual(t, e1.Error(), e2.Error())

	require.Equal(t, n+1, len(e2.Frames()))
	e3 := e2.JsonResponse.WithErrorf("dir existing")
	require.Equal(t, e3.Error(), e2.Error())

	e4 := ErrObjectExist(errors.New("dir existing1"))
	require.NotEqual(t, e3.Error(), e4.Error())

	require.Equal(t, n+1, len(e.Frames()))
	require.Equal(t, n, len(e1.Frames()))
	require.Equal(t, n, len(e2.Frames()))
	require.Equal(t, n, len(e3.Frames()))
	require.Equal(t, n, len(e4.Frames()))
}

func TestWithWrap(t *testing.T) {
	innerErr := newError("innerError")
	e := ErrObjectExist(innerErr)

	msgString := "thisIsMessageString"
	e1 := errors.Wrapf(e, msgString)
	require.Equal(t, &e.JsonResponse, e1)
	require.Equal(t, msgString, e.Message)
	require.Equal(t, msgString, e1.(*httpx.JsonResponse).Message)

	require.Equal(t, e1, httpx.Wrap(e1))
	require.Equal(t, "{\"Code\":\"ObjectExist\",\"Errno\":400,\"Message\":\"thisIsMessageString\"}", fmt.Sprintf("%v", e1))

	require.Equal(t, "{\"Code\":\"ObjectExist\",\"Errno\":400,\"Message\":\"Code:ObjectExist, Errno:400\"}", fmt.Sprintf("%v", ErrObjectExist()))
}
