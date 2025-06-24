package randx

func Int63() int64 { r := GetRand(); defer r.Release(); return r.Int63() }

func Uint32() uint32 { r := GetRand(); defer r.Release(); return r.Uint32() }

func Uint64() uint64 { r := GetRand(); defer r.Release(); return r.Uint64() }

func Int31() int32 { r := GetRand(); defer r.Release(); return r.Int31() }

func Int() int { r := GetRand(); defer r.Release(); return r.Int() }

func Int63n(n int64) int64 { r := GetRand(); defer r.Release(); return r.Int63n(n) }

func Int31n(n int32) int32 { r := GetRand(); defer r.Release(); return r.Int31n(n) }

func Intn(n int) int { r := GetRand(); defer r.Release(); return r.Intn(n) }

func Float64() float64 { r := GetRand(); defer r.Release(); return r.Float64() }

func Float32() float32 { r := GetRand(); defer r.Release(); return r.Float32() }

func Perm(n int) []int { r := GetRand(); defer r.Release(); return r.Perm(n) }

func Shuffle(n int, swap func(i, j int)) { r := GetRand(); defer r.Release(); r.Shuffle(n, swap) }

func Read(p []byte) (n int, err error) { r := GetRand(); defer r.Release(); return r.Read(p) }

func NormFloat64() float64 { r := GetRand(); defer r.Release(); return r.NormFloat64() }

func ExpFloat64() float64 { r := GetRand(); defer r.Release(); return r.ExpFloat64() }

// RandRange 范围随机 [min, max]
func RandRange(min int, max int) int {
	r := GetRand()
	defer r.Release()
	return r.RandRange(min, max)
}

// RandRangeInt32 范围随机 [min, max]
func RandRangeInt32(min int32, max int32) int {
	r := GetRand()
	defer r.Release()
	return r.RandRangeInt32(min, max)
}

// RandRangeInt64 范围随机 [min, max]
func RandRangeInt64(min int64, max int64) int64 {
	r := GetRand()
	defer r.Release()
	return r.RandRangeInt64(min, max)
}

func Bool() bool {
	r := GetRand()
	defer r.Release()
	return r.Bool()
}
