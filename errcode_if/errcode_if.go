package errcode_if

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
