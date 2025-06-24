package redis

import (
	"context"
	"time"

	"github.com/madlabx/pkgx/cachestore"
	"github.com/madlabx/pkgx/errcode"
	"github.com/madlabx/pkgx/errors"
)

type Config struct {
	Addr     string
	Password string
}

type Client struct {
	ctx   context.Context
	rc    *redis.Client
	query string
}

func NewTestClient(pCtx context.Context, dbc *redis.Client) *Client {
	return &Client{
		ctx: context.WithoutCancel(pCtx),
		rc:  dbc,
	}
}

func NewClient(pCtx context.Context, conf Config) (*Client, error) {
	ctx := context.WithoutCancel(pCtx)
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,     // Redis服务器地址
		Password: conf.Password, // 密码，没有则留空
		DB:       0,             // 选择哪个数据库，0是默认
	},
	)

	// 测试连接
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	//
	//// 设置一个键值对
	//err = rdb.Set(ctx, "key", "value", 0).Err()
	//if err != nil {
	//	return nil, errors.Wrap(err)
	//}

	return &Client{
		ctx: ctx,
		rc:  rdb,
	}, nil
}

type RecordUnmarshaler interface {
	Unmarshal(string) error
}

func (rc *Client) GetRaw() *redis.Client {
	return rc.rc
}

func (rc *Client) ResetClient(realDbClient *redis.Client) {
	rc.rc = realDbClient
}

func (rc *Client) Set(nr cachestore.Record, expireAfterInSec int64) error {
	return rc.rc.Set(rc.ctx, cachestore.UniqCacheKey(nr), nr.GetValue(), time.Second*time.Duration(expireAfterInSec)).Err()
}

func (rc *Client) Del(nr cachestore.Record) error {
	return rc.rc.Del(rc.ctx, cachestore.UniqCacheKey(nr)).Err()
}

func (rc *Client) Where(query string) *Client {
	rc.query = query
	return rc
}

// TODO remove reflect, return (Record, error)
// TODO support to define key with tag, then no need to call Where
func (rc *Client) Get(nr cachestore.Record) error {
	content, err := rc.rc.Get(rc.ctx, cachestore.UniqCacheKey(nr)).Result()
	if errors.Is(err, redis.Nil) {
		return errcode.ErrObjectNotExist()
	} else if err != nil {
		return errors.Wrap(err)
	}

	return nr.Unmarshal(content)
}

func (rc *Client) Update(nr cachestore.Record) error {
	//need pipeline
	ttl, err := rc.rc.TTL(rc.ctx, cachestore.UniqCacheKey(nr)).Result()
	if errors.Is(err, redis.Nil) {
		return errcode.ErrObjectNotExist()
	} else if err != nil {
		return errors.Wrap(err)
	}

	if ttl == -2 {
		return errcode.ErrObjectNotExist()
	}

	if ttl == -1 {
		ttl = 0
	}

	return rc.rc.Set(rc.ctx, cachestore.UniqCacheKey(nr), nr.GetValue(), ttl).Err()
}

func List[T cachestore.Record](rc *Client, nr T) ([]T, error) {

	var cursor uint64
	pattern := nr.TableName() + "*"
	var keys []string
	var records []T

	for {
		var err error
		var ckeys []string
		ckeys, cursor, err = rc.rc.Scan(rc.ctx, cursor, pattern, 0).Result()
		if err != nil {
			return nil, errors.Wrap(err)
		}

		if len(ckeys) > 0 {
			keys = append(keys, ckeys...)
		}
		if cursor == 0 {
			break
		}

	}
	// 使用Pipeline批量获取key的值
	//var pipe redis.Pipeliner
	pipe := rc.GetRaw().Pipeline()
	for _, key := range keys {
		pipe.Get(rc.ctx, key)
	}

	results, err := pipe.Exec(rc.ctx)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// 打印结果
	for _, result := range results {
		if result.Err() != nil {
			return nil, errors.Wrap(err)
		}
		tmp := nr.Clone()
		err = tmp.Unmarshal(result.(*redis.StringCmd).Val())
		if err != nil {
			return nil, errors.Wrap(err)
		}

		records = append(records, tmp.(T))
	}

	return records, nil
}

func (rc *Client) CompareAndDelete(nr cachestore.Record, compareValueFunc func(cachestore.Record, cachestore.Record) bool) (bool, error) {
	var result bool
	//_ := rc.rc.TxPipeline()
	unitKey := cachestore.UniqCacheKey(nr)
	errExternal := rc.rc.Watch(rc.ctx, func(tx *redis.Tx) error {
		nrInRedis := nr.Clone()
		content, err := tx.Get(rc.ctx, cachestore.UniqCacheKey(nr)).Result()
		if errors.Is(err, redis.Nil) {
			return errcode.ErrObjectNotExist()
		} else if err != nil {
			return errors.Wrap(err)
		}

		if err = nrInRedis.Unmarshal(content); err != nil {
			return errors.Wrap(err)
		}

		if compareValueFunc(nr, nrInRedis) {
			result = true
			_, err = tx.Del(rc.ctx, unitKey).Result()
			if err != nil {
				return errors.Wrap(err)
			}
		} else {
			result = false
		}
		return errors.Wrap(err)
	}, unitKey)

	if errors.Is(errExternal, redis.Nil) {
		return false, errcode.ErrObjectNotExist()
	} else if errExternal != nil {
		return false, errors.Wrap(errExternal)
	}
	return result, errExternal
}
