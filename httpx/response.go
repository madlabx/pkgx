package httpx

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/errcodex"
	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/utils"
)

var errCodeDic errcodex.ErrorCodeDictionaryIf

func init() {
	errCodeDic = &errcodex.DefaultErrCodeDic{}
}

var _ errcodex.ErrorCodeIf = &JsonResponse{}

// JsonResponse should be:
type JsonResponse struct {
	cause error

	Status int `json:"-"`

	Code    string `json:"Code,omitempty"`
	Errno   int    `json:"Errno,omitempty"`
	Message string `json:"Message,omitempty"`

	RequestId string `json:"RequestId,omitempty"`
	Result    any    `json:"Result,omitempty"`
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

// return true while Code is same
func (jr *JsonResponse) Is(target error) bool {
	var ec errcodex.ErrorCodeIf
	if errors.As(target, &ec) {
		return ec.GetCode() == jr.Code
	}

	return false
}

// JsonString won't output
func (jr *JsonResponse) JsonString() string {
	//TODO refactor
	njr := jr.Copy()

	if njr.Message == "" && njr.Result == nil {
		njr.Message = njr.Error()
	}

	return utils.ToString(njr)
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

	if jr.cause != nil {
		builder.WriteString(fmt.Sprintf(", Err:%s", jr.cause))
	}
	return builder.String()
}

// Error output err at first, then Message, then flatErrString
func (jr *JsonResponse) Error() string {
	if jr.cause != nil {
		return jr.cause.Error()
	}

	if jr.Message != "" {
		return jr.Message
	}

	if !jr.IsOK() {
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
			_, _ = fmt.Fprintf(s, "%s,\"Cause\":\"%+v\"}", newRawJson, jr.cause)
			return
		}
		fallthrough
	case 's':
		_, _ = fmt.Fprintf(s, "%s", jr.JsonString())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", jr.JsonString())
	}
}

// WithMessagef append Message, won't impact IsOK()
func (jr *JsonResponse) WithMessagef(format string, a ...any) error {
	if format == "" {
		return jr
	}

	if jr.Message == "" {
		jr.Message = fmt.Sprintf(format, a...)
	} else {
		jr.Message += fmt.Sprintf(","+format, a...)
	}

	return jr
}

// WithError to be simple, do overwrite err, will impact IsOK()
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
		return Wrap(err)
	} else {
		if jr.cause == nil {
			newJr := &JsonResponse{}
			if errors.As(err, &newJr) {
				jr.cause = newJr.Cause()
			} else {
				jr.cause = err
			}
		} else {
			jr.cause = errors.WrapfWithRelativeStackDepth(jr.cause, depth, err.Error())
		}
	}

	return jr
}

// WithErrorf to be simple, do overwrite err, will impact IsOK()
func (jr *JsonResponse) WithErrorf(format string, a ...any) error {
	if format == "" {
		return jr
	}

	depth := 1
	if jr.IsOK() {
		//original jr is OK, update with new error
		return Wrap(errors.Errorf(format, a...))
	} else {
		if jr.cause == nil {
			jr.cause = errors.Errorf(format, a...)
		}
		jr.cause = errors.WrapfWithRelativeStackDepth(jr.cause, depth, format, a...)
		return jr
	}
}

// WithResult to be simple, do overwrite
func (jr *JsonResponse) WithResult(result any) *JsonResponse {
	jr.Result = result
	return jr
}

func (jr *JsonResponse) clone(obj *JsonResponse) {
	jr.Errno = obj.Errno
	jr.Code = obj.Code
	jr.cause = obj.cause
	jr.Status = obj.Status
	jr.RequestId = obj.RequestId
	jr.Result = obj.Result
}

func (jr *JsonResponse) Copy() *JsonResponse {
	//TODO whether need deep copy cause... no need to do, normally error type will not change
	obj := *jr
	return &obj
}

func (jr *JsonResponse) cjson(c echo.Context) error {
	if jr.Code == "" && jr.Errno == 0 && jr.Result == nil {
		return c.NoContent(jr.Status)
	}

	if jr.Message == "" && jr.Result == nil {
		jr.Message = jr.Error()
	}

	err := c.JSON(jr.Status, jr)
	if err != nil {
		err = jr.Unwrap()
	}

	return err
}

// WithStack add cause with StackTrace
func (jr *JsonResponse) WithStack(relativeDepths ...int) *JsonResponse {
	if jr.cause == nil {
		jr.cause = jr.Cause()
	}
	relativeDepth := 1
	if len(relativeDepths) > 0 {
		relativeDepth = relativeDepths[0]
	}
	jr.cause = errors.WrapWithRelativeStackDepth(jr.cause, relativeDepth)

	return jr
}

func (jr *JsonResponse) Unwrap() error {
	return jr.cause
}

func (jr *JsonResponse) IsOK() bool {
	//jr.Status is not reliable
	return jr.Code == errCodeDic.GetSuccess().GetCode()
}

// Cause return children cause. will not recursively retrieve cause.Cause
func (jr *JsonResponse) Cause() error {
	if jr.cause != nil {
		//TODO consider to return jr.cause.Cause()??
		return jr.cause
	}

	if !jr.IsOK() {
		return fmt.Errorf(jr.flatErrString())
	}

	return nil
}

// wrap with JsonResponse, with Stack
func Wrap(err error) *JsonResponse {
	if err == nil {
		return nil
	}
	var (
		jr *JsonResponse
		eh *echo.HTTPError
		ec errcodex.ErrorCodeIf
	)
	switch {
	case errors.As(err, &jr):
		return jr.WithStack(1)
	case errors.As(err, &ec):
		jr = &JsonResponse{
			Status: ec.GetHttpStatus(),
			Code:   ec.GetCode(),
			Errno:  ec.GetErrno(),
			cause:  errors.WrapWithRelativeStackDepth(ec.Unwrap(), 1),
		}
	case errors.As(err, &eh):
		jr = &JsonResponse{
			Status: eh.Code,
			Code:   http.StatusText(eh.Code),
			Errno:  eh.Code,
			cause:  errors.WrapWithRelativeStackDepth(eh, 1),
		}

		//TODO 对于PathError是否还需要单独处理，不打印出path
	//case errors.As(cause, &ep):
	//	ne = ErrorResp(handleErrToHttpStatus(cause), handleErrToECode(cause), cause)
	default:
		jr = &JsonResponse{
			Status: http.StatusInternalServerError,
			Code:   http.StatusText(http.StatusInternalServerError),
			Errno:  http.StatusInternalServerError,
			cause:  errors.WrapWithRelativeStackDepth(err, 1),
		}
	}

	return jr
}

func RegisterErrCodeDictionary(dic errcodex.ErrorCodeDictionaryIf) {
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

//
//func TrimHttpStatusText(status int) string {
//	trimmedSpace := strings.Replace(http.StatusText(status), " ", "", -1)
//	trimmedSpace = strings.Replace(trimmedSpace, "-", "", -1)
//	return trimmedSpace
//}
