package dbc

import (
	"context"
	"maps"
	"strings"
	"sync"

	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/log"
	"gorm.io/gorm"
)

var ErrStopWalk = errors.New("stop walk")

type MultipleDb struct {
	m        sync.RWMutex
	dbMap    map[uint64]*DbClient
	extDbMap map[uint64]*DbClient
	tables   []interface{}
	ctx      context.Context
}

func NewTestMultipleDbClient(db *gorm.DB) *MultipleDb {
	return &MultipleDb{
		dbMap: map[uint64]*DbClient{
			1: NewTestDbClient(db),
		},
	}
}

func NewMultipleDb(pctx context.Context, tables ...interface{}) *MultipleDb {
	return &MultipleDb{
		dbMap:    make(map[uint64]*DbClient),
		extDbMap: make(map[uint64]*DbClient),
		tables:   tables,
		ctx:      pctx,
	}
}

// For test
func (d *MultipleDb) AddTestDbClient(id uint64, db *gorm.DB) error {
	d.m.Lock()
	defer d.m.Unlock()
	d.dbMap[id] = NewTestDbClient(db)
	return nil
}

func (d *MultipleDb) Add(id uint64, conf SqlConfig) error {
	db, err := NewDbClient(d.ctx, conf, d.tables...)
	if err != nil {
		return err
	}

	d.m.Lock()
	defer d.m.Unlock()
	if conf.IsExt() {
		d.extDbMap[id] = db
	} else {
		d.dbMap[id] = db
	}
	return nil
}

func (d *MultipleDb) Remove(id uint64) {
	d.m.Lock()
	defer d.m.Unlock()
	delete(d.dbMap, id)
	delete(d.extDbMap, id)
}

func (d *MultipleDb) Db() map[uint64]*DbClient {
	ret := make(map[uint64]*DbClient)
	d.m.RLock()
	defer d.m.RUnlock()
	for id, db := range d.dbMap {
		ret[id] = db
	}
	return ret
}

func (d *MultipleDb) ExtDb() *MultipleDb {
	ret := make(map[uint64]*DbClient)
	d.m.RLock()
	defer d.m.RUnlock()
	maps.Copy(ret, d.extDbMap)
	return &MultipleDb{
		dbMap:    make(map[uint64]*DbClient),
		extDbMap: ret,
		tables:   d.tables,
		ctx:      d.ctx,
	}
}

// AllReadyDb 只返回所有已经初始化完成的 db
func (d *MultipleDb) AllReadyDb() map[uint64]*DbClient {
	d.m.RLock()
	defer d.m.RUnlock()
	ret := make(map[uint64]*DbClient)
	maps.Copy(ret, d.dbMap)
	for id, db := range d.extDbMap {
		if db.IsInitCompleted() {
			ret[id] = db
		}
	}
	return ret
}

func (d *MultipleDb) AllDb() map[uint64]*DbClient {
	ret := make(map[uint64]*DbClient)
	d.m.RLock()
	defer d.m.RUnlock()
	maps.Copy(ret, d.dbMap)
	maps.Copy(ret, d.extDbMap)
	return ret
}

// SelectExtDbByPath 根据路径选择 ext db
func (d *MultipleDb) SelectExtDbByPath(path string) *MultipleDb {
	ret := make(map[uint64]*DbClient)
	d.m.RLock()
	defer d.m.RUnlock()
	for id, db := range d.extDbMap {
		if strings.HasPrefix(path, db.extDbPrefix) {
			ret[id] = db
			break
		}
	}
	return &MultipleDb{
		dbMap:    make(map[uint64]*DbClient),
		extDbMap: ret,
		tables:   d.tables,
		ctx:      d.ctx,
	}
}

// SelectExtDbByPathWithInnerDb 根据路径选择 ext db，并且附带内部磁盘的 db
func (d *MultipleDb) SelectExtDbByPathWithInnerDb(path string) *MultipleDb {
	ret := make(map[uint64]*DbClient)
	inner := make(map[uint64]*DbClient)
	d.m.RLock()
	defer d.m.RUnlock()
	for id, db := range d.extDbMap {
		if strings.HasPrefix(path, db.extDbPrefix) {
			ret[id] = db
			break
		}
	}
	maps.Copy(inner, d.dbMap)
	return &MultipleDb{
		dbMap:    inner,
		extDbMap: ret,
		tables:   d.tables,
		ctx:      d.ctx,
	}
}

// SelectExtDbByPathWhenNotFoundUseInnerDb 根据路径选择 ext db，如果 ext db 不存在，则使用内部 db
func (d *MultipleDb) SelectExtDbByPathWhenNotFoundUseInnerDb(path string) *MultipleDb {
	ret := make(map[uint64]*DbClient)
	inner := make(map[uint64]*DbClient)
	d.m.RLock()
	defer d.m.RUnlock()
	for id, db := range d.extDbMap {
		if strings.HasPrefix(path, db.extDbPrefix) {
			ret[id] = db
			break
		}
	}
	if len(ret) == 0 {
		maps.Copy(inner, d.dbMap)
	}
	return &MultipleDb{
		dbMap:    inner,
		extDbMap: ret,
		tables:   d.tables,
		ctx:      d.ctx,
	}
}

// DbById 根据 id 来拿 db
// 在 scan 设备的时候，堆积了很多事件在 chan 里面，但此时将设备拔出，对应的 db 会删除
// 此时事件还在处理中，通过 id 只会拿到 nil 的 db 对象，会导致整个程序 panic 掉
// 解决方法：
// 1. 所有用到 DbById 的地方，先 check nil
// 2. DbById 函数要返回一个 error
// 3. DbById 返回一个带有错误的 gorm.DB 对象（选择这个，改动最小）
func (d *MultipleDb) DbById(id uint64) *DbClient {
	d.m.RLock()
	defer d.m.RUnlock()
	if db, ok := d.dbMap[id]; ok {
		return db
	}
	if db, ok := d.extDbMap[id]; ok {
		return db
	}
	err := errors.Errorf("db id %d not found", id)
	log.Errorf("err in DbById, err:%+v", err)
	return NewBadDbclient(err)
}

// WalkRawDb_DO_NOT_USE go through all DBs and execute f
func (d *MultipleDb) WalkRawDb_DO_NOT_USE(f func(uint64, *gorm.DB) error) error {
	for did, diskDb := range d.Db() {
		if err := f(did, diskDb.DB()); err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}

// Walk stop walking when f returns an error,
// if f returns ErrStopWalk, Walk will return nil,
// otherwise, it will return the error returned by f
func (d *MultipleDb) Walk(f func(uint64, *DbClient) error) error {
	for diskId, diskDb := range d.Db() {
		if err := f(diskId, diskDb); err != nil {
			if errors.Is(err, ErrStopWalk) {
				return nil
			}
			return errors.Wrap(err)
		}
	}
	return nil
}

func (d *MultipleDb) WalkExtDb(f func(uint64, *DbClient) error) error {
	ret := make(map[uint64]*DbClient)
	d.m.RLock()
	maps.Copy(ret, d.extDbMap)
	d.m.RUnlock()

	for id, db := range ret {
		if err := f(id, db); err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}

func (d *MultipleDb) WalkAllDb(f func(uint64, *DbClient) error) error {
	ret := make(map[uint64]*DbClient)
	d.m.RLock()
	maps.Copy(ret, d.extDbMap)
	maps.Copy(ret, d.dbMap)
	d.m.RUnlock()

	for id, db := range ret {
		if err := f(id, db); err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}
