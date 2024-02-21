package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/utils"
	"net/http"
	"os"
)

var (
	handleGetECodeSuccess       func() int
	handleGetECodeInternalError func() int
	handleErrToECode            func(error) int
	handleErrToHttpStatus       func(error) int
	handleECodeToStr            func(int) string
)

func RegisterHandle(funcGetECodeSuccess, funcGetECodeInternalError func() int,
	funcErrToECode, funcErrToHttpStatus func(error) int,
	funcECodeToStr func(int) string) {

	handleGetECodeSuccess = funcGetECodeSuccess
	handleGetECodeInternalError = funcGetECodeInternalError
	handleErrToECode = funcErrToECode
	handleErrToHttpStatus = funcErrToHttpStatus
	handleECodeToStr = funcECodeToStr
}

type JsonResponse struct {
	Status    int     `json:"-"`
	Code      *string `json:"Code,omitempty"`
	CodeInt   *int    `json:"CodeInt,omitempty"`
	Message   *string `json:"Message,omitempty"`
	Result    any     `json:"Result,omitempty"`
	RequestId *string `json:"RequestId,omitempty"`
}

func (e *JsonResponse) ToString() string {
	jsonString, _ := json.Marshal(e)
	return string(jsonString)
}

func (e *JsonResponse) Error() string {
	jsonString, _ := json.Marshal(e)
	return string(jsonString)
}

func (e *JsonResponse) IsNoContent() bool {
	return e.Code == nil && e.CodeInt == nil && e.Result == nil && e.RequestId == nil
}

func GetStatusErrorCode(err error) int {
	if err == nil {
		return handleGetECodeSuccess()
	}
	switch e := err.(type) {
	case *JsonResponse:
		return *e.CodeInt
	default:
		return handleGetECodeSuccess()
	}
}

func WrapperError(err error) error {
	if err == nil {
		return nil
	}
	switch e := err.(type) {
	case *JsonResponse:
		return e
	case *echo.HTTPError:
		return MessageResp(e.Code, handleErrToECode(err), fmt.Sprintf("%v", e.Message))
	case *os.PathError:
		return ErrorResp(handleErrToHttpStatus(err), handleErrToECode(err), err)
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

func MessageResp(status, code int, msg string) error {
	codeStr := handleECodeToStr(code)
	return &JsonResponse{
		Status:  status,
		CodeInt: &code,
		Code:    &codeStr,
		Message: &msg,
	}
}

func ErrorResp(status, code int, err error) error {
	codeStr := handleECodeToStr(code)
	errStr := ""
	var msgPtr *string
	if err != nil {
		errStr = err.Error()
		msgPtr = &errStr
	}

	switch e := err.(type) {
	case *JsonResponse:
		e.Status = status
		e.CodeInt = &code
		e.Code = &codeStr
		return e
	default:
		return &JsonResponse{
			Status:  status,
			CodeInt: &code,
			Code:    &codeStr,
			Message: msgPtr,
		}
	}
}

func SendResp(c echo.Context, resp error) error {

	if resp == nil {
		return c.NoContent(http.StatusOK)
	}

	var e *JsonResponse
	switch {
	case errors.As(resp, &e):
		if e.IsNoContent() {
			return c.NoContent(e.Status)
		}

		if e.RequestId == nil {
			rid := utils.NewRequestId()
			e.RequestId = &rid
		}
		return c.JSON(e.Status, e)
	default:
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"CodeInt":   handleGetECodeInternalError(),
			"Code":      handleECodeToStr(handleGetECodeInternalError()),
			"Message":   e.Error(),
			"RequestId": utils.NewRequestId(),
		})
	}
}
