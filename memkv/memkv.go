package memkv

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/madlabx/pkgx/cachestore"
	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/log"
	"gorm.io/gorm"
)

//TODO, add wrapper数据结构模式

// Cache 结构，用于内存缓存
type Cache struct {
	ctx     context.Context
	items   *sync.Map
	db      KvDbClientIf
	records []any
	conf    CacheConf
}

type CacheConf struct {
	GcInterval int `vx_default:"10"` //in sec
}

// NewCache 创建一个新的缓存实例
func NewCache(pCtx context.Context, client KvDbClientIf, conf CacheConf, records ...any) *Cache {
	ctx, _ :=context.WithCancel(pCtx),
	cache := &Cache{
		ctx:     ctx,
		items:   new(sync.Map),
		db:      client,
		conf:    conf,
		records: records,
	}

	go cache.gcLoop()

	return cache
}

func (c *Cache) Dump() {
	log.Infof("Dump cache")
	sb := strings.Builder{}
	isFirst := true

	c.items.Range(func(k, v any) bool {
		if isFirst {
			sb.WriteString(fmt.Sprintf("Dump cache:{\"%v\": \"%v\"", k, v))
		} else {
			sb.WriteString(fmt.Sprintf(",\"%v\": \"%v\"", k, v))
		}
		return true
	})
	sb.WriteString("}")
	log.Infof(sb.String())
}

func (c *Cache) doMemoryClean() {
	c.items.Range(func(k, v any) bool {
		record, ok := v.(cachestore.Record)
		if ok {
			if time.Now().Unix() > record.GetExpireAt() {
				c.items.CompareAndDelete(k, v)
			}
		} else {
			c.items.Delete(k)
		}

		return true
	})
}

func (c *Cache) doDbClean() {
	if c.db == nil {
		return
	}
	var err error
	for _, r := range c.records {
		err = c.db.DeleteExpired(r)
		if err != nil {
			log.Errorf("Ignore error when doDbClean, record:%v, err:%v", reflect.TypeOf(r).Name(), err)
		}
	}
}

