package httpx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/madlabx/pkgx/errors"

	"github.com/labstack/echo"
)

type JsonResponse struct {
	RequestId *string `json:"RequestId,omitempty"`
	Status    int     `json:"-"`
	Code      *string `json:"Code,omitempty"`
	CodeInt   *int    `json:"CodeInt,omitempty"`
	Message   *string `json:"Message,omitempty"`
	Result    any     `json:"Result,omitempty"`
	err       error
}

func (e *JsonResponse) String() string {
	jsonString, _ := json.Marshal(e)
	return string(jsonString)
}

func (e *JsonResponse) Error() string {
	if e.err != nil {
		return e.err.Error()
	}

	return ""
}

func (e *JsonResponse) IsNoContent() bool {
	return e.Code == nil && e.CodeInt == nil && e.Result == nil && e.RequestId == nil
}

func (e *JsonResponse) Unwrap() error {
	return e.err
}

func Wrap(err error) error {
	if err == nil {
		return nil
	}
	var (
		ej *JsonResponse
		eh *echo.HTTPError
	)
	switch {
	case errors.As(err, &ej):
		return ej
	case errors.As(err, &eh):
		return ErrStrResp(eh.Code, handleErrToECode(err), fmt.Sprintf("%v", eh.Message))
		//TODO 对于PathError是否还需要单独处理，不打印出path
	//case errors.As(err, &ep):
	//	ne = ErrorResp(handleErrToHttpStatus(err), handleErrToECode(err), err)
	default:
		return ErrorResp(handleErrToHttpStatus(err), handleErrToECode(err), err)
	}
}

func SuccessResp(result interface{}) error {
	codeInt := handleGetECodeSuccess()
	codeStr := handleECodeToStr(codeInt)
	return &JsonResponse{
		Status:  http.StatusOK,
		CodeInt: &codeInt,
		Code:    &codeStr,
		Result:  result,
	}
}

func ResultResp(status, code int, result interface{}) error {
	codeStr := handleECodeToStr(code)
	return &JsonResponse{
		Status:  status,
		CodeInt: &code,
		Code:    &codeStr,
		Result:  result,
	}
}

func StatusResp(status int) error {
	return &JsonResponse{
		Status: status,
	}
}

func ErrStrResp(status, code int, format string, a ...any) *JsonResponse {
	msg := fmt.Sprintf(format, a...)
	codeStr := handleECodeToStr(code)
	return &JsonResponse{
		err:     errors.New(msg),
		Status:  status,
		CodeInt: &code,
		Code:    &codeStr,
		Message: &msg,
	}
}

func ErrorResp(status, code int, err error) *JsonResponse {
	var (
		msgPtr  *string
		e       *JsonResponse
		codeStr = handleECodeToStr(code)
		errStr  = ""
	)

	if err != nil {
		errStr = err.Error()
		msgPtr = &errStr
	}

	switch {
	case errors.As(err, &e):
		e.Status = status
		e.CodeInt = &code
		e.Code = &codeStr
		return e
	default:
		return &JsonResponse{
			err:     errors.WithStack(err),
			Status:  status,
			CodeInt: &code,
			Code:    &codeStr,
			Message: msgPtr,
		}
	}
}

func newEtag(modTime time.Time, length int64) string {
	timestampHex := strconv.FormatInt(modTime.Unix(), 16)
	// 将长度转换为16进制
	lengthHex := strconv.FormatInt(length, 16)
	// 将两部分用'-'连接
	return timestampHex + "-" + lengthHex
}

func ServeContent(w http.ResponseWriter, req *http.Request, name string, modtime time.Time, length int64, content io.ReadSeeker) {
	rid := handleNewRequestId()
	w.Header().Set(echo.HeaderXRequestID, rid)
	w.Header().Set("Etag", newEtag(modtime, length))
	http.ServeContent(w, req, name, modtime, content)
}

func SendResp(c echo.Context, resp error) (err error) {
	if c.Response().Committed {
		return resp
	}

	rid := handleNewRequestId()
	c.Response().Header().Set(echo.HeaderXRequestID, rid)
	if resp == nil {
		err = c.NoContent(http.StatusOK)
	} else {

		var e *JsonResponse
		switch {
		case errors.As(resp, &e):
			if e.IsNoContent() {
				err = c.NoContent(e.Status)
			} else {

				if e.RequestId == nil {
					e.RequestId = &rid
				}
				err = c.JSON(e.Status, e)
			}
		default:
			err = c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"RequestId": rid,
				"CodeInt":   handleGetECodeInternalError(),
				"Code":      handleECodeToStr(handleGetECodeInternalError()),
				"Message":   resp.Error(),
			})
		}
	}

	//if err != nil {
	//	return err
	//} else {
	//	return resp
	//}
	return err
}
