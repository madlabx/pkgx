package cachestore

type Record interface {
	GetKey() string
	GetValue() string
	Unmarshal(string) error
	SetExpireAt(int64)  //meaningless for redis
	GetExpireAt() int64 //meaningless for redis
	TableName() string
	Clone() Record
}

func UniqCacheKey(r Record) string {
	return r.TableName() + "_" + r.GetKey()
}

type ConsistentRecord interface {
	Record
	GetPrimaryName() string
}
