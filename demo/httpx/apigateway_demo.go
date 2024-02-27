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
