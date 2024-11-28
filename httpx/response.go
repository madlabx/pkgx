package httpx

import (
	"fmt"
	"net/http"
	"strings"

	emperrors "emperror.dev/errors"
	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/errcode_if"
	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/utils"
)

var errCodeDic errcode_if.ErrorCodeDictionaryIf

func init() {
	errCodeDic = &errcode_if.DefaultErrCodeDic{}
}

var _ errcode_if.ErrorCodeIf = &JsonResponse{}

// JsonResponse should be:
type JsonResponse struct {
	err error

	Status int `json:"-"`

	Code    string `json:"Code,omitempty"`
	Errno   int    `json:"Errno,omitempty"`
	Message string `json:"Message,omitempty"`

	RequestId string `json:"RequestId,omitempty"`
	Result    any    `json:"Result,omitempty"`
}

func (jr *JsonResponse) Frames() []emperrors.Frame {
	var st interface{ StackTrace() emperrors.StackTrace }

	if errors.As(jr.err, &st) {
		return st.StackTrace()
	}

	return nil
}

func (jr *JsonResponse) StackTrace() emperrors.StackTrace {
	return errors.StackTrace(jr.err)
}

func (jr *JsonResponse) GetHttpStatus() int {
	return jr.Status
}

func (jr *JsonResponse) GetCode() string {
	return jr.Code
}

func (jr *JsonResponse) GetErrno() int {
	return jr.Errno
}

func (jr *JsonResponse) CompleteMessage() *JsonResponse {
	if jr.Message == "" && jr.Result == nil {
		jr.Message = jr.Error()
	}

	return jr
}

// JsonString won't output
func (jr *JsonResponse) JsonString() string {
	//TODO refactor
	njr := jr.Copy()

	return utils.ToString(njr.CompleteMessage())
}

func (jr *JsonResponse) flatErrString() string {
	var builder strings.Builder
	if jr.Code != "" {
		builder.WriteString(fmt.Sprintf("Code:%v, Errno:%v", jr.Code, jr.Errno))
	}

	if jr.Message != "" {
		builder.WriteString(fmt.Sprintf(", Message:%v", jr.Message))
	}

	if jr.Result != nil {
		builder.WriteString(fmt.Sprintf(", Result:%v", utils.ToString(jr.Result)))
	}

	if jr.err != nil {
		builder.WriteString(fmt.Sprintf(", Err:%s", jr.err))
	}
	return builder.String()
}

// Error output err at first, then Message, then flatErrString
func (jr *JsonResponse) Error() string {
	if jr.err != nil {
		return jr.err.Error()
	}

	if !jr.IsOK() {
		if jr.Message != "" {
			return jr.Message
		}
		return jr.flatErrString()
	}

	return ""
}

// nolint: errcheck
func (jr *JsonResponse) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			rawJson := jr.JsonString()
			newRawJson := rawJson[:len(rawJson)-1]
			_, _ = fmt.Fprintf(s, "%s,\"Cause\":\"%+v\"}", newRawJson, jr.err)
			return
		}
		fallthrough
	case 's':
		_, _ = fmt.Fprintf(s, "%s", jr.JsonString())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", jr.JsonString())
	}
}

// WithMessagef set Message, won't impact IsOK()
// in jr.cjson, if Message is not "", won't return err.Error
func (jr *JsonResponse) WithMsgf(format string, a ...any) *JsonResponse {
	jr.Message = fmt.Sprintf(format, a...)
	return jr
}

func (jr *JsonResponse) WithMessagef(format string, a ...any) error {
	return jr.WithMsgf(format, a...)
}

// WithError if has err, join with Error
// in jr.cjson, if Message is nil, use err.Error
func (jr *JsonResponse) WithError(err error, depths ...int) *JsonResponse {
	if err == nil {
		return jr
	}

	depth := 1
	if len(depths) > 0 {
		depth = depths[0]
	}

	if jr.IsOK() {
		//original jr is OK, update with new error
		*jr = *(Wrap(err))
	} else {
		//overwrite the err
		newJr := &JsonResponse{}
		if errors.As(err, &newJr) {
			jr.err = newJr.Unwrap()
		} else {
			jr.err = errors.WrapWithRelativeStackDepth(err, depth)
		}
	}

	return jr
}

// WithErrorf set err
// in jr.cjson, if Message is nil, use err.Error
func (jr *JsonResponse) WithErrorf(format string, a ...any) *JsonResponse {
	if format == "" {
		return jr
	}

	if jr.IsOK() {
		//original jr is OK, update with new error
		*jr = *(Wrap(errors.ErrorfWithRelativeStackDepth(1, format, a...)))
	} else {
		jr.err = errors.ErrorfWithRelativeStackDepth(1, format, a...)
	}

	return jr
}

