package randx

import (
	"math/rand"
	"sync"
	"time"
)

type Rand struct {
	*rand.Rand
}

var (
	_pool = sync.Pool{
		New: func() interface{} {
			return &Rand{rand.New(rand.NewSource(time.Now().UnixNano()))}
		},
	}
)

func GetRand() *Rand {
	return _pool.Get().(*Rand)
}

func NewRand() *Rand {
	return GetRand()
}

func (r *Rand) Release() {
	_pool.Put(r)
}

// RandRange 范围随机 [min, max]
func (r *Rand) RandRange(min int, max int) int {
	if max <= min {
		if max == min {
			return min
		}
		return 0
	}
	return r.Intn(max-min+1) + min
}

// RandRangeInt32 范围随机 [min, max]
func (r *Rand) RandRangeInt32(min int32, max int32) int {
	return r.RandRange(int(min), int(max))
}

// RandRangeInt64 范围随机 [min, max]
func (r *Rand) RandRangeInt64(min int64, max int64) int64 {
	if max <= min {
		if max == min {
			return min
		}
		return 0
	}
	return r.Int63n(max-min+1) + min
}

func (r *Rand) Bool() bool {
	return r.Intn(2) == 0
}
