package errcode

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/madlabx/pkgx/errcode_if"
	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/httpx"
	"github.com/madlabx/pkgx/log"
	"github.com/madlabx/pkgx/utils"
	"gorm.io/gorm"
)

var _ errcode_if.ErrorCodeIf = &ErrorCode{}

func init() {
	httpx.RegisterErrCodeDictionary(&ErrorCode{})
}

type ErrorCode struct {
	httpx.JsonResponse
}

// ConvertJsonResponse implement HttpxJsonResponseWrapper
func (ec *ErrorCode) ToHttpXJsonResponse() *httpx.JsonResponse {
	return &ec.JsonResponse
}

func (ec *ErrorCode) GetHttpStatus() int {
	return ec.Status
}

func (ec *ErrorCode) GetCode() string {
	return ec.Code
}

func (ec *ErrorCode) GetErrno() int {
	return ec.Errno
}

func (ec *ErrorCode) GetBadRequest() errcode_if.ErrorCodeIf {
	return ErrBadRequest()
}

func (ec *ErrorCode) GetInternalError() errcode_if.ErrorCodeIf {
	return ErrInternalServerError()
}

func (ec *ErrorCode) GetSuccess() errcode_if.ErrorCodeIf {
	return OK
}

func (ec *ErrorCode) NewRequestId() string {
	return uuid.New().String()
}

func (ec *ErrorCode) ToCode(errno int) string {
	return utils.TrimHttpStatusText(errno)
}

func (ec *ErrorCode) ToHttpStatus(errno int) int {
	httpStatusText := http.StatusText(errno)
	if httpStatusText == "" {
		return 0
	} else {
		return errno
	}
}

// HttpCode required by interface httpx.JsonResponseError
func (ec *ErrorCode) HttpCode() int {
	//TODO support customized code
	return ec.Status
}

//func (ec *ErrorCode) Is(target error) bool {
//	te, ok := target.(*ErrorCode)
//	return ok && ec.Code == te.Code
//}

// HttpCode required by interface httpx.JsonResponseError
//
//	func (ec *ErrorCode) ToJsonResponseWithStack(depth int) *httpx.JsonResponse {
//		//TODO support customized code
//		return errors.WrapWithRelativeStackDepth()
//	}

var (
	once        sync.Once
	errCodeDict map[string]*ErrorCode
)

func Len() int {
	return len(errCodeDict)
}

func DumpErrorCodes() string {
	output := fmt.Sprintf("Total:%d\n", len(errCodeDict))
	return utils.ToPrettyString(errCodeDict) + output
}

func New(errno int, opts ...string) func(errs ...error) *ErrorCode {
	once.Do(func() {
		errCodeDict = make(map[string]*ErrorCode)
	})

	errCode := newErrCode(errno, opts...)
	errCodeDict[errCode.Code] = errCode

	return func(errs ...error) *ErrorCode {
		//return clone
		ecopy := &ErrorCode{JsonResponse: *errCode.Copy()}
		if errs == nil {
			_ = ecopy.WithStack(2)
		} else {
			if len(errs) > 1 {
				log.Panicf("cannot accept more than one error")
			}
			//TODO only accept 1 error
			_ = ecopy.WithError(errs[0], 2)
		}
		return ecopy
	}
}

func newErrCode(errno int, opts ...string) *ErrorCode {
	httpStatusText := http.StatusText(errno)

	var httpStatus int
	if errno < constClientErrorBaseIndex && httpStatusText != "" {
		httpStatus = errno
	} else if errno >= constClientErrorBaseIndex && errno < constInternalErrorBaseIndex {
		httpStatus = http.StatusBadRequest
	} else if errno >= constInternalErrorBaseIndex && errno < constInvalidErrorBaseIndex {
		httpStatus = http.StatusInternalServerError
	} else {
		return InvalidErrorCode
	}

	var code string
	if len(opts) > 0 {
		code = opts[0]
	} else {
		trimmedSpace := strings.Replace(httpStatusText, " ", "", -1)
		code = strings.Replace(trimmedSpace, "-", "", -1)
	}

	err := &ErrorCode{httpx.JsonResponse{
		Status: httpStatus,
		Errno:  errno,
		Code:   code,
	}}

	return err
}

const (
	constClientErrorBaseIndex     = 4000
	constInternalErrorBaseIndex   = 5000
	constDependencyErrorBaseIndex = 6000
	constInvalidErrorBaseIndex    = 9999
)