func (c *Cache) gcLoop() {
	if c.conf.GcInterval <= 0 {
		log.Infof("No gcLoop to clean expired map items for cache due to gcinterval:%v", c.conf.GcInterval)
		return
	}

	ticker := time.NewTicker(time.Minute * time.Duration(c.conf.GcInterval))
	for {
		select {
		case <-ticker.C:
			c.doMemoryClean()
			c.doDbClean()

		case <-c.ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (c *Cache) Update(rt cachestore.Record) error {
	if c.db != nil {
		err := c.db.Get(rt.Clone())
		if err != nil {
			return err
		}
	} else {
		if _, inMemory := c.items.Load(cachestore.UniqCacheKey(rt)); !inMemory {
			return ErrNotFound
		}
	}

	prevValue, loaded := c.items.Swap(cachestore.UniqCacheKey(rt), rt.GetValue())

	// check exist
	if c.db != nil {
		if (loaded && !cmp.Equal(prevValue, rt.GetValue())) || (!loaded && prevValue == nil) {
			if err := c.db.Set(rt); err != nil {
				return errors.Wrap(err)
			}
		}
	}

	return nil
}

func (c *Cache) CreateOrUpdate(rt cachestore.Record, expireAfterInSec int64) (exist bool, err error) {
	//TODO refine: if failed， rt should not change. it is better to use clone here
	originRt := rt.Clone()
	origExpireAt := rt.GetExpireAt()
	expireAt := time.Now().Unix() + expireAfterInSec
	if expireAfterInSec == 0 {
		expireAt = 0
	}
	rt.SetExpireAt(expireAt)
	prevValue, loaded := c.items.Swap(cachestore.UniqCacheKey(rt), rt.GetValue())

	if c.db != nil {
		if loaded && !cmp.Equal(prevValue, rt.GetValue()) {
			if err = c.db.Set(rt); err != nil {
				//rollback expireAt
				rt.SetExpireAt(origExpireAt)
				return false, errors.Wrap(err)
			}

			return true, nil
		}

		if !loaded && prevValue == nil {
			if err = c.db.Get(originRt); err == nil {
				loaded = true
			}

			// 存储到数据库
			if err = c.db.Set(rt); err != nil {
				//rollback expireAt
				rt.SetExpireAt(origExpireAt)
				return false, errors.Wrap(err)
			}
		}
	}

	return loaded, nil
}

func (c *Cache) Set(rt cachestore.Record, expireAfterInSec int64) error {
	//TODO refine: if failed， rt should not change. it is better to use clone here
	origExpireAt := rt.GetExpireAt()
	expireAt := time.Now().Unix() + expireAfterInSec
	if expireAfterInSec == 0 {
		expireAt = 0
	}
	rt.SetExpireAt(expireAt)
	prevValue, loaded := c.items.Swap(cachestore.UniqCacheKey(rt), rt.GetValue())

	if c.db != nil {
		if (loaded && !cmp.Equal(prevValue, rt.GetValue())) || (!loaded && prevValue == nil) {
			// 存储到数据库
			if err := c.db.Set(rt); err != nil {
				//rollback expireAt
				rt.SetExpireAt(origExpireAt)
				return errors.Wrap(err)
			}
		}
	}

	return nil
}

func ListWithKeyPrefix[T cachestore.ConsistentRecord](c *Cache, filterWithKeyPrefix T) ([]T, error) {
	var (
		err error
	)

	rsMerged := []T{}
	rsInDb := []T{}
	rsInMemMap := make(map[string]T)

	c.items.Range(func(k, v interface{}) bool {
		if strings.HasPrefix(k.(string), cachestore.UniqCacheKey(filterWithKeyPrefix)) {
			tmp := filterWithKeyPrefix.Clone() //get a clone
			err = tmp.Unmarshal(v.(string))
			if err != nil {
				return false
			}
			rsInMemMap[k.(string)] = tmp.(T)
		}
		return true
	})

	if c.db != nil {
		err = c.db.ListWithKeyPrefix(&rsInDb, filterWithKeyPrefix, filterWithKeyPrefix.GetPrimaryName(), filterWithKeyPrefix.GetKey())
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(err)
		}
	}

	for _, rs := range rsInDb {
		_, ok := rsInMemMap[cachestore.UniqCacheKey(rs)]
		if !ok {
			rsMerged = append(rsMerged, rs)
		}
	}

	for _, rs := range rsInMemMap {
		rsMerged = append(rsMerged, rs)
	}

	return rsMerged, nil
}

func (c *Cache) Get(filter cachestore.Record) (cachestore.Record, error) {
	var (
		dest cachestore.Record
		err  error
	)

	// 首先尝试从内存缓存获取
	value, inMemory := c.items.Load(cachestore.UniqCacheKey(filter))
	if inMemory {
		err = filter.Unmarshal(value.(string))
		if err != nil {
			return nil, err
		}
		dest = filter
	} else if c.db != nil {
		// Try from db
		// 如果缓存中没有，从数据库获取
		err := c.db.Get(filter)
		if err != nil {
			return nil, err // 其他错误
		}
		dest = filter
	}

	if dest == nil {
		return nil, ErrNotFound
	}

	if dest.GetExpireAt() != 0 && time.Now().Unix() > dest.GetExpireAt() {
		return nil, ErrExpired
	}

	if !inMemory {
		c.items.Store(cachestore.UniqCacheKey(dest), dest.GetValue())
	}

	return dest, nil
}

func (c *Cache) Exist(filter cachestore.Record) (bool, error) {
	// 首先尝试从内存缓存获取
	_, inMemory := c.items.Load(cachestore.UniqCacheKey(filter))
	if inMemory {
		return true, nil
	}

	if c.db != nil {
		// Try from db
		// 如果缓存中没有，从数据库获取
		err := c.db.Get(filter)
		return err == nil, err
	}

	return false, nil
}
