package httpx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/errcode"
	"github.com/madlabx/pkgx/errors"
)

var errCodeDic errcode.ErrorCodeDictionaryIf

func init() {
	errCodeDic = &errcode.DefaultErrCodeDic{}
}

// JsonResponse 如果err为空
type JsonResponse struct {
	err error

	Status int `json:"-"`

	Code    string `json:"Code,omitempty"`
	Errno   int    `json:"Errno,omitempty"`
	Message string `json:"Message,omitempty"`

	RequestId string `json:"RequestId,omitempty"`
	Result    any    `json:"Result,omitempty"`
}

func (jr *JsonResponse) String() string {
	jsonString, _ := json.Marshal(jr)
	return string(jsonString)
}

func (jr *JsonResponse) Error() string {
	if jr.err != nil {
		return jr.err.Error()
	}

	return ""
}

// WithError to be simple, do overwrite
func (jr *JsonResponse) WithError(err error, depths ...int) *JsonResponse {
	if err == nil {
		return jr
	}
	
	depth := 1
	if len(depths) > 0 {
		depth = depths[0]
	}

	newJr := &JsonResponse{}
	if errors.As(err, &newJr) {
		jr.err = newJr.ToError()
	} else {
		jr.err = errors.WrapWithRelativeStackDepth(err, depth)
	}

	//if jr.err != nil {
	//	jr.err = errors.Join(jr.err, err)
	//}

	return jr
}

// WithMsg to be simple, do overwrite
func (jr *JsonResponse) WithErrorf(format string, a ...any) *JsonResponse {
	jr.err = errors.WrapWithRelativeStackDepth(fmt.Errorf(format, a...), 1)
	return jr
}

// WithResult to be simple, do overwrite
func (jr *JsonResponse) WithResult(result any) *JsonResponse {
	jr.Result = result
	return jr
}

func (jr *JsonResponse) json(c echo.Context) error {
	if jr.Code == "" && jr.Errno == 0 && jr.Result == nil {
		return c.NoContent(jr.Status)
	}
	jr.Message = jr.Error()
	
	err := c.JSON(jr.Status, jr)
	if err != nil {
		err = jr.Unwrap()
	}

	return err
}

func (jr *JsonResponse) Unwrap() error {
	return jr.err
}

func (jr *JsonResponse) IsOK() bool {
	return jr.Status == errCodeDic.GetSuccess().GetHttpStatus()
}

func (jr *JsonResponse) ToError() error {
	if jr.err != nil {
		return jr.err
	}

	if !jr.IsOK() {
		return fmt.Errorf("Errno:%v, Code:%v", jr.Errno, jr.Code)
	}

	return nil
}

func Wrap(err error) *JsonResponse {
	if err == nil {
		return nil
	}
	var (
		jr *JsonResponse
		eh *echo.HTTPError
		ec errcode.ErrorCodeIf
	)
	switch {
	case errors.As(err, &jr):
		return jr
	case errors.As(err, &ec):
		jr = &JsonResponse{
			Status: ec.GetHttpStatus(),
			Code:   ec.GetCode(),
			Errno:  ec.GetErrno(),
			err: errors.WrapWithRelativeStackDepth(ec.Unwrap(), 1),
		}
	case errors.As(err, &eh):
		jr = &JsonResponse{
			Status: eh.Code,
			Code:   http.StatusText(eh.Code),
			Errno:  eh.Code,
			err:    errors.WrapWithRelativeStackDepth(eh, 1),
		}

		//TODO 对于PathError是否还需要单独处理，不打印出path
	//case errors.As(err, &ep):
	//	ne = ErrorResp(handleErrToHttpStatus(err), handleErrToECode(err), err)
	default:
		jr = &JsonResponse{
			Status: http.StatusInternalServerError,
			Code:   http.StatusText(http.StatusInternalServerError),
			Errno:  http.StatusInternalServerError,
			err:    errors.WrapWithRelativeStackDepth(err, 1),
		}
	}

	return jr
}