// WithResult to be simple, do overwrite
func (jr *JsonResponse) WithResult(result any) *JsonResponse {
	jr.Result = result
	return jr
}

func (jr *JsonResponse) Copy() *JsonResponse {
	//TODO whether need deep copy err... no need to do, normally error type will not change
	obj := *jr
	return &obj
}

func (jr *JsonResponse) cjson(c echo.Context) error {
	if jr.Code == "" && jr.Errno == 0 && jr.Result == nil {
		return c.NoContent(jr.Status)
	}

	err := c.JSON(jr.Status, jr.CompleteMessage())
	if err != nil {
		err = jr.Unwrap()
	}

	return err
}

// WithStack add err with StackTrace
func (jr *JsonResponse) WithStack(relativeDepths ...int) *JsonResponse {
	if jr.IsOK() {
		return jr
	}

	var st errors.ErrorWithStackTrace
	if errors.As(jr.err, &st) {
		return jr
	}

	if jr.err == nil {
		jr.err = errors.NewStd(jr.Error())
	}

	relativeDepth := 1
	if len(relativeDepths) > 0 {
		relativeDepth = relativeDepths[0]
	}

	jr.err = errors.WrapWithRelativeStackDepth(jr.err, relativeDepth)

	return jr
}

// implement interface Is()
func (jr *JsonResponse) Is(target error) bool {
	if target == nil {
		return jr == target
	}
	var ec errcode_if.ErrorCodeIf
	if errors.As(target, &ec) {
		return ec.GetCode() == jr.Code
	}

	return false
}

func (jr *JsonResponse) Unwrap() error {
	if jr == nil {
		return nil
	}
	return jr.err
}

func (jr *JsonResponse) IsOK() bool {
	//jr.Status is not reliable
	return jr.Code == errCodeDic.GetSuccess().GetCode()
}

// Cause return children err. will not recursively retrieve err.Cause
func (jr *JsonResponse) Cause() string {
	if jr.err != nil {
		//TODO consider to return jr.err.Cause()??
		return jr.err.Error()
	}

	if !jr.IsOK() {
		return jr.flatErrString()
	}

	return ""
}

// wrap with JsonResponse, with Stack
func Wrap(err error) *JsonResponse {
	if err == nil {
		return nil
	}

	var (
		ej JsonResponseWrapper
		jr *JsonResponse
		eh *echo.HTTPError
		ec errcode_if.ErrorCodeIf
	)
	switch {
	case errors.As(err, &ej):
		jr = ej.ToHttpXJsonResponse()
		fallthrough
	case errors.As(err, &jr):
		if jr.Status == 0 {
			jr.Status = jr.Errno
		}
		return jr
	case errors.As(err, &ec):
		jr = &JsonResponse{
			Status: ec.GetHttpStatus(),
			Code:   ec.GetCode(),
			Errno:  ec.GetErrno(),
			err:    errors.WrapWithRelativeStackDepth(ec.Unwrap(), 1),
		}
	case errors.As(err, &eh):
		jr = &JsonResponse{
			Status: eh.Code,
			Code:   TrimHttpStatusText(eh.Code),
			Errno:  eh.Code,
			err:    errors.WrapWithRelativeStackDepth(eh, 1),
		}

		//TODO 对于PathError是否还需要单独处理，不打印出path
	//case errors.As(err, &ep):
	//	ne = ErrorResp(handleErrToHttpStatus(err), handleErrToECode(err), err)
	default:
		jr = &JsonResponse{
			Status: http.StatusInternalServerError,
			Code:   TrimHttpStatusText(http.StatusInternalServerError),
			Errno:  http.StatusInternalServerError,
			err:    errors.WrapWithRelativeStackDepth(err, 1),
		}
	}

	return jr
}

func RegisterErrCodeDictionary(dic errcode_if.ErrorCodeDictionaryIf) {
	errCodeDic = dic
}

func SuccessResp(result any) *JsonResponse {
	ec := errCodeDic.GetSuccess()
	return &JsonResponse{
		Status: ec.GetHttpStatus(),
		Errno:  ec.GetErrno(),
		Code:   ec.GetCode(),
		Result: result,
	}
}

//func ResultResp(status int, code errcodex.ErrorCodeIf, result any) *JsonResponse {
//	return &JsonResponse{
//		Status: status,
//		Errno:  code.GetErrno(),
//		Code:   code.GetCode(),
//		Result: result,
//	}
//}

func StatusResp(status int) *JsonResponse {
	return &JsonResponse{
		Status: status,
	}
}

func TrimHttpStatusText(status int) string {
	trimmedSpace := strings.Replace(http.StatusText(status), " ", "", -1)
	trimmedSpace = strings.Replace(trimmedSpace, "-", "", -1)
	return trimmedSpace
}
