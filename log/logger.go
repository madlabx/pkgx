package log

import (
	"context"
	"io"
	"os"

	"github.com/madlabx/pkgx/lumberjackx"
	"github.com/sirupsen/logrus"
)

type FileConfig struct {
	Filename string `vx_default:"main"`

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `vx_default:"10"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `vx_default:"7"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `vx_default:"5"`

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	LocalTime bool `vx_default:"true"`

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `vx_default:"true"`
}

type Logger struct {
	*logrus.Logger
}

func New() *Logger {
	return &Logger{logrus.New()}
}

func NewLogger(pCtx context.Context, cfg FileConfig) *Logger {
	lg := New()
	SetLoggerOutput(lg, pCtx, cfg)
	return lg
}

//type LoggerIf interface {
//	SetOutput(output io.Writer)
//}

func SetLoggerLevel(lo *Logger, level string) error {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	lo.SetLevel(l)
	return nil
}

func SetLoggerOutput(lo *Logger, pCtx context.Context, cfg FileConfig) *Logger {
	if lo == nil {
		lo = New()
	}

	switch cfg.Filename {
	case "main":
		lo.SetOutput(StandardLogger().Out)
	case "stdout":
		lo.SetOutput(os.Stdout)
	case "stderr":
		lo.SetOutput(os.Stderr)
	case "discard":
		lo.SetOutput(io.Discard)
	case "":
		lo.SetOutput(os.Stdout)
	default:
		lo.SetOutput(&lumberjackx.Logger{
			Ctx:        context.WithoutCancel(pCtx),
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,    // megabytes
			MaxBackups: cfg.MaxBackups, //file number
			MaxAge:     cfg.MaxAge,     //days
			Compress:   cfg.Compress,   // disabled by default
			LocalTime:  cfg.LocalTime,
		})
	}
	return lo
}

func SetLoggerFormatter(lo *Logger, formatter logrus.Formatter) {
	lo.SetFormatter(formatter)
}

func StandardLogger() *Logger {
	return &Logger{logrus.StandardLogger()}
}