func RegisterErrCodeDictionary(dic errcode.ErrorCodeDictionaryIf) {
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

func ResultResp(status int, code errcode.ErrorCodeIf, result any) *JsonResponse {
	return &JsonResponse{
		Status: status,
		Errno:  code.GetErrno(),
		Code:   code.GetCode(),
		Result: result,
	}
}

func StatusResp(status int) *JsonResponse {
	return &JsonResponse{
		Status: status,
	}
}

func errStrResp(status int, code errcode.ErrorCodeIf, format string, a ...any) *JsonResponse {

	return &JsonResponse{
		err:    errors.Errorf(format, a...),
		Status: status,
		Errno:  code.GetErrno(),
		Code:   code.GetCode(),
	}
}

//
//func ErrorResp(status int, code string, err error) *JsonResponse {
//	var (
//		msgPtr  *string
//		jr      *JsonResponse
//		codeStr = handleECodeToStr(code)
//		errStr  = ""
//	)
//
//	if err != nil {
//		errStr = err.Error()
//		msgPtr = &errStr
//	}
//
//	switch {
//	case errors.As(err, &jr):
//		jr.Status = status
//		jr.Errno = &code
//		jr.Code = &codeStr
//		return jr
//	default:
//		return &JsonResponse{
//			err:     err,
//			Status:  status,
//			Errno: &code,
//			Code:    &codeStr,
//			Message: msgPtr,
//		}
//	}
//}

func NewEtag(modTime time.Time, length int64) string {
	timestampHex := strconv.FormatInt(modTime.Unix(), 16)
	// 将长度转换为16进制
	lengthHex := strconv.FormatInt(length, 16)
	// 将两部分用'-'连接
	return timestampHex + "-" + lengthHex
}

// CheckIfNoneMatch if Etag same, true
func CheckIfNoneMatch(r *http.Request, currentEtag string) bool {
	inm := r.Header.Get("If-None-Match")
	if inm == "" {
		return false
	}
	return etagWeakMatch(inm, currentEtag)
}

// etagWeakMatch reports whether a and b match using weak ETag comparison.
// Assumes a and b are valid ETags.
func etagWeakMatch(a, b string) bool {
	return strings.TrimPrefix(a, "W/") == strings.TrimPrefix(b, "W/")
}

var unixEpochTime = time.Unix(0, 0)

// isZeroTime reports whether t is obviously unspecified (either zero or Unix()=0).
func isZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpochTime)
}

// CheckIfModifiedSince if not modified, return true
func CheckIfModifiedSince(r *http.Request, modtime time.Time) bool {
	if r.Method != "GET" && r.Method != "HEAD" {
		return false
	}
	ims := r.Header.Get("If-Modified-Since")
	if ims == "" || isZeroTime(modtime) {
		return false
	}
	t, err := http.ParseTime(ims)
	if err != nil {
		return false
	}
	// The Last-Modified header truncates sub-second precision so
	// the modTime needs to be truncated too.
	modtime = modtime.Truncate(time.Second)
	if ret := modtime.Compare(t); ret <= 0 {
		return true
	}
	return false
}

func SendResp(c echo.Context, resp error) (err error) {
	if c.Response().Committed {
		return resp
	}

	rid := errCodeDic.NewRequestId()
	c.Response().Header().Set(echo.HeaderXRequestID, rid)
	if resp == nil {
		return c.NoContent(http.StatusOK)
	}

	jr := Wrap(resp)
	jr.RequestId = rid

	return jr.json(c)
}

func ServeContent(w http.ResponseWriter, req *http.Request, name string, modTime time.Time, length int64, content io.ReadSeeker) {
	rid := errCodeDic.NewRequestId()
	w.Header().Set(echo.HeaderXRequestID, rid)
	w.Header().Set("Etag", NewEtag(modTime, length))
	http.ServeContent(w, req, name, modTime, content)
}
