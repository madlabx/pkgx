package main

import (
	"context"
	"strconv"

	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/httpx"
	"github.com/madlabx/pkgx/log"
	"github.com/madlabx/pkgx/viperx"
)

type Boy struct {
	Age  int `hx_default:"5"`
	Name int
}

type Trans struct {
	Bandwidth uint64   `hx_place:"query" hx_must:"false" hx_query_name:"bandwidth" hx_default:"1" hx_range:"1-2"`
	Loss      *float64 `hx_place:"body" hx_must:"false" hx_default:"1.4" hx_range:"1.2-3.4"`
	Loss2     float64  `hx_default:"1.5"`
	LossStr   string   `hx_default:"de"`
	B         []Boy
	C         struct {
		C1 int
		C2 int
		Boy
	}
}

type TusReq struct {
	Name       *string `hx_place:"query" hx_query_name:"host_name" hx_must:"true" hx_default:"22" hx_range:"alice,bob"`
	TaskId     int64   `hx_place:"body" hx_must:"false" hx_default:"12" hx_range:"0-21"`
	CreateTime int64
	CCC        int
	Timeout    Int64 `hx_tag:";;true;;0-21"`
	T          *Trans
}
type Int64 struct {
	Valued bool
	V      int64
}

func (pt Int64) String() string {
	if pt.Valued {
		return strconv.FormatInt(pt.V, 10)
	}
	return "null1"
}

func (pt Int64) MarshalJSON() (b []byte, err error) {
	if pt.Valued {
		return []byte(strconv.FormatInt(pt.V, 10)), nil
	} else {
		return []byte("null"), nil
	}
}
func (pt *Int64) UnmarshalJSON(b []byte) (err error) {
	s := string(b)
	if s == "null" {
		pt.V = 123
		return nil
	}
	pt.V, err = strconv.ParseInt(string(b), 10, 64)
	pt.Valued = true
	return err
}

type OnlyQuery struct {
	Name string `hx_place:"query" hx_query_name:"host_name" hx_must:"true" hx_default:"" hx_range:"alice,bob"`
}

func main() {
	log.SetLoggerOutput(log.StandardLogger(), context.Background(), log.FileConfig{Filename: "stdout"})
	agw, err := httpx.NewApiGateway(context.Background(), "127.0.0.1", "8080", "test", &httpx.LogConfig{
		//Output: "access.log",
	}, nil)

	_ = log.SetLevelStr(viperx.GetString("sys.loglevel", "debug"))

	log.SetFormatter(&log.TextFormatter{
		QuoteEmptyFields: true,
		DisableSorting:   true})

	e := agw.Echo

	e.GET("/v1/hx_tag/api", func(ctx echo.Context) error {
		req := TusReq{CCC: 7}
		if err = httpx.BindAndValidate(ctx, &req); err != nil {
			log.Infof("Failed to bind, error:%v", err)
			return httpx.SendResp(ctx, httpx.Wrap(err))
		}

		log.Infof("Request Status:%v", req)
		log.Infof("Request Status:%v", req.Timeout)
		//return ctx.JSON(200, req)
		return httpx.SendResp(ctx, httpx.SuccessResp(req))
	})

	e.GET("/v1/hx_tag/only_query", func(ctx echo.Context) error {
		req := OnlyQuery{}
		if err := httpx.BindAndValidate(ctx, &req); err != nil {
			log.Infof("Failed to bind, error:%v", err)
			return httpx.SendResp(ctx, httpx.Wrap(err))
		}
		log.Info("Request Status")
		return httpx.SendResp(ctx, httpx.SuccessResp(req))
	})

	log.Errorf("Routes:\n%v", agw.RoutesToString())
	defer func() {
		_ = agw.Stop()
	}()

	if err := agw.Run(); err != nil {
		log.Error(err)
	}
}

//
//package main
//
//import (
//	"net/http"
//	"reflect"
//	"strconv"
//
//	"github.com/labstack/echo"
//)
//
//type Trans struct {
//	Bandwidth uint64
//	Loss      *float64 `json:"loss" default:"2.4"`
//}
//
//func setDefaultLossValue(next echo.HandlerFunc) echo.HandlerFunc {
//	return func(c echo.Context) error {
//		t := new(Trans)
//		if err := c.Bind(t); err != nil {
//			return err
//		}
//
//		v := reflect.ValueOf(t).Elem()
//		field := v.FieldByName("Loss")
//		tag := v.Type().Field(1).Tag.Get("default")
//		if field.IsValid() && field.IsNil() {
//			defaultValue, err := strconv.ParseFloat(tag, 64)
//			if err != nil {
//				return err
//			}
//			field.Set(reflect.ValueOf(&defaultValue))
//		}
//
//		c.Set("trans", t)
//		return next(c)
//	}
//}
//
//func main() {
//	e := echo.New()
//
//	e.Use(setDefaultLossValue)
//
//	e.POST("/", func(c echo.Context) error {
//		t, ok := c.Get("trans").(*Trans)
//		if !ok {
//			return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
//		}
//		return c.JSON(http.StatusOK, t)
//	})
//
//	e.Logger.Fatal(e.Start(":8080"))
//}
