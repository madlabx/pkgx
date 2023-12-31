package lumberjackx_test

import (
	"github.com/madlabx/pkgx/lumberjackx"
	"log"
)

// To use xlumberjack with the standard library's log package, just pass it into
// the SetOutput function when your application starts.
func Example() {
	log.SetOutput(&lumberjackx.Logger{
		Filename:   "/var/log/myapp/foo.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   // days
		Compress:   true, // disabled by default
	})
}
