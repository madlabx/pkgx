package apigateway

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/madlabx/pkgx/log"
	"github.com/madlabx/pkgx/lumberjackx"
	"github.com/madlabx/pkgx/viperx"
	"github.com/sirupsen/logrus"
)

type LogConfig struct {
	Output         string
	Level          string
	Size           int
	BackNum        int
	AgeDays        int
	Formatter      logrus.Formatter
	BodyBufferSize int
}

type ApiGateway struct {
	Ctx    context.Context
	Echo   *echo.Echo
	Logger *logrus.Logger
	Lc     *LogConfig
}

func New(pctx context.Context, lc *LogConfig) (*ApiGateway, error) {
	agw := &ApiGateway{
		Ctx:  context.WithoutCancel(pctx),
		Echo: echo.New(),
		Lc:   lc,
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
	if agw.Lc == nil {
		agw.Logger = log.StandardLogger()
		agw.Lc = &LogConfig{
			BodyBufferSize: 2000,
		}
		return nil
	}

	agw.Logger = log.New()

	lc := agw.Lc
	if lc.Level == "" {
		lc.Level = "info" //by default, apply info
	}

	level, err := logrus.ParseLevel(lc.Level)
	if err != nil {
		return err
	}

	switch lc.Output {
	case "stdout":
		agw.Logger.SetOutput(os.Stdout)
	case "stderr":
		agw.Logger.SetOutput(os.Stderr)
	case "":
		agw.Logger.SetOutput(os.Stdout)
	default:
		agw.Logger.SetOutput(&lumberjackx.Logger{
			Ctx:        context.WithoutCancel(agw.Ctx),
			Filename:   lc.Output,
			MaxSize:    lc.Size,    // megabytes
			MaxBackups: lc.BackNum, //file number
			MaxAge:     lc.AgeDays, //days
			Compress:   true,       // disabled by default
			LocalTime:  true,
		})
	}

	agw.Logger.SetLevel(level)
	agw.Logger.SetFormatter(lc.Formatter)

	if agw.Lc.BodyBufferSize == 0 {
		agw.Lc.BodyBufferSize = 2000
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
	format := "${time_rfc3339} ${status} ${method} ${latency_human} ${host} ${remote_ip} ${bytes_in} ${bytes_out} ${uri} ${id} ${error}\n"
	e.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		Handler: func(c echo.Context, reqBody []byte, resBody []byte) {
			lq := int(math.Min(float64(len(reqBody)), float64(agw.Lc.BodyBufferSize)))
			lp := int(math.Min(float64(len(resBody)), float64(agw.Lc.BodyBufferSize)))

			contentType := c.Response().Header().Get(echo.HeaderContentType)

			if isPrintableTextContent(contentType) || len(resBody) == 0 {
				log.Infof("%v, reqBody[%v]:{%v}, resBody[%v]:{%v}", c.Request().URL.String(), len(reqBody), string(reqBody[:lq]), len(resBody), string(resBody[:lp]))
			} else {
				log.Infof("%v, reqBody[%v]:{%v}, resBody[%v]:[Non-printable ContentType:%v]", c.Request().URL.String(), len(reqBody), string(reqBody[:lq]), len(resBody), contentType)
			}

			//accessLogger.Infof("%v, reqBody[%v]:{%v}, resBody[%v]:{%v}", c.Request().URL.String(), len(reqBody), string(reqBody[:lq]), len(resBody), string(resBody[:lp]))
		},
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: viperx.GetString("sys.accessFormat", format),
		//Output: accessLogger.Out,
		Output: log.StandardLogger().Out,
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
