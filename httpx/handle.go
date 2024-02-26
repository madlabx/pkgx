package httpx

import (
	"net/http"
	"strings"

	uuid "github.com/satori/go.uuid"
)

var (
	handleGetECodeSuccess       func() int
	handleGetECodeBadRequest    func() int
	handleGetECodeInternalError func() int
	handleErrToECode            func(error) int
	handleErrToHttpStatus       func(error) int
	handleECodeToStr            func(int) string
	handleNewRequestId          func() string
)

func init() {
	handleGetECodeSuccess = func() int {
		return http.StatusOK
	}

	handleGetECodeBadRequest = func() int {
		return http.StatusBadRequest
	}

	handleGetECodeInternalError = func() int {
		return http.StatusInternalServerError
	}

	handleErrToECode = func(err error) int {
		if err == nil {
			return handleGetECodeSuccess()
		}
		return handleGetECodeInternalError()
	}

	handleErrToHttpStatus = handleErrToECode

	handleECodeToStr = func(status int) string {
		str := http.StatusText(status)
		if len(str) == 0 {
			str = "Unknown"
		}
		return str
	}

	handleNewRequestId = func() string {
		return strings.ToUpper(uuid.NewV4().String())
	}
}

func RegisterHandle(funcGetECodeSuccess, funcGetECodeInternalError, funcGetECodeBadRequest func() int,
	funcErrToECode, funcErrToHttpStatus func(error) int,
	funcECodeToStr func(int) string,
	funcNewRequestId func() string) {

	if funcGetECodeSuccess != nil {
		handleGetECodeSuccess = funcGetECodeSuccess
	}
	if funcGetECodeInternalError != nil {
		handleGetECodeInternalError = funcGetECodeInternalError
	}
	if funcGetECodeBadRequest != nil {
		handleGetECodeBadRequest = funcGetECodeBadRequest
	}
	if funcErrToECode != nil {
		handleErrToECode = funcErrToECode
	}
	if funcErrToHttpStatus != nil {
		handleErrToHttpStatus = funcErrToHttpStatus
	}
	if funcECodeToStr != nil {
		handleECodeToStr = funcECodeToStr
	}
	if funcNewRequestId != nil {
		handleNewRequestId = funcNewRequestId
	}
}
