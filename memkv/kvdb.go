package memkv

type KvDbClientIf interface {
	Set(records any) error
	Get(filterAndDest any) error
	ListWithKeyPrefix(dest any, filter any, keyFieldName, keyPrefix string) error
	DeleteExpired(filter any) error
}
