package httpx

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/color"
	"github.com/valyala/fasttemplate"
)

type AccessLogTiming string

const (
	defaultBufSize = 4096

	AccessLogBeforeRun AccessLogTiming = "before"
	AccessLogAfterRun  AccessLogTiming = "after"
	AccessLogBoth      AccessLogTiming = "both"
)

type (
	Filter func(echo.Context) bool

	// LoggerConfig defines the config for Logger middleware.
	LoggerConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// OutBodyFilter defines a function to print body_out, false by default due to additional memory used
		OutBodyFilter middleware.Skipper

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
		// - header_in:<NAME>
		// - header_out:<NAME>
		// - query:<NAME>
		// - form:<NAME>
		// - body_in (request body)
		// - body_out (response body)   , should also define OutBodyFilter to log only necessary.
		//
		// Example "${remote_ip} ${status}"
		//
		// Optional. Default value DefaultLoggerConfig.Format.
		FormatAfter  string `yaml:"format_after"`
		FormatBefore string `yaml:"format_before"`
		Timing       AccessLogTiming

		// Optional. Default value DefaultLoggerConfig.CustomTimeFormat.
		CustomTimeFormat string `yaml:"custom_time_format"`

		// Output is a writer where logs in JSON format are written.
		// Optional. Default value os.Stdout.
		Output io.Writer

		templateAfter  *fasttemplate.Template
		templateBefore *fasttemplate.Template
		colorer        *color.Color
		pool           *sync.Pool
		bodyBufferSize int64
	}
)

