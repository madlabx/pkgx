package main

import (
	"context"

	"github.com/madlabx/pkgx/lumberjackx"

	"github.com/madlabx/pkgx/viperx"

	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/httpx"
	"github.com/madlabx/pkgx/log"
)

type TusReq struct {
	Name       string `hx_place:"query" hx_mandatory:"true" hx_name:"host_name" hx_default:"default_name" hx_range:"alice,bob"`
	TaskId     int64  `hx_place:"body" hx_mandatory:"false" hx_name:"task_id" hx_default:"7" hx_range:"0-21"`
	CreateTime int64  `hx_flag:"place:body;mandatory:true;range:32-"`
	Timeout    int64  `hx_flag:";true;;32-"`
}

func main() {
	log.New()
	log.SetOutput(&lumberjackx.Logger{
		Ctx:        context.Background(),
		Filename:   "./agw.log",
		MaxSize:    viperx.GetInt("sys.logMaxSize", 10), // megabytes
		MaxBackups: viperx.GetInt("sys.logMaxBackups", 5),
		MaxAge:     viperx.GetInt("sys.logMaxAge", 1), //days
		Compress:   true,                              // disabled by default
		LocalTime:  true,
	})
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
		if err := httpx.BindAndValidate(ctx.Request(), &req); err != nil {
			log.Info("Failed to bind, error:%v", err)
			return httpx.SendResp(ctx, httpx.Wrap(err))
		}
		log.Info("Request Status")
		return httpx.SendResp(ctx, httpx.SuccessResp("1110"))
	})

	log.Errorf("Routes:\n%v", agw.RoutesToString())
	defer func() {
		_ = agw.Stop()
	}()

	if err := agw.Run("127.0.0.1", "8080"); err != nil {
		log.Error(err)
	}
}
