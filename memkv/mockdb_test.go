package memkv

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/madlabx/pkgx/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

var _ KvDbClientIf = &MockDbClient{}

// MockDbClient 用于模拟数据库客户端行为
type MockDbClient struct {
	mock.Mock
}

func (m *MockDbClient) ListWithKeyPrefix(dest any, filter any, keyFieldName, keyPrefix string) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockDbClient) Set(record any) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *MockDbClient) Get(record any) error {
	args := m.Called(record)

	return args.Error(0)
}

func (m *MockDbClient) DeleteExpired(record any) error {
	args := m.Called(record)
	return args.Error(0)
}

// TestSuite 是一组相关测试的集合
type TestMockDbSuite struct {
	suite.Suite
	cache  *Cache
	mockDB *MockDbClient
}

func (s *TestMockDbSuite) SetupTest() {
	s.mockDB = new(MockDbClient)

	s.cache = NewCache(context.Background(), s.mockDB, CacheConf{})

}

func (s *TestMockDbSuite) TearDownSuite() {

}

func (s *TestMockDbSuite) BeforeTest(suitename, testname string) {
	if testname == "TestOnlyMemCache" {
		s.cache = NewCache(context.Background(), nil, CacheConf{})
	}
}

func TestMockDbRunSuite(t *testing.T) {
	suite.Run(t, new(TestMockDbSuite))
}

func (s *TestMockDbSuite) TestNewCache() {
	s.NotNil(s.cache)
	s.NotNil(s.cache.items)
}

func (s *TestMockDbSuite) TestSetAndInExpire() {
	rt := (&RefreshTokenMock{IKey: "key1", IValue: "value1", IName: "name1"})
	s.mockDB.On("Set", rt).Return(nil)

	err := s.cache.Set(rt, 5)
	s.Nil(err)

	s.cache.Dump()

	s.mockDB.On("Get", rt).Return(nil).Run(func(args mock.Arguments) {
		token := args[0].(*RefreshTokenMock)
		token.SetValue("value1")
	})

	_, err = s.cache.Get(rt)
	s.Nil(err)

	s.cache.Dump()
	_, err = s.cache.Get(rt)
	s.Nil(err)
}

func (s *TestMockDbSuite) TestSetAndOutOfExpire() {
	rt := (&RefreshTokenMock{IKey: "key1", IValue: "value1", IName: "name1"})
	s.mockDB.On("Set", rt).Return(nil)

	err := s.cache.Set(rt, 1)
	s.Nil(err)

	s.cache.Dump()
	s.mockDB.On("Get", rt).Return(nil).Run(func(args mock.Arguments) {
		token := args[0].(*RefreshTokenMock)
		token.SetValue("getNewValue")
	})
	_, err = s.cache.Get(rt)
	s.Nil(err)

	time.Sleep(time.Second * 2)

	s.cache.Dump()
	fmt.Printf("time:%v", time.Now().Unix())
	_, err = s.cache.Get(rt)
	s.Equal(ErrExpired, err)
}

func (s *TestMockDbSuite) TestGet_HitCache() {

	rt := (&RefreshTokenMock{IKey: "key2", IValue: "value2", IName: "name2"})

	s.mockDB.On("Set", rt).Return(nil)
	err := s.cache.Set(rt, 5)
	s.Nil(err)
	s.mockDB.On("Get", rt).Return(nil).Run(func(args mock.Arguments) {
		token := args[0].(*RefreshTokenMock)
		token.SetValue("getNewValue")
	})
	result, err := s.cache.Get(rt)
	s.NoError(err)
	s.Equal(rt, result)

}

func (s *TestMockDbSuite) TestGet_MissCache_HitDB() {
	rt := (&RefreshTokenMock{IKey: "key3"})

	s.mockDB.On("Get", rt).Return(nil).Run(func(args mock.Arguments) {
		token := args[0].(*RefreshTokenMock)
		token.SetValue("getNewValue")
		token.SetExpireAt(time.Now().Unix() + 5)
	})

	s.mockDB.On("Set", rt, 5).Return(nil)

	newRecord, err := s.cache.Get(rt)
	s.Equal(nil, err)
	s.Equal("getNewValue", newRecord.GetValue())

	s.mockDB.On("FindUniq", rt).Return(gorm.ErrRecordNotFound)
	newRecord, err = s.cache.Get(rt)
	s.Equal(nil, err)
	s.Equal("getNewValue", newRecord.GetValue())

}

func (s *TestMockDbSuite) TestGet_MissCache_MissDB() {
	rt := (&RefreshTokenMock{IKey: "key3", IValue: "value3", IName: "name3"})

	s.mockDB.On("Get", rt).Return(gorm.ErrRecordNotFound)

	_, err := s.cache.Get(rt)
	ok := errors.Is(gorm.ErrRecordNotFound, err)
	s.True(ok)
}

func (s *TestMockDbSuite) TestOnlyMemCache() {
	s.Nil(s.cache.db)

	rt := &RefreshTokenMock{IKey: "key3"}
	ret, err := s.cache.Get(rt)
	s.Nil(ret)
	s.True(errors.Is(err, gorm.ErrRecordNotFound))

	err = s.cache.Set(rt, 3600)
	s.Nil(err)
}
