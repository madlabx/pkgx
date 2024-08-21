package httpx

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	labstacklog "github.com/labstack/gommon/log"
	"github.com/madlabx/pkgx/log"
	"github.com/sirupsen/logrus"
)

const (
	DefaultBodyBufferSize = 4096
)

type LogConfig struct {
	LogFile        log.FileConfig
	Level          string          `vx_default:"info"`
	Timing         AccessLogTiming `vx_default:"both"`
	BodyBufferSize int64           `vx_default:"4096"`
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
	// - header_in:<NAME>
	// - header_out:<NAME>
	// - query:<NAME>
	// - form:<NAME>
	// - body_in (request body)
	// - body_out (response body)
	//ContentFormatBefore string `vx_default:"${time_custom} BEF ${method} ${uri} ${host} ${remote_ip} ${bytes_in}"`
	ContentFormatBefore string
	//ContentFormatAfter  string `vx_default:"${time_custom} AFT ${status} ${method} ${latency_human} ${uri} ${host} ${remote_ip} ${bytes_in} ${bytes_out} ${error}"`
	ContentFormatAfter string
}

type ApiGateway struct {
	ctx context.Context
	*echo.Echo
	Logger            *log.Logger
	LogConf           *LogConfig
	EntryFormat       logrus.Formatter
	loggerSkipper     middleware.Skipper
	bodyLoggerSkipper middleware.Skipper
}

func NewApiGateway(pCtx context.Context, lc *LogConfig, logFormat logrus.Formatter) (*ApiGateway, error) {
	agw := &ApiGateway{
		ctx:         context.WithoutCancel(pCtx),
		Echo:        echo.New(),
		LogConf:     lc,
		EntryFormat: logFormat,
	}

	//if lc == nil, log to log.StandardLogger
	if err := agw.initAccessLog(); err != nil {
		return nil, err
	}

	return agw, nil
}

func (agw *ApiGateway) SetLoggerSkipper(s middleware.Skipper) {
	agw.loggerSkipper = s
}

func (agw *ApiGateway) SetBodyLoggerSkipper(s middleware.Skipper) {
	agw.bodyLoggerSkipper = s
}

func (agw *ApiGateway) Run(ip, port string) error {
	agw.configEcho()
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
		agw.Logger = log.NewLogger(agw.ctx, agw.LogConf.LogFile)
	}

	level, err := logrus.ParseLevel(agw.LogConf.Level)
	if err != nil {
		return err
	}
	agw.Logger.SetLevel(level)

	// Set body format
	if agw.EntryFormat == nil {
		agw.EntryFormat = &log.TextFormatter{QuoteEmptyFields: true}
	}
	agw.Logger.SetFormatter(agw.EntryFormat)

	return nil
}

func (agw *ApiGateway) configEcho() {
	var (
		e = agw.Echo
	)

	e.Logger.SetOutput(agw.Logger.Out)
	level, _ := logrus.ParseLevel(agw.LogConf.Level)
	switch {
	case level <= logrus.ErrorLevel:
		e.Logger.SetLevel(labstacklog.ERROR)
	case level == logrus.WarnLevel:
		e.Logger.SetLevel(labstacklog.WARN)
	default:
		e.Logger.SetLevel(labstacklog.INFO)
	}

	bodyFilter := func(c echo.Context) bool {
		//文件上传下载不要打印
		//return c.Request().Method == http.MethodPost
		//if strings.Contains(c.Request().URL.Path, "/v1/file_service/obj/download_file") ||
		//	strings.Contains(c.Request().URL.Path, "/v1/file_service/obj/upload_file") {
		//	return true
		//}
		return true
	}
	if agw.bodyLoggerSkipper != nil {
		bodyFilter = agw.bodyLoggerSkipper
	}
	e.Use(LoggerWithConfig(LoggerConfig{
		OutBodyFilter:    bodyFilter,
		FormatAfter:      agw.LogConf.ContentFormatAfter,
		FormatBefore:     agw.LogConf.ContentFormatBefore,
		CustomTimeFormat: "2006/01/02 15:04:05.000",
		Output:           agw.Logger.Out,
		bodyBufferSize:   agw.LogConf.BodyBufferSize,
		Timing:           agw.LogConf.Timing,
		Skipper:          agw.loggerSkipper,
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
	ctx, cancel := context.WithTimeout(agw.ctx, 5*time.Second)
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
