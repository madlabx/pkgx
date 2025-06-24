package httpx

import "github.com/madlabx/pkgx/cachestore"

var _ cachestore.Record = &idempotentRecord{}

type idempotentRecord struct {
	Key      string
	ExpireAt int64
}

func (i *idempotentRecord) GetKey() string {
	return i.Key
}

func (i *idempotentRecord) GetValue() string {
	return ""
}

func (i *idempotentRecord) Unmarshal(s string) error {
	i.Key = s
	return nil
}

func (i *idempotentRecord) SetExpireAt(i2 int64) {
	i.ExpireAt = i2
}

func (i *idempotentRecord) GetExpireAt() int64 {
	return i.ExpireAt
}

func (i *idempotentRecord) TableName() string {
	return "apigateway_idempotent_record"
}

func (i *idempotentRecord) Clone() cachestore.Record {
	return &idempotentRecord{
		Key: i.Key,
	}
}
