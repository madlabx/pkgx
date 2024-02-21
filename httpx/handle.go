package httpx

var (
	handleGetECodeSuccess       func() int
	handleGetECodeBadRequest    func() int
	handleGetECodeInternalError func() int
	handleErrToECode            func(error) int
	handleErrToHttpStatus       func(error) int
	handleECodeToStr            func(int) string
)

func RegisterHandle(funcGetECodeSuccess, funcGetECodeInternalError, funcGetECodeBadRequest func() int,
	funcErrToECode, funcErrToHttpStatus func(error) int,
	funcECodeToStr func(int) string) {

	handleGetECodeSuccess = funcGetECodeSuccess
	handleGetECodeInternalError = funcGetECodeInternalError
	handleGetECodeBadRequest = funcGetECodeBadRequest
	handleErrToECode = funcErrToECode
	handleErrToHttpStatus = funcErrToHttpStatus
	handleECodeToStr = funcECodeToStr
}
