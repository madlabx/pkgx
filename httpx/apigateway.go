package httpx

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/madlabx/pkgx/log"
	"github.com/madlabx/pkgx/viperx"
	"github.com/sirupsen/logrus"
)

const (
	DefaultB0dyBufferSize = 2000
)

type LogConfig struct {
	FileConfig     log.FileConfig
	Level          string
	BodyBufferSize int `vx_default:"2000"`
}

type ApiGateway struct {
	Ctx       context.Context
	Echo      *echo.Echo
	Logger    *logrus.Logger
	LogConf   *LogConfig
	LogFormat logrus.Formatter
}

func NewApiGateway(pCtx context.Context, lc *LogConfig, logFormat logrus.Formatter) (*ApiGateway, error) {
	agw := &ApiGateway{
		Ctx:       context.WithoutCancel(pCtx),
		Echo:      echo.New(),
		LogConf:   lc,
		LogFormat: logFormat,
	}

	//if lc == nil, log to log.StandardLogger
	if err := agw.initAccessLog(); err != nil {
		return nil, err
	}

	agw.configEcho()
	return agw, nil
}

func (agw *ApiGateway) Run(ip, port string) error {
	return agw.startEcho(fmt.Sprintf("%s:%s", ip, port))
}

func (agw *ApiGateway) Stop() error {
	return agw.shutdownEcho()
}

func (agw *ApiGateway) initAccessLog() error {
	if agw.LogConf == nil {
		agw.LogConf = &LogConfig{}
		agw.Logger = log.StandardLogger()
	} else {
		agw.Logger = log.NewLogger(agw.Ctx, agw.LogConf.FileConfig)
	}

	lc := agw.LogConf

	// Set level
	if lc.Level == "" {
		lc.Level = "info" //by default, apply info
	}
	level, err := logrus.ParseLevel(lc.Level)
	if err != nil {
		return err
	}
	agw.Logger.SetLevel(level)

	// Set body format
	if agw.LogFormat == nil {
		agw.LogFormat = &log.TextFormatter{QuoteEmptyFields: true}
	}
	agw.Logger.SetFormatter(agw.LogFormat)

	if lc.BodyBufferSize == 0 {
		lc.BodyBufferSize = DefaultB0dyBufferSize
	}

	return nil
}

func isPrintableTextContent(contentType string) bool {
	if strings.HasPrefix(contentType, "text/") ||
		strings.Contains(contentType, "json") ||
		strings.Contains(contentType, "xml") ||
		strings.Contains(contentType, "html") {
		return true
	}

	return false
}

func (agw *ApiGateway) configEcho() {
	// Tags to construct the Logger format.
	//
	// - time_unix
	// - time_unix_nano
	// - time_rfc3339
	// - time_rfc3339_nano
	// - time_custom
	// - id (Request ID)
	// - remote_ip
	// - uri
	// - host
	// - method
	// - path
	// - protocol
	// - referer
	// - user_agent
	// - status
	// - error
	// - latency (In nanoseconds)
	// - latency_human (Human readable)
	// - bytes_in (Bytes received)
	// - bytes_out (Bytes sent)
	// - header:<NAME>
	// - query:<NAME>
	// - form:<NAME>
	var (
		e = agw.Echo
	)
	e.Logger.SetOutput(log.StandardLogger().Out)
	format := "${time_custom} ${status} ${method} ${latency_human} ${host} ${remote_ip} ${bytes_in} ${bytes_out} ${uri} ${id} ${error}\n"
	e.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		Handler: func(c echo.Context, reqBody []byte, resBody []byte) {
			lq := int(math.Min(float64(len(reqBody)), float64(agw.LogConf.BodyBufferSize)))
			lp := int(math.Min(float64(len(resBody)), float64(agw.LogConf.BodyBufferSize)))

			contentType := c.Response().Header().Get(echo.HeaderContentType)

			if isPrintableTextContent(contentType) {
				agw.Logger.Infof("%v, reqBody[%v]:{%v}, resBody[%v]:{%v}", c.Request().URL.String(), len(reqBody), string(reqBody[:lq]), len(resBody), string(resBody[:lp]))
			} else {
				agw.Logger.Infof("%v, reqBody[%v]:{%v}, resBody[%v]:[Non-printable ContentType:%v]", c.Request().URL.String(), len(reqBody), string(reqBody[:lq]), len(resBody), contentType)
			}
		},
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:           viperx.GetString("sys.accessFormat", format),
		CustomTimeFormat: "2006/01/02 15:04:05.000",
		Output:           agw.Logger.Out,
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		//AllowMethods: []string{Echo.GET, Echo.PUT, Echo.POST, Echo.DELETE},
	}))

	//TODO 检查是否可以恢复。不注释回无法下载css
	//e.Use(func(next Echo.HandlerFunc) Echo.HandlerFunc {
	//	return func(c Echo.Context) error {
	//		c.Response().Header().Set("Content-Security-Policy", `default-src 'self'; style-src 'unsafe-inline';`)
	//		return next(c)
	//	}
	//})
}

func (agw *ApiGateway) startEcho(addr string) error {
	return agw.Echo.Start(addr)
}

func (agw *ApiGateway) shutdownEcho() error {
	ctx, cancel := context.WithTimeout(agw.Ctx, 5*time.Second)
	defer cancel()
	return agw.Echo.Shutdown(ctx)
}

func (agw *ApiGateway) RoutesToString() string {
	e := agw.Echo
	routes := e.Routes()
	sort.Slice(routes, func(i, j int) bool { return routes[i].Path < routes[j].Path })

	var builder strings.Builder
	for _, r := range routes {
		builder.WriteString(fmt.Sprintf("%-10v %-20v %v\n", r.Method, r.Path, r.Name))
	}

	return builder.String()
}
