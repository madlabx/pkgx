package log

import (
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

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

func IgnoreErrf(err error, format string, args ...interface{}) {
	if err != nil {
		logrus.Errorf("Ignore err during %v, err:%v", fmt.Sprintf(format, args...), err)
	}
}