var (
	OK = newErrCode(http.StatusOK)

	InvalidErrorCode = &ErrorCode{httpx.JsonResponse{
		Status: 0,
		Errno:  constInvalidErrorBaseIndex,
		Code:   "FatalErrorInvalidErrorCode",
	}}
)

var (
	ErrSuccess             = New(http.StatusOK)
	ErrInsufficientStorage = New(http.StatusInsufficientStorage)
	ErrInternalServerError = New(http.StatusInternalServerError)
	ErrTimeout             = New(http.StatusGatewayTimeout)
	ErrForbidden           = New(http.StatusForbidden)
	ErrNotFound            = New(http.StatusNotFound)
	ErrConflict            = New(http.StatusConflict)
	ErrUnauthorized        = New(http.StatusUnauthorized)
	ErrPreconditionFailed  = New(http.StatusPreconditionFailed)
	ErrTooManyRequests     = New(http.StatusTooManyRequests)
	ErrNotImplemented      = New(http.StatusNotImplemented)
	ErrBadRequest          = New(http.StatusBadRequest)

	ErrIdempotent = New(http.StatusAccepted, "AlreadyAccepted")

	ErrLengthRequired               = New(http.StatusLengthRequired)
	ErrRequestURITooLong            = New(http.StatusRequestURITooLong)
	ErrUnsupportedMediaType         = New(http.StatusUnsupportedMediaType)
	ErrRequestedRangeNotSatisfiable = New(http.StatusRequestedRangeNotSatisfiable)
	ErrExpectationFailed            = New(http.StatusExpectationFailed)
	ErrTeapot                       = New(http.StatusTeapot)
	ErrMisdirectedRequest           = New(http.StatusMisdirectedRequest)
	ErrUnprocessableEntity          = New(http.StatusUnprocessableEntity)
	ErrLocked                       = New(http.StatusLocked)
	ErrFailedDependency             = New(http.StatusFailedDependency)
	ErrUpgradeRequired              = New(http.StatusUpgradeRequired)
	ErrPreconditionRequired         = New(http.StatusPreconditionRequired)
	ErrRequestHeaderFieldsTooLarge  = New(http.StatusRequestHeaderFieldsTooLarge)
	ErrUnavailableForLegalReasons   = New(http.StatusUnavailableForLegalReasons)
	ErrMultiStatus                  = New(http.StatusMultiStatus)
	ErrAlreadyReported              = New(http.StatusAlreadyReported)
	ErrIMUsed                       = New(http.StatusIMUsed)

	ErrMultipleChoices   = New(http.StatusMultipleChoices)
	ErrMovedPermanently  = New(http.StatusMovedPermanently)
	ErrFound             = New(http.StatusFound)
	ErrSeeOther          = New(http.StatusSeeOther)
	ErrNotModified       = New(http.StatusNotModified)
	ErrUseProxy          = New(http.StatusUseProxy)
	ErrTemporaryRedirect = New(http.StatusTemporaryRedirect)
	ErrPermanentRedirect = New(http.StatusPermanentRedirect)

	ErrPaymentRequired       = New(http.StatusPaymentRequired)
	ErrMethodNotAllowed      = New(http.StatusMethodNotAllowed)
	ErrNotAcceptable         = New(http.StatusNotAcceptable)
	ErrProxyAuthRequired     = New(http.StatusProxyAuthRequired)
	ErrRequestTimeout        = New(http.StatusRequestTimeout)
	ErrGone                  = New(http.StatusGone)
	ErrRequestEntityTooLarge = New(http.StatusRequestEntityTooLarge)
	ErrTooEarly              = New(http.StatusTooEarly)

	ErrBadGateway                    = New(http.StatusBadGateway)
	ErrServiceUnavailable            = New(http.StatusServiceUnavailable)
	ErrGatewayTimeout                = New(http.StatusGatewayTimeout)
	ErrHTTPVersionNotSupported       = New(http.StatusHTTPVersionNotSupported)
	ErrVariantAlsoNegotiates         = New(http.StatusVariantAlsoNegotiates)
	ErrLoopDetected                  = New(http.StatusLoopDetected)
	ErrNotExtended                   = New(http.StatusNotExtended)
	ErrNetworkAuthenticationRequired = New(http.StatusNetworkAuthenticationRequired)
)

