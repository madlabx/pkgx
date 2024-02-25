package main

import (
	"context"
	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/apigateway"
	"github.com/madlabx/pkgx/errorx"
	"github.com/madlabx/pkgx/httpx"
	"github.com/madlabx/pkgx/log"
)

func main() {
	log.New()
	agw, err := apigateway.New(context.Background(), &apigateway.LogConfig{})
	errorx.CheckFatalError(err)

	e := agw.Echo

	e.Any("/status", func(ctx echo.Context) error {
		log.Info("Request Status")
		return httpx.SendResp(ctx, httpx.StatusResp(200))
	})

	log.Errorf("Routes:\n%v", agw.RoutesToString())
	defer func() {
		_ = agw.Stop()
	}()

	if err := agw.Run("127.0.0.1", "8080"); err != nil {
		log.Error(err)
	}
}
