package errcodex

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var (
	defaultErrSuccess       = newDefaultErrCode(http.StatusOK)
	defaultErrBadRequest    = newDefaultErrCode(http.StatusBadRequest)
	defaultErrInternalError = newDefaultErrCode(http.StatusInternalServerError)
)

type DefaultErrCode struct {
	Code   string
	Status int
	Errno  int
}

func newDefaultErrCode(code int) *DefaultErrCode {
	return &DefaultErrCode{
		trimHttpStatusText(code),
		code,
		code,
	}
}

func (de *DefaultErrCode) GetCode() string {
	return de.Code
}
func (de *DefaultErrCode) GetHttpStatus() int {
	return de.Status
}
func (de *DefaultErrCode) GetErrno() int {
	return de.Errno
}
func (de *DefaultErrCode) Unwrap() error {
	if de.Status != http.StatusOK {
		return fmt.Errorf("Status:%d, Code:%s, Errno:%d", de.Status, de.Code, de.Errno)
	} else {
		return nil
	}
}

type DefaultErrCodeDic struct {
}

func (de *DefaultErrCodeDic) GetBadRequest() ErrorCodeIf {
	return defaultErrBadRequest
}

func (de *DefaultErrCodeDic) GetInternalError() ErrorCodeIf {
	return defaultErrInternalError
}

func (de *DefaultErrCodeDic) GetSuccess() ErrorCodeIf {
	return defaultErrSuccess
}

func (de *DefaultErrCodeDic) NewRequestId() string {
	return uuid.New().String()
}

func (de *DefaultErrCodeDic) ToCode(errno int) string {
	return trimHttpStatusText(errno)
}

func (de *DefaultErrCodeDic) ToHttpStatus(errno int) int {
	httpStatusText := http.StatusText(errno)
	if httpStatusText == "" {
		return 0
	} else {
		return errno
	}
}

func trimHttpStatusText(status int) string {
	trimmedSpace := strings.Replace(http.StatusText(status), " ", "", -1)
	trimmedSpace = strings.Replace(trimmedSpace, "-", "", -1)
	return trimmedSpace
}

type ErrorCodeIf interface {
	GetHttpStatus() int
	GetCode() string
	GetErrno() int
	Unwrap() error
}

type ErrorCodeDictionaryIf interface {
	GetSuccess() ErrorCodeIf
	GetBadRequest() ErrorCodeIf
	GetInternalError() ErrorCodeIf

	ToCode(int) string
	ToHttpStatus(int) int
	NewRequestId() string
}
