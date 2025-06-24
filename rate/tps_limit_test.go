package rate

import (
	"testing"
	"time"

	"github.com/madlabx/pkgx/errcode"
	"github.com/madlabx/pkgx/log"
)

func withTpsLimiter(fn func() error, options ...TpsOption) func() error {
	return func() error {
		tl, err := GetTpsLimiter(options...)
		if err != nil {
			return err
		}

		if !tl.TryAcquire(1) {
			return errcode.ErrTooManyRequests()
		}
		defer tl.Release(1)
		return fn()
	}
}

func sambaEnable() error {
	log.Errorf("Enter sambeEnable")
	defer log.Errorf("Exit sambe Enable")
	time.Sleep(time.Second * 1)
	return nil
}
func sambaUpdate() error {
	log.Errorf("Enter sambaUpdate")
	defer log.Errorf("Exit sambaUpdate")
	time.Sleep(time.Second * 1)
	return nil
}

func TestStack(t *testing.T) {
	f1 := withTpsLimiter(sambaEnable,
		WithTag("sambaWrite"),
		WithParal(1))

	f2 := withTpsLimiter(sambaUpdate,
		WithTag("sambaWrite"),
		WithParal(1))

	go func() {
		if err := f1(); err != nil {
			log.Errorf("Failed to run f1, err:%v", err)
		}
	}()
	go func() {
		if err := f1(); err != nil {
			log.Errorf("Failed to run f1 for second, err:%v", err)
		}
	}()
	go func() {
		if err := f1(); err != nil {
			log.Errorf("Failed to run f1 for second, err:%v", err)
		}
	}()
	go func() {
		if err := f1(); err != nil {
			log.Errorf("Failed to run f1 for second, err:%v", err)
		}
	}()
	go func() {
		if err := f1(); err != nil {
			log.Errorf("Failed to run f1 for second, err:%v", err)
		}
	}()
	go func() {
		if err := f1(); err != nil {
			log.Errorf("Failed to run f1 for second, err:%v", err)
		}
	}()
	go func() {
		if err := f2(); err != nil {
			log.Errorf("Failed to run f2, err:%v", err)
		}
	}()
	go func() {
		if err := f2(); err != nil {
			log.Errorf("Failed to run f2, err:%v", err)
		}
	}()

	go func() {
		if err := f2(); err != nil {
			log.Errorf("Failed to run f2, err:%v", err)
		}
	}()

	time.Sleep(3 * time.Second)
}
