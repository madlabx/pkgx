package log

import (
	"context"
	"fmt"
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

func New() *logrus.Logger {
	return logrus.New()
}

func NewLogger(pCtx context.Context, cfg FileConfig) *logrus.Logger {
	lg := New()
	SetLoggerOutput(lg, pCtx, cfg)
	return lg
}

type LoggerIf interface {
	SetOutput(output io.Writer)
}

func SetLoggerLevel(lo *logrus.Logger, level string) error {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	lo.SetLevel(l)
	return nil
}

func SetLoggerOutput(lo *logrus.Logger, pCtx context.Context, cfg FileConfig) *logrus.Logger {
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

func SetLoggerFormatter(lo *logrus.Logger, formatter logrus.Formatter) {
	lo.SetFormatter(formatter)
}

func StandardLogger() *logrus.Logger {
	return logrus.StandardLogger()
}

func SetOutput(out io.Writer) {
	logrus.SetOutput(out)
}

func Set(out io.Writer) {
	logrus.SetOutput(out)
}

func SetLevelStr(level string) error {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(l)
	return nil
}

func SetLevel(level logrus.Level) {
	logrus.SetLevel(level)
}

func SetFormatter(formatter logrus.Formatter) {
	logrus.SetFormatter(formatter)
}

func WithError(err error) *logrus.Entry {
	return logrus.WithField(logrus.ErrorKey, err)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return logrus.WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return logrus.WithFields(fields)
}

func Debug(args ...interface{}) {
	logrus.Debug(args...)
}

func Print(args ...interface{}) {
	logrus.Print(args...)
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}

func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

func Warning(args ...interface{}) {
	logrus.Warning(args...)
}

func Error(args ...interface{}) {
	logrus.Error(args...)
}

func Panic(args ...interface{}) {
	logrus.Panic(args...)
}

func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

func Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

func Printf(format string, args ...interface{}) {
	logrus.Printf(format, args...)
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func Warningf(format string, args ...interface{}) {
	logrus.Warningf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func StdoutPrintf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func Panicf(format string, args ...interface{}) {
	logrus.Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

func Debugln(args ...interface{}) {
	logrus.Debugln(args...)
}

func Println(args ...interface{}) {
	logrus.Println(args...)
}

func Infoln(args ...interface{}) {
	logrus.Infoln(args...)
}

func Warnln(args ...interface{}) {
	logrus.Warnln(args...)
}

func Warningln(args ...interface{}) {
	logrus.Warningln(args...)
}

func Errorln(args ...interface{}) {
	logrus.Errorln(args...)
}

func Panicln(args ...interface{}) {
	logrus.Panicln(args...)
}

func Fatalln(args ...interface{}) {
	logrus.Fatalln(args...)
}

func Eventf(format string, args ...interface{}) {
	logrus.Infof("--EVENT-- "+format, args...)
}

func FatalIf(args ...interface{}) {
	logrus.Fatal(args...)
}
