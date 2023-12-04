package errorx

import (
	"encoding/json"
	"errors"
	"net/url"
)

const (
	ECODE_SUCCESS      = 0
	ECODE_ORIGIN_ERROR = 5555

	// error in input, IPC
	ECODE_FAILED_GET_ONVIF_URI = 10000

	// error in output
	ECODE_FAILED_START_ONVIF_SERVICE = 20001

	// error in transcoder
	ECODE_FAILED_START_TRANSCODER = 30001

	// error in API
	ECODE_BAD_REQUEST_PARAM       = 400
	ECODE_WRONG_SIGN              = 401
	ECODE_RESOURCE_IS_NOT_READY   = 503
	ECODE_NOT_FOUND               = 404
	ECODE_TIMEOUT                 = 504
	ECODE_FAILED_HTTP_REQUEST     = 3004
	ECODE_ALREADY_EXISTING        = 3005
	ECODE_WRONG_ACTION            = 3006
	ECODE_FAILED_REQUEST_TO_AGENT = 3007
	ECODE_WRONG_BODY_FROM_AGENT   = 3008
	ECODE_FAILURE_START_FFMPEG    = 3009
	ECODE_FAILURE_STOP_FFMPEG     = 3010
	ECODE_FAILED_TO_DEL_CHAN      = 3011
	ECODE_FAILED_NO_RESOURCE      = 3012
	ECODE_RESOURCE_NOT_ENOUGH     = 3014
	ECODE_CANNOT_CONNECT_BLADE    = 3015
	ECODE_INTERNAL_ERROR          = 5000
	ECODE_LIST_VOD_TIMEOUT        = 5002
	ECODE_GET_PRESET_TIMEOUT      = 5003
	ECODE_WRONG_CONFIG            = 5004
	ECODE_IPC_ERROR               = 4000
	ECODE_CONN_FAILED             = 4001
	ECODE_INVALID_PROTO           = 4002
	ECODE_NOT_200_OK              = 4003
	ECODE_DATA_NOT_TCP            = 4004
	ECODE_MODE_NOT_ACTIVE         = 4005
	ECODE_ZERO_SSRC               = 4006
	ECODE_NO_SSRC                 = 4007
	ECODE_INVALID_SDP             = 4008
	ECODE_KA_TIMEOUT              = 4009
	ECODE_DEV_UNREG               = 4012
	ECODE_SUMNUM_NOT_MATCH        = 4013
	ECODE_BAD_XML_BODY            = 4014
	ECODE_DEV_LOCKED              = 4015
	ECODE_FAILED_IN_GB_SCHED      = 5300
	ECODE_BAD_SDP                 = 5301
	ECODE_CONNECT_TIMEOUT         = 4040
	ECODE_CSV_FILE_WRONG          = 5302
	ECODE_DB_FAILURE              = 5303
	ECODE_WRONG_STRATEGY_RESULT   = 5304
)

var ecodeMap map[int]string = map[int]string{
	ECODE_SUCCESS:                    "success",
	ECODE_ORIGIN_ERROR:               "未知的外部错误",
	ECODE_FAILED_GET_ONVIF_URI:       "无法获取ONVIF URI",
	ECODE_FAILED_START_ONVIF_SERVICE: "无法启动ONVIF服务",
	ECODE_FAILED_START_TRANSCODER:    "无法启动转码组件",
	ECODE_WRONG_SIGN:                 "签名错误",
	ECODE_NOT_FOUND:                  "找不到对象",
	ECODE_TIMEOUT:                    "请求超时",
	ECODE_FAILED_HTTP_REQUEST:        "HTTP请求失败",
	ECODE_ALREADY_EXISTING:           "对象已存在",
	ECODE_WRONG_ACTION:               "不允许的操作",
	ECODE_FAILED_REQUEST_TO_AGENT:    "无法请求Agent服务",
	ECODE_WRONG_BODY_FROM_AGENT:      "Agent服务返回的body错误",
	ECODE_FAILURE_START_FFMPEG:       "无法启动转码组件",
	ECODE_FAILURE_STOP_FFMPEG:        "无法停止转码组件",
	ECODE_FAILED_TO_DEL_CHAN:         "无法删除通道",
	ECODE_FAILED_NO_RESOURCE:         "板卡或设备资源不足",
	ECODE_CANNOT_CONNECT_BLADE:       "无法连接到板块",
	ECODE_INTERNAL_ERROR:             "内部错误",
	ECODE_BAD_REQUEST_PARAM:          "请求参数错误",
	ECODE_LIST_VOD_TIMEOUT:           "list vod timeout",
	ECODE_GET_PRESET_TIMEOUT:         "get preset timeout",
	ECODE_WRONG_CONFIG:               "配置错误",
	ECODE_IPC_ERROR:                  "摄像头侧有错误",
	ECODE_CONN_FAILED:                "无法连接",
	ECODE_INVALID_PROTO:              "协议错误",
	ECODE_NOT_200_OK:                 "非200返回",
	ECODE_DATA_NOT_TCP:               "不支持TCP",
	ECODE_MODE_NOT_ACTIVE:            "不支持主动模式",
	ECODE_ZERO_SSRC:                  "错误的SSRC值0",
	ECODE_NO_SSRC:                    "缺少SSRC",
	ECODE_INVALID_SDP:                "错误SDP",
	ECODE_KA_TIMEOUT:                 "心跳超时",
	ECODE_DEV_UNREG:                  "设备未注册",
	ECODE_SUMNUM_NOT_MATCH:           "错误的XML消息：sumnum和数量不匹配",
	ECODE_BAD_XML_BODY:               "错误的XML消息",
	ECODE_RESOURCE_IS_NOT_READY:      "系统资源未准备好",
	ECODE_RESOURCE_NOT_ENOUGH:        "系统容量不够",
	ECODE_FAILED_IN_GB_SCHED:         "failed in gb sched",
	ECODE_BAD_SDP:                    "错误的sdp消息",
	ECODE_CONNECT_TIMEOUT:            "连接超时",
	ECODE_CSV_FILE_WRONG:             "错误的CSV文件",
	ECODE_DB_FAILURE:                 "数据库错误",
	ECODE_WRONG_STRATEGY_RESULT:      "运行结果错误",
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
