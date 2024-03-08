package main

import (
	"context"

	"github.com/madlabx/pkgx/viperx"

	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/httpx"
	"github.com/madlabx/pkgx/log"
)

type Trans struct {
	Bandwidth uint64   `hx_place:"query" hx_must:"false" hx_query_name:"bandwidth" hx_default:"default_name" hx_range:"1-2"`
	Loss      *float64 `hx_place:"body" hx_must:"false" hx_default:"1.4" hx_range:"1.2-3.4"`
	Loss2     float64  `hx_default:"1.5"`
	LossStr   string   `hx_default:"de"`
}
type TusReq struct {
	Name       string `hx_place:"query" hx_query_name:"host_name" hx_must:"true" hx_default:"" hx_range:"alice,bob"`
	TaskId     int64  `hx_place:"body" hx_must:"false" hx_default:"" hx_range:"0-21"`
	CreateTime int64  `hx_tag:"query;create;true;;0-21"`
	Timeout    int64  `hx_tag:";;true;;0-21"`
	Trans
}

func main() {
	log.SetLoggerOutput(log.StandardLogger(), context.Background(), log.FileConfig{Filename: "stdout"})
	agw, err := httpx.NewApiGateway(context.Background(), &httpx.LogConfig{
		//Output: "access.log",
	}, nil)
	errors.CheckFatalError(err)

	_ = log.SetLevelStr(viperx.GetString("sys.loglevel", "debug"))

	log.SetFormatter(&log.TextFormatter{
		QuoteEmptyFields: true,
		DisableSorting:   true})

	e := agw.Echo

	httpx.RegisterHandle(func() int { return 0 }, nil, nil, nil, nil, nil, nil)

	e.Any("/v1/file_service/health", func(ctx echo.Context) error {
		req := TusReq{}
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

	if err := agw.Run("127.0.0.1", "8080"); err != nil {
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
