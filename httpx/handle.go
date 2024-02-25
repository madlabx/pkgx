package httpx

import "net/http"

var (
	handleGetECodeSuccess       func() int
	handleGetECodeBadRequest    func() int
	handleGetECodeInternalError func() int
	handleErrToECode            func(error) int
	handleErrToHttpStatus       func(error) int
	handleECodeToStr            func(int) string
	handleNewRequestId func(opt any) string
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
}

func RegisterHandle(funcGetECodeSuccess, funcGetECodeInternalError, funcGetECodeBadRequest func() int,
		funcErrToECode, funcErrToHttpStatus func(error) int,
		funcECodeToStr func(int) string)
		funcNewRequestId()string{

	handleGetECodeSuccess = funcGetECodeSuccess
	handleGetECodeInternalError = funcGetECodeInternalError
	handleGetECodeBadRequest = funcGetECodeBadRequest
	handleErrToECode = funcErrToECode
	handleErrToHttpStatus = funcErrToHttpStatus
	handleECodeToStr = funcECodeToStr
	handleNewRequestId = funcNewRequestId
}
