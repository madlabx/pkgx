package log

import (
	"errors"
	"testing"
)

func Test_Errorf(t *testing.T) {
	StandardLogger().ReportCaller = true
	Errorf("StandardLogger.out:%p", StandardLogger().Formatter)
	newLogger := WithField("logId", "field")
	Errorf("xx")
	Error("bb")
	newLogger.Errorf("test2")
	Errorf("StandardLogger.out:%p", newLogger.Logger.Formatter)
	newLogger.Errorf("test")
	IgnoreErrf(errors.New("wef"), "sd")
}
