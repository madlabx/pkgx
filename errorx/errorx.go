package errorx

import (
	"encoding/json"
	"errors"
	"net/url"
)

const (
	ECODE_SUCCESS      = 0
	ECODE_ORIGIN_ERROR = 5555

	// error in API
	ECODE_BAD_REQUEST_PARAM     = 400
	ECODE_WRONG_SIGN            = 401
	ECODE_RESOURCE_IS_NOT_READY = 503
	ECODE_NOT_FOUND             = 404
	ECODE_TIMEOUT               = 504
	ECODE_FAILED_HTTP_REQUEST   = 3004
	ECODE_ALREADY_EXISTING      = 3005
	ECODE_CONNECT_TIMEOUT       = 4040
	ECODE_INTERNAL_ERROR        = 5000
	ECODE_CSV_FILE_WRONG        = 5302
	ECODE_DB_FAILURE            = 5303
	ECODE_WRONG_STRATEGY_RESULT = 5304
)

var ecodeMap map[int]string = map[int]string{
	ECODE_SUCCESS:               "success",
	ECODE_ORIGIN_ERROR:          "外部组件错误",
	ECODE_WRONG_SIGN:            "签名错误",
	ECODE_NOT_FOUND:             "找不到对象",
	ECODE_TIMEOUT:               "请求超时",
	ECODE_FAILED_HTTP_REQUEST:   "HTTP请求失败",
	ECODE_ALREADY_EXISTING:      "对象已存在",
	ECODE_BAD_REQUEST_PARAM:     "请求参数错误",
	ECODE_RESOURCE_IS_NOT_READY: "系统资源未准备好",
	ECODE_CONNECT_TIMEOUT:       "连接超时",
	ECODE_CSV_FILE_WRONG:        "错误的CSV文件",
	ECODE_DB_FAILURE:            "数据库错误",
	ECODE_WRONG_STRATEGY_RESULT: "运行结果错误",
	ECODE_INTERNAL_ERROR:        "内部错误",
}

func ErrStr(code int) string {
	es, ok := ecodeMap[code]
	if !ok {
		return "Unknown error"
	}
	return es
}

type StatusError struct {
	Status  int    `json:"Status"`
	Code    int    `json:"Code"`
	Message string `json:"Message"`
	CodeStr string `json:"CodeStr"`
}

func (e *StatusError) Error() string {
	jsonString, _ := json.Marshal(e)
	return string(jsonString)
}

func UnWrapperError(err error) *StatusError {
	switch e := err.(type) {
	case *StatusError:
		return e
	case *url.Error:
		if e.Timeout() {
			return NewStatusError(504, ECODE_TIMEOUT, e)
		}
		return NewStatusError(500, ECODE_ORIGIN_ERROR, err)
	default:
		return NewStatusError(500, ECODE_ORIGIN_ERROR, err)
	}
}

func GetStatusErrorCode(err error) int {
	if err == nil {
		return ECODE_SUCCESS
	}
	switch e := err.(type) {
	case *StatusError:
		return e.Code
	default:
		return ECODE_SUCCESS
	}
}

func WrapperError(err error, newString string) error {
	return errors.New(newString + "->" + err.Error() + "")
}

func NewStatusError(status, code int, err error) *StatusError {
	switch e := err.(type) {
	case *StatusError:
		return e
	default:
		return &StatusError{
			Status:  status,
			Code:    code,
			CodeStr: ErrStr(code),
			Message: err.Error(),
		}
	}
}

func NewStatusErrStr(status, code int, errstr string) *StatusError {
	return &StatusError{
		Status:  status,
		Code:    code,
		CodeStr: ErrStr(code),
		Message: errstr,
	}
}

func Err400(err string) *StatusError {
	return NewStatusErrStr(400, ECODE_BAD_REQUEST_PARAM, err)
}

func Err404(err string) *StatusError {
	return NewStatusErrStr(404, ECODE_BAD_REQUEST_PARAM, err)
}

func IErrStr(status int, errstr string) *StatusError {
	return NewStatusErrStr(status, ECODE_INTERNAL_ERROR, errstr)
}
