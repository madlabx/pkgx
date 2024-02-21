package apigateway

import (
	"context"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/madlabx/pkgx/log"
	"github.com/madlabx/pkgx/lumberjackx"
	"github.com/madlabx/pkgx/viperx"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

type LogConfig struct {
	logOutput  string
	level      string
	logSize    int
	logBackNum int
	logAgeDays int
}

type ApiGateway struct {
	echo   *echo.Echo
	logger *logrus.Logger
}

func New(ctx context.Context, logConfig LogConfig) (*ApiGateway, error) {
	agw := &ApiGateway{
		echo: echo.New(),
	}
	if err := agw.initAccessLog(ctx, logConfig); err != nil {
		return nil, err
	}

	configEcho(agw.echo)
	return agw, nil
}

func (agw *ApiGateway) Run(ip, port string) error {
	showEcho(agw.echo)
	return startEcho(agw.echo, fmt.Sprintf("%s:%s", ip, port))
}

func (agw *ApiGateway) Stop() {
	shutdownEcho(agw.echo)
}

func (agw *ApiGateway) GetEcho() *echo.Echo {
	return agw.echo
}

func (agw *ApiGateway) initAccessLog(ctx context.Context, lc LogConfig) error {
	if agw.logger == nil {
		agw.logger = log.New()
	}

	level, err := logrus.ParseLevel(lc.level)
	if err != nil {
		return err
	}

	switch lc.logOutput {
	case "stdout":
		agw.logger.SetOutput(os.Stdout)
	case "stderr":
		agw.logger.SetOutput(os.Stderr)
	case "":
		agw.logger.SetOutput(os.Stdout)
	default:
		agw.logger.SetOutput(&lumberjackx.Logger{
			Ctx:        ctx,
			Filename:   lc.logOutput,
			MaxSize:    lc.logSize,    // megabytes
			MaxBackups: lc.logBackNum, //file number
			MaxAge:     lc.logAgeDays, //days
			Compress:   true,          // disabled by default
			LocalTime:  true,
		})
	}

	agw.logger.SetLevel(level)

	agw.logger.SetFormatter(&log.TextFormatter{QuoteEmptyFields: true})

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

func configEcho(e *echo.Echo) {
	// Tags to construct the logger format.
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
	format := "${time_rfc3339} ${status} ${method} ${latency_human} ${host} ${remote_ip} ${bytes_in} ${bytes_out} ${uri} ${id} ${error}\n"
	e.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		Handler: func(c echo.Context, reqBody []byte, resBody []byte) {
			lq := int(math.Min(float64(len(reqBody)), 2000))
			lp := int(math.Min(float64(len(resBody)), 2000))

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
		//AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	//TODO 检查是否可以恢复。不注释回无法下载css
	//e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
	//	return func(c echo.Context) error {
	//		c.Response().Header().Set("Content-Security-Policy", `default-src 'self'; style-src 'unsafe-inline';`)
	//		return next(c)
	//	}
	//})
}

func startEcho(e *echo.Echo, addr string) error {
	err := e.Start(addr)
	if err != nil {
		log.Errorf("Failed to bind address: %s, err[%v]", addr, err)
		return err
	}
	log.Infof("Start service listen on: %s", addr)
	return nil
}

func shutdownEcho(e *echo.Echo) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := e.Shutdown(ctx)
	if err != nil {
		log.Errorf("Failed to close echo: %v", e)
	}
	log.Infof("Close service: %v", e)
}

func showEcho(e *echo.Echo) {

	routes := make([]struct {
		m string
		p string
	}, len(e.Routes()))
	for i, r := range e.Routes() {
		routes[i].m = r.Method
		routes[i].p = r.Path
	}
	sort.Slice(routes, func(i, j int) bool { return routes[i].p < routes[j].p })

	for _, r := range routes {
		log.Infof("%s %s", r.m, r.p)
	}
}
