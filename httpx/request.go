package httpx

import (
	"net/http"
	"reflect"

	"github.com/labstack/echo"
)

func ValidateMust(input interface{}, keys ...string) error {
	v := reflect.ValueOf(input)
	t := reflect.TypeOf(input)

	for _, key := range keys {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i)
			if field.Name == key && value.IsZero() {
				return ErrStrResp(http.StatusBadRequest, handleGetECodeBadRequest(), "Need "+key)
			}
		}
	}

	return nil
}

func QueryMustParam(c echo.Context, key string) (string, error) {
	var err error
	value := c.QueryParam(key)
	if len(value) == 0 {
		err = ErrStrResp(http.StatusBadRequest, handleGetECodeBadRequest(), "Missing "+key)
	}

	return value, err
}

func QueryOptionalParam(c echo.Context, key string) (string, bool) {
	value := c.QueryParam(key)
	return value, len(value) != 0
}

func BindAndValidate(c echo.Context, req any) error {
	//获取c的请求

	//用reflect遍历req的结构体的成员，如果成员是结构体，需要做递归遍历，需要处理req是指针的情况。
	//举例结构体成员有定义tag  如下
	//  Name       string `hx_place:"body" hx_mandatory:"true" hx_name:"host_name" hx_default:"default_name" hx_range:"alice,bob"`
	//	TaskId     int64  `hx_place:"body" hx_mandatory:"false" hx_name:"task_id" hx_default:"7" hx_range:"0-21"
	//如果hx_place为body，从echo.Context的请求里获取该成员，名称为hx_name，是否必须为hx_mandatory，默认值为hx_default,允许值范围为hx_range
	//如果hx_place为query，从echo.Context的uri的query parameter里获取该成员，名称为hx_name，是否必须为hx_mandatory，默认值为hx_default,允许值范围为hx_range
	//
	//type TusReq struct {
	//	Name       string `hx_place:"query" hx_mandatory:"true" hx_name:"host_name" hx_default:"default_name" hx_range:"alice,bob"`
	//	TaskId     int64  `hx_place:"body" hx_mandatory:"false" hx_name:"task_id" hx_default:"7" hx_range:"0-21"`
	//	CreateTime int64  `hx_flag:"place:body;mandatory:true;range:32-"`
	//	Timeout    int64  `hx_flag:";true;;32-"`
	//}
	//
}
