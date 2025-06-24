package rate

import (
	"sync"

	"github.com/madlabx/pkgx/errors"
)

type TpsOption func(*TpsLimiter)

// options of bbr limiter.
type tpsOptions struct {
	tag   string
	paral int
}

// WithWindow with window size.
func WithTag(tag string) TpsOption {
	return func(o *TpsLimiter) {
		o.tag = tag
	}
}

func WithParal(paral int) TpsOption {
	return func(o *TpsLimiter) {
		o.paral = paral
	}
}

type TpsLimiter struct {
	tpsOptions
	sem semaphore.Semaphore
}

func (tl *TpsLimiter) TryAcquire(n int) bool {
	return tl.sem.TryAcquire(n)
}

func (tl *TpsLimiter) Release(n int) int {
	return tl.sem.Release(n)
}

var tpsMap sync.Map

func GetTpsLimiter(opts ...TpsOption) (*TpsLimiter, error) {

	opt := TpsLimiter{}
	for _, o := range opts {
		o(&opt)
	}

	if opt.tag == "" {
		return nil, errors.New("tag is empty")
	}

	opt.sem = semaphore.New(opt.paral)

	tl, _ := tpsMap.LoadOrStore(opt.tag, &opt)
	return tl.(*TpsLimiter), nil
}
