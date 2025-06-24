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
	"github.com/madlabx/pkgx/memkv"
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
type HandlerOnIdempotentErrFunc func(c echo.Context, requestId string) error

type ApiGateway struct {
	ctx context.Context
	*echo.Echo
	addr                     string
	port                     string
	name                     string
	Logger                   *log.Logger
	LogConf                  *LogConfig
	EntryFormat              logrus.Formatter
	loggerSkipper            middleware.Skipper
	bodyLoggerSkipper        middleware.Skipper
	isIdempotent             bool
	idempotentKeyExpiry      int64
	idempotentNameHeader     string
	idempotentNameQueryParam string
	idempotentKeyCache       *memkv.Cache
	onIdempotenceCheckError  HandlerOnIdempotentErrFunc
}

func NewApiGateway(pCtx context.Context, addr, port, name string, lc *LogConfig, logFormat logrus.Formatter) (*ApiGateway, error) {
	agw := &ApiGateway{
		addr:         addr,
		port:         port,
		name:         name,
		ctx:          context.WithoutCancel(pCtx),
		Echo:         echo.New(),
		LogConf:      lc,
		EntryFormat:  logFormat,
		isIdempotent: false,
	}

	//if lc == nil, log to log.StandardLogger
	if err := agw.initAccessLog(); err != nil {
		return nil, err
	}

	return agw, nil
}

// not support dynamic change
func (agw *ApiGateway) EnableIdempotentCheck(idempotentNameInHeader, idempotentNameInQueryParam string, expiryInSec int64, fn HandlerOnIdempotentErrFunc) {
	agw.isIdempotent = true
	agw.idempotentNameHeader = idempotentNameInHeader
	agw.idempotentNameQueryParam = idempotentNameInQueryParam
	agw.idempotentKeyExpiry = expiryInSec
	agw.onIdempotenceCheckError = fn
}

func (agw *ApiGateway) SetLoggerSkipper(s middleware.Skipper) {
	agw.loggerSkipper = s
}

func (agw *ApiGateway) SetBodyLoggerSkipper(s middleware.Skipper) {
	agw.bodyLoggerSkipper = s
}

func (agw *ApiGateway) Name() string {
	return agw.name
}

func (agw *ApiGateway) enableIdempotence() {
	agw.idempotentKeyCache = memkv.NewCache(agw.ctx, nil, memkv.CacheConf{})
}

func (agw *ApiGateway) Run() error {
	if agw.isIdempotent {
		agw.enableIdempotence()
	}

	agw.configEcho()
	return agw.startEcho(fmt.Sprintf("%s:%s", agw.addr, agw.port))
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

func (agw *ApiGateway) idempotenceCheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		rid := c.Request().Header.Get(agw.idempotentNameHeader)
		if rid == "" {
			rid = c.QueryParam(agw.idempotentNameQueryParam)
		}

		if rid == "" {
			if err := agw.onIdempotenceCheckError(c, rid); err != nil {
				return err
			}

			return next(c)
		}

		isExist, err := agw.idempotentKeyCache.CreateOrUpdate(&idempotentRecord{Key: rid}, agw.idempotentKeyExpiry)
		if err != nil {
			return err
		}

		if isExist {
			if err = agw.onIdempotenceCheckError(c, rid); err != nil {
				return err
			}
		}

		return next(c)
	}
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

	if agw.isIdempotent {
		e.Use(agw.idempotenceCheck)
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
