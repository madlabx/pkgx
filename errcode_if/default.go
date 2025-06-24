package errcode_if

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/madlabx/pkgx/utils"
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
		utils.TrimHttpStatusText(code),
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
	return utils.TrimHttpStatusText(errno)
}

func (de *DefaultErrCodeDic) ToHttpStatus(errno int) int {
	httpStatusText := http.StatusText(errno)
	if httpStatusText == "" {
		return 0
	} else {
		return errno
	}
}
