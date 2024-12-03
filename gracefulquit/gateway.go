package gracefulquit

import (
	"context"
	"net/http"

	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/httpx"
)

type Gateway struct {
	agw  *httpx.ApiGateway
	ctx  context.Context
	addr string
	port string
	name string
}

func (gw *Gateway) Run() error {
	err := gw.agw.Run(gw.addr, gw.port)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (gw *Gateway) Stop() error {
	return gw.agw.Stop()
}

func (gw *Gateway) Name() string {
	return gw.name
}