var (
	ErrIdempotentError   = New(http.StatusAccepted, "AlreadyAccepted")
	ErrWrongSign         = New(http.StatusBadRequest, "WrongSign")
	ErrExpiredRequest    = New(http.StatusBadRequest, "ExpiredRequest")
	ErrObjectExist       = New(http.StatusBadRequest, "ObjectExist")
	ErrObjectNotExist    = New(http.StatusBadRequest, "ObjectNotExist")
	ErrUserExist         = New(http.StatusBadRequest, "UserExist")
	ErrUserNotExist      = New(http.StatusBadRequest, "UserNotExist")
	ErrSessionNotExist   = New(http.StatusBadRequest, "SessionNotExist")
	ErrSessionExist      = New(http.StatusBadRequest, "SessionExist")
	ErrUnsupportedDevice = New(http.StatusBadRequest, "UnsupportedDevice")
	ErrInvalidDeviceId   = New(http.StatusBadRequest, "InvalidDeviceId")

	ErrRTokenDisabled  = New(http.StatusUnauthorized, "RTokenDisabled")
	ErrInvalidJwt      = New(http.StatusUnauthorized, "InvalidJwt")
	ErrUnmatchedJwtKey = New(http.StatusUnauthorized, "UnmatchedJwtKey")
	ErrInvalidPassword = New(http.StatusUnauthorized, "InvalidPassword")
	ErrInvalidNonce    = New(http.StatusUnauthorized, "InvalidNonce")
	ErrExpiredNonce    = New(http.StatusUnauthorized, "ExpiredNonce")
	ErrInvalidSign     = New(http.StatusUnauthorized, "InvalidSign")
	ErrExpiredToken    = New(http.StatusUnauthorized, "ExpiredToken")
	ErrInvalidToken    = New(http.StatusUnauthorized, "InvalidToken")
	ErrInvalidIssuer   = New(http.StatusUnauthorized, "InvalidIssuer")

	ErrDeviceCommunicateErr = New(http.StatusInternalServerError, "DeviceCommunicateErr")
	ErrNotReady             = New(http.StatusInternalServerError, "NotReady")
	ErrInvalidDataType      = New(http.StatusInternalServerError, "InvalidDataType")

	ErrDeviceOffline  = New(http.StatusBadRequest, "DeviceOffline")
	ErrInUpgrading    = New(http.StatusBadRequest, "InUpgrading")
	ErrUpgradeTimeout = New(http.StatusInternalServerError, "UpgradeTimeout")

	ErrOnethingTooBusy = New(http.StatusTooManyRequests, "OnethingTooBusy")
	ErrObjectBusy      = New(http.StatusTooManyRequests, "ObjectBusy")

	ErrInvalidVrfCode = New(http.StatusBadRequest, "InvalidSmsCode")
	ErrExpiredVrfCode = New(http.StatusBadRequest, "ExpiredSmsCode")

	ErrWrongWxNode = New(http.StatusBadRequest, "WrongWxNode")

	ErrOmDevicePppoe                  = New(http.StatusInternalServerError, "OmDevicePppoe")
	ErrOmDeviceOffline                = New(http.StatusInternalServerError, "OmDeviceOffline")
	ErrOmDeviceEmptyResponse          = New(http.StatusInternalServerError, "OmDeviceEmptyResponse")
	ErrOmResponseError                = New(http.StatusInternalServerError, "OmResponseError")
	ErrOnethingAlreadyBound           = New(http.StatusInternalServerError, "OnethingAlreadyBound")
	ErrOnethingBadRequest             = New(http.StatusInternalServerError, "OnethingBadRequest")
	ErrOnethingNoDevice               = New(http.StatusInternalServerError, "OnethingNoDevice")
	ErrOnethingNotBound               = New(http.StatusInternalServerError, "OnethingNotBound")
	ErrOnethingWrongActCodeOrUnAct    = New(http.StatusInternalServerError, "OnethingWrongActCodeOrUnAct")
	ErrOnethingNoBindingDevice        = New(http.StatusInternalServerError, "OnethingNoBindingDevice")
	ErrOnethingError                  = New(http.StatusInternalServerError, "OnethingError")
	ErrOnethingParseError             = New(http.StatusInternalServerError, "OnethingParseError")
	ErrOnethingBlockedForViolation    = New(http.StatusInternalServerError, "OnethingBlockedForViolation")
	ErrOnethingAccountExist           = New(http.StatusInternalServerError, "OnethingAccountExist")
	ErrOnethingBlockedForUnregistered = New(http.StatusInternalServerError, "OnethingBlockedForUnregistered")
)

func ErrHttpStatus(status int) *ErrorCode {
	return New(status)()
}

func ErrStrResp(status int, b any, format string, a ...any) error {
	err := New(status)()
	return err.WithErrorf(format, a...)
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound) ||
		errors.Is(err, ErrObjectNotExist()) ||
		errors.Is(err, ErrUserNotExist())
}