var (
	// DefaultLoggerConfig is the default Logger middleware config.
	DefaultLoggerConfig = LoggerConfig{
		Skipper:       middleware.DefaultSkipper,
		OutBodyFilter: DefaultOutBodyFilter,
		FormatBefore: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`,"bytes_in":${bytes_in}`,
		FormatAfter: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}`,
		CustomTimeFormat: "2006-01-02 15:04:05.000",
		Output:           os.Stdout,
		colorer:          color.New(),
		bodyBufferSize:   defaultBufSize,
		Timing:           AccessLogBoth,
	}
	configAccess *LoggerConfig
	once         sync.Once
)

// DefaultOutBodyFilter returns false which processes the middleware.
func DefaultOutBodyFilter(echo.Context) bool {
	return false
}

//// Logger returns a middleware that logs HTTP requests.
//func Logger() echo.MiddlewareFunc {
//	return LoggerWithConfig(DefaultLoggerConfig)
//}

func LoggerWithConfig(config LoggerConfig) echo.MiddlewareFunc {
	once.Do(func() {
		configAccess = &config
	})

	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultLoggerConfig.Skipper
	}

	if config.FormatAfter == "" {
		config.FormatAfter = DefaultLoggerConfig.FormatAfter
	}

	if config.Output == nil {
		config.Output = DefaultLoggerConfig.Output
	}

	if config.Timing == "" {
		config.Timing = DefaultLoggerConfig.Timing
	}

	if config.Timing != AccessLogAfterRun {
		if config.FormatBefore == "" {
			config.FormatBefore = DefaultLoggerConfig.FormatBefore
		}
		config.templateBefore = fasttemplate.New(config.FormatBefore+"\n", "${", "}")
	}

	if config.Timing != AccessLogBeforeRun {
		if config.FormatAfter == "" {
			config.FormatAfter = DefaultLoggerConfig.FormatAfter
		}
		config.templateAfter = fasttemplate.New(config.FormatAfter+"\n", "${", "}")
	}
	config.colorer = color.New()
	config.colorer.SetOutput(config.Output)
	config.pool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}

	loggingRequestBody := func(c echo.Context, bytesIn int64) string {
		if bytesIn > 0 && bytesIn <= config.bodyBufferSize &&
			isPrintableTextContent(c.Request().Header.Get(echo.HeaderContentType)) {
			// Request
			var reqBody []byte
			if c.Request().Body != nil { // Read
				reqBody, _ = io.ReadAll(c.Request().Body)
			}
			c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Reset
			bytesIn = min(bytesIn, int64(len(reqBody)))
			return fmt.Sprintf("in[%v]:%v", bytesIn, string(reqBody[:bytesIn]))
		}
		return fmt.Sprintf("in[%v]", bytesIn)
	}

	loggingResponseBody := func(c echo.Context, doPrintBodyOut bool, bytesOut int64, respBody []byte) string {
		if doPrintBodyOut && bytesOut > 0 && bytesOut <= config.bodyBufferSize &&
			isPrintableTextContent(c.Response().Header().Get(echo.HeaderContentType)) {
			//skip "\n"
			bytesOut = min(bytesOut, int64(len(respBody)))
			bytesOut = max(0, bytesOut-1)
			return fmt.Sprintf("out[%v]:%v", bytesOut, string(respBody[:bytesOut]))
		}
		return fmt.Sprintf("out[%v]", bytesOut)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()
			start := time.Now()

			doPrintBodyOut := config.OutBodyFilter(c)
			respBody := newLimitBuffer(config.bodyBufferSize)
			if doPrintBodyOut {
				mw := io.MultiWriter(c.Response().Writer, respBody)
				writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
				c.Response().Writer = writer
			}

			loggingTemplate := func(buf *bytes.Buffer, tag string) (int, error) {
				switch tag {
				case "time_unix":
					return buf.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
				case "time_unix_nano":
					return buf.WriteString(strconv.FormatInt(time.Now().UnixNano(), 10))
				case "time_rfc3339":
					return buf.WriteString(time.Now().Format(time.RFC3339))
				case "time_rfc3339_nano":
					return buf.WriteString(time.Now().Format(time.RFC3339Nano))
				case "time_custom":
					return buf.WriteString(time.Now().Format(config.CustomTimeFormat))
				case "id":
					id := req.Header.Get(echo.HeaderXRequestID)
					if id == "" {
						id = res.Header().Get(echo.HeaderXRequestID)
					}
					return buf.WriteString(id)
				case "remote_ip":
					return buf.WriteString(c.RealIP())
				case "host":
					return buf.WriteString(req.Host)
				case "uri":
					return buf.WriteString(req.RequestURI)
				case "method":
					return buf.WriteString(req.Method)
				case "path":
					p := req.URL.Path
					if p == "" {
						p = "/"
					}
					return buf.WriteString(p)
				case "protocol":
					return buf.WriteString(req.Proto)
				case "referer":
					return buf.WriteString(req.Referer())
				case "user_agent":
					return buf.WriteString(req.UserAgent())

				case "bytes_in":
					cl := req.Header.Get(echo.HeaderContentLength)
					if cl == "" {
						cl = "0"
					}
					return buf.WriteString(cl)
				case "body_in":
					cl := req.Header.Get(echo.HeaderContentLength)
					if cl == "" {
						cl = "0"
					}
					bytesIn, _ := strconv.Atoi(cl)
					return buf.WriteString(loggingRequestBody(c, int64(bytesIn)))

				case "latency":
					l := time.Now().Sub(start)
					return buf.WriteString(strconv.FormatInt(int64(l), 10))
				case "latency_human":
					return buf.WriteString(time.Now().Sub(start).String())
				case "bytes_out":
					return buf.WriteString(strconv.FormatInt(res.Size, 10))
				case "body_out":
					return buf.WriteString(loggingResponseBody(c, doPrintBodyOut, res.Size, respBody.Bytes()))
				case "status":
					n := res.Status
					s := config.colorer.Green(n)
					switch {
					case n >= 500:
						s = config.colorer.Red(n)
					case n >= 400:
						s = config.colorer.Yellow(n)
					case n >= 300:
						s = config.colorer.Cyan(n)
					}
					return buf.WriteString(s)
				default:
					switch {
					case strings.HasPrefix(tag, "header_in:"):
						return buf.Write([]byte(c.Request().Header.Get(tag[11:])))
					case strings.HasPrefix(tag, "header_out:"):
						return buf.Write([]byte(c.Response().Header().Get(tag[12:])))
					case strings.HasPrefix(tag, "query:"):
						return buf.Write([]byte(c.QueryParam(tag[6:])))
					case strings.HasPrefix(tag, "form:"):
						return buf.Write([]byte(c.FormValue(tag[5:])))
					case strings.HasPrefix(tag, "cookie:"):
						cookie, err := c.Cookie(tag[7:])
						if err == nil {
							return buf.Write([]byte(cookie.Value))
						}
					}
				}
				return 0, nil
			}

			buf := config.pool.Get().(*bytes.Buffer)
			defer config.pool.Put(buf)

			if config.templateBefore != nil {
				//Log after run
				buf.Reset()
				if _, err = config.templateBefore.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
					return loggingTemplate(buf, tag)
				}); err != nil {
					return
				}
				if _, err = config.Output.Write(buf.Bytes()); err != nil {
					return
				}
			}

			if err = next(c); err != nil {
				c.Error(err)
			}

			if config.templateAfter == nil {
				return
			}

			//Log after run
			buf.Reset()
			if _, err = config.templateAfter.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
				return loggingTemplate(buf, tag)
			}); err != nil {
				return
			}

			_, err = config.Output.Write(buf.Bytes())

			return
		}
	}
}

func isPrintableTextContent(contentType string) bool {
	return strings.HasPrefix(contentType, echo.MIMEApplicationJSON)
}

type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *bodyDumpResponseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func newLimitBuffer(size int64) *limitBuffer {

	if size <= 0 {
		size = defaultBufSize
	}
	return &limitBuffer{
		buf: make([]byte, size),
	}
}

// limitBuffer only receive first n bytes, ignore the left, and not log rise any error
type limitBuffer struct {
	buf []byte
	n   int
}

// Write only receive first n bytes, ignore the left, and not log rise any error.
// Return with len(p) to fake the normal behavior
func (b *limitBuffer) Write(p []byte) (n int, err error) {
	toWrite := min(len(p), b.Available())
	_ = copy(b.buf[b.n:], p[:toWrite])
	b.n += toWrite
	return len(p), nil
}

func (b *limitBuffer) Available() int { return len(b.buf) - b.n }
func (b *limitBuffer) Bytes() []byte  { return b.buf[:b.n] }
