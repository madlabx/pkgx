package dbc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/madlabx/pkgx/errcode"
	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/memkv"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestSuite 是一组相关测试的集合
type TestPSqlSuite struct {
	suite.Suite
	cache *memkv.Cache
	R     *require.Assertions
}

func (s *TestPSqlSuite) SetT(t *testing.T) {
	s.Suite.SetT(t)
	s.R = s.Require()
}

var baseDb memkv.KvDbClientIf

func (s *TestPSqlSuite) SetupTest() {
	var err error
	fmt.Printf("Try connect db\n")
	baseDb, err = NewDbClient(context.Background(),
		SqlConfig{
			//Type:     "psql",
			//Host:     "127.0.0.1",
			//Port:     "5432",
			//User:     "postgres",
			//Password: "Lzj123456",
			//Dbname:   "nasdb",
			Log: LogConfig{
				Level: "error",
			},
			Type:   "sqllite",
			Dbname: "sqllite.db",
		},
		&RefreshTokenMock{})
	fmt.Printf("Done connect db\n")
	s.Nil(err)
	s.cache = memkv.NewCache(context.Background(), baseDb, memkv.CacheConf{})
}

func (s *TestPSqlSuite) TearDownSuite() {

}

func TestPSqlRunSuite(t *testing.T) {
	suite.Run(t, new(TestPSqlSuite))
}

func (s *TestPSqlSuite) TestNewCache() {
	s.NotNil(s.cache)
}

func (s *TestPSqlSuite) TestSetAndInExpire() {
	rt := (&RefreshTokenMock{IKey: "key2", IValue: "value1", IName: "name1"})

	err := s.cache.Set(rt, 5)

	s.R.Nil(err)

	s.cache.Dump()
	_, err = s.cache.Get(rt)
	s.R.Nil(err)

	s.cache.Dump()
	_, err = s.cache.Get(rt)
	s.R.Nil(err)
}

func (s *TestPSqlSuite) TestSetAndOutOfExpire() {
	rt := (&RefreshTokenMock{IKey: "key1", IValue: "value1", IName: "name1"})

	err := s.cache.Set(rt, 1)

	s.Nil(err)

	s.cache.Dump()
	_, err = s.cache.Get(rt)
	s.Nil(err)

	time.Sleep(time.Second * 2)

	s.cache.Dump()
	fmt.Printf("time:%v", time.Now().Unix())
	_, err = s.cache.Get(rt)
	s.Equal(memkv.ErrExpired, err)
}

func (s *TestPSqlSuite) TestGet_HitCache() {
	rt := (&RefreshTokenMock{IKey: "key2", IValue: "value1", IName: "name2"})

	err := s.cache.Set(rt, 5)

	s.Nil(err)

	result, err := s.cache.Get(rt)
	s.NoError(err)
	s.Equal(rt, result)
}

func (s *TestPSqlSuite) TestGet_MissCache_HitDB() {
	//prepare

	rt := (&RefreshTokenMock{IKey: "key3", IValue: "getNewValue"})
	rt.SetExpireAt(time.Now().Unix() + 5)
	err := baseDb.Set(rt)
	s.R.Nil(err)

	newRecord, err := s.cache.Get(&RefreshTokenMock{IKey: "key3"})
	s.R.Equal(nil, err)
	s.R.NotNil(newRecord)
	s.R.Equal("getNewValue", newRecord.GetValue())

	newRecord, err = s.cache.Get(rt)
	s.Require().Equal(nil, err)
	s.Require().Equal("getNewValue", newRecord.GetValue())

}

func (s *TestPSqlSuite) TestGet_MissCache_MissDB() {
	rt := (&RefreshTokenMock{IKey: "key3", IValue: "value1", IName: "name3"})

	_, err := s.cache.Get(rt)
	ok := errors.Is(errcode.ErrObjectNotExist(), err)
	s.True(ok)
}
