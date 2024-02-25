package errorx

import (
	"github.com/madlabx/pkgx/log"
)

func CheckFatalError(err error) {
	if err != nil {
		log.Panicf("FatalError:%v", err)
	}
}
