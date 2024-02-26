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

	handleECodeToStr = http.StatusText

	handleNewRequestId = func() string {
		return strings.ToUpper(uuid.NewV4().String())
	}
}

func setCb(f1, f2 any) {
	if f2 != nil {
		f1 = f2
	}
}

func RegisterHandle(funcGetECodeSuccess, funcGetECodeInternalError, funcGetECodeBadRequest func() int,
	funcErrToECode, funcErrToHttpStatus func(error) int,
	funcECodeToStr func(int) string,
	funcNewRequestId func() string) {

	setCb(handleGetECodeSuccess, funcGetECodeSuccess)
	setCb(handleGetECodeInternalError, funcGetECodeInternalError)
	setCb(handleGetECodeBadRequest, funcGetECodeBadRequest)
	setCb(handleErrToECode, funcErrToECode)
	setCb(handleErrToHttpStatus, funcErrToHttpStatus)
	setCb(handleECodeToStr, funcECodeToStr)
	setCb(handleNewRequestId, funcNewRequestId)
}
