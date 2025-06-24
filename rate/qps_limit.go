package rate

import (
	"sync"

	"github.com/madlabx/pkgx/log"
	"golang.org/x/time/rate"
)

type QpsOption func(*QpsLimiter)

// options of bbr limiter.
type qpsOptions struct {
	qps   float64
	burst int
}

type LimitLevel int

const (
	EnumLevelGlobal LimitLevel = iota
	EnumLevelUser
	EnumLevelClientId
	EnumLevelClientIp
)

// WithWindow with window size.
func QpsLimitOpt(level LimitLevel, qps float64, burst int) QpsOption {
	return func(o *QpsLimiter) {
		switch level {
		case EnumLevelGlobal:
			o.global = &qpsOptions{qps, burst}
		case EnumLevelUser:
			o.user = &qpsOptions{qps, burst}
		case EnumLevelClientId:
			o.deviceId = &qpsOptions{qps, burst}
		case EnumLevelClientIp:
			o.deviceIp = &qpsOptions{qps, burst}
		}
	}
}

type QpsLimiterItem struct {
	level LimitLevel
	rate  int
}

type QpsLimiter struct {
	global   *qpsOptions
	user     *qpsOptions
	deviceId *qpsOptions
	deviceIp *qpsOptions

	globalMutex     sync.Mutex
	globalLimiter   *rate.Limiter
	userMutex       sync.Mutex
	userLimiter     map[string]*rate.Limiter
	deviceIdMutex   sync.Mutex
	deviceIdLimiter map[string]*rate.Limiter
	deviceIpMutex   sync.Mutex
	deviceIpLimiter map[string]*rate.Limiter

	okMutex sync.RWMutex
	ok      bool
}

// Server ratelimiter middleware
func NewQpsLimiter(opts ...QpsOption) *QpsLimiter {

	opt := &QpsLimiter{ok: true}
	for _, o := range opts {
		o(opt)
	}

	return opt
}

func (l *QpsLimiter) OK() bool {
	ok := l.ok
	l.ok = true
	return ok
}
func (l *QpsLimiter) GetOk() bool {
	l.okMutex.RLock()
	defer l.okMutex.RUnlock()
	return l.ok
}
func (l *QpsLimiter) SetOk(ok bool) {
	l.okMutex.Lock()
	defer l.okMutex.Unlock()
	l.ok = ok
}
func (l *QpsLimiter) GlobalAllow() *QpsLimiter {
	if !l.GetOk() || l.global == nil {
		return l
	}

	l.globalMutex.Lock()
	if l.globalLimiter == nil {
		l.globalLimiter = rate.NewLimiter(rate.Limit(l.global.qps), l.global.burst)
	}
	l.SetOk(l.globalLimiter.Allow())

	log.Errorf("globalLimiter:%#v", l.globalLimiter)
	return l
}

func (l *QpsLimiter) UserAllow(user string) *QpsLimiter {
	log.Errorf("user:%v", user)
	if !l.GetOk() || user == "" || l.user == nil {
		return l
	}

	ul, ok := l.userLimiter[user]
	if ok {
		log.Errorf("userLimiter:%#v", ul)
		l.SetOk(ul.Allow())
	} else {
		l.userMutex.Lock()
		if l.userLimiter == nil {
			l.userLimiter = make(map[string]*rate.Limiter)
		}
		ul = rate.NewLimiter(rate.Limit(l.user.qps), l.user.burst)
		l.userLimiter[user] = ul
		log.Errorf("userLimiter:%#v", ul)
		l.userMutex.Unlock()

		l.SetOk(l.userLimiter[user].Allow())
	}

	return l
}

func (l *QpsLimiter) ClientIdAllow(deviceId string) *QpsLimiter {
	if !l.GetOk() || deviceId == "" || l.deviceId == nil {
		return l
	}

	dl, ok := l.deviceIdLimiter[deviceId]
	if ok {
		l.SetOk(dl.Allow())
	} else {
		l.deviceIdMutex.Lock()
		if l.deviceIdLimiter == nil {
			l.deviceIdLimiter = make(map[string]*rate.Limiter)
		}
		l.deviceIdLimiter[deviceId] = rate.NewLimiter(rate.Limit(l.deviceId.qps), l.deviceId.burst)
		l.deviceIdMutex.Unlock()

		l.SetOk(l.deviceIdLimiter[deviceId].Allow())
	}

	return l
}

func (l *QpsLimiter) ClientIpAllow(deviceIp string) *QpsLimiter {
	if !l.GetOk() || deviceIp == "" || l.deviceIp == nil {
		return l
	}
	dl, ok := l.deviceIpLimiter[deviceIp]
	if ok {
		l.SetOk(dl.Allow())
	} else {
		l.deviceIpMutex.Lock()
		if l.deviceIpLimiter == nil {
			l.deviceIpLimiter = make(map[string]*rate.Limiter)
		}
		l.deviceIpLimiter[deviceIp] = rate.NewLimiter(rate.Limit(l.deviceIp.qps), l.deviceIp.burst)
		l.deviceIpMutex.Unlock()

		l.SetOk(l.deviceIpLimiter[deviceIp].Allow())
	}

	return l
}
