package memkv

import "github.com/madlabx/pkgx/cachestore"

var _ cachestore.Record = &RefreshTokenMock{}

// RefreshTokenMock 作为Record接口的一个简单实现，用于测试
type RefreshTokenMock struct {
	IKey      string `gorm:"column:key;primaryKey;uniqueIndex"` // 主键，字段大小，不允许为空
	IValue    string `gorm:"column:value"`                      // 令牌值，字段大小，不允许为空
	IName     string `gorm:"column:name"`                       // 令牌名称，字段大小，默认值
	IExpireAt int64  `gorm:"column:expire_at"`                  // 过期时间，不允许为空，使用当前时间戳作为默认值
}

func (r *RefreshTokenMock) Clone() cachestore.Record {
	//TODO implement me
	n := *r
	return &n
}

func (r *RefreshTokenMock) Unmarshal(s string) error {
	r.IValue = s
	return nil
}

func (r *RefreshTokenMock) SetValue(v string)  { r.IValue = v }
func (r *RefreshTokenMock) GetKey() string     { return r.IKey }
func (r *RefreshTokenMock) GetValue() string   { return r.IValue }
func (r *RefreshTokenMock) GetExpireAt() int64 { return r.IExpireAt }
func (r *RefreshTokenMock) SetExpireAt(expireInSec int64) {
	r.IExpireAt = expireInSec
}

func (r *RefreshTokenMock) TableName() string { return "refresh_token_mock" }
