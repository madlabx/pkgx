package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/madlabx/pkgx/cachestore"
	"github.com/madlabx/pkgx/errcode"
	"github.com/madlabx/pkgx/testingx"
	"github.com/madlabx/pkgx/utils"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	// Define our test cases
	testingx.Skip(t, "NeedRedis")
	rc, err := NewClient(context.Background(), Config{})
	require.Nil(t, err)

	keyPreview := utils.RandomString(10)
	keys := []string{"key1", "key2", "key3"}

	defer func() {
		for _, k := range keys {
			require.Nil(t, rc.Del(&RefreshTokenMock{IKey: keyPreview + k}))
		}
	}()

	for _, k := range keys {
		require.Nil(t, rc.Set(&RefreshTokenMock{IKey: keyPreview + k}, 1000))
	}

	results, err := List(rc, &RefreshTokenMock{})
	require.Nil(t, err)

	require.Equal(t, 3, len(results))

	fmt.Printf("results:%v\n", utils.ToString(results))

}

func TestCompareAndDelete(t *testing.T) {
	// Define our test cases
	testingx.Skip(t, "NeedRedis")
	rc, err := NewClient(context.Background(), Config{})
	require.Nil(t, err)

	keyPreview := utils.RandomString(10)
	keys := []string{"key1", "key2", "key3"}

	defer func() {
		for _, k := range keys {
			require.Nil(t, rc.Del(&RefreshTokenMock{IKey: keyPreview + k}))
		}
	}()

	for _, k := range keys {
		require.Nil(t, rc.Set(&RefreshTokenMock{IKey: keyPreview + k}, 1000))
	}

	done, err := rc.CompareAndDelete(&RefreshTokenMock{IKey: keyPreview + keys[1], IValue: "key2IName"}, func(src, dst cachestore.Record) bool {
		return src.(*RefreshTokenMock).IValue == dst.(*RefreshTokenMock).IValue
	})
	require.Nil(t, err)
	require.Equal(t, false, done)

	results2, err := List(rc, &RefreshTokenMock{})
	require.Nil(t, err)
	require.Equal(t, 3, len(results2))

	done, err = rc.CompareAndDelete(&RefreshTokenMock{IKey: keyPreview + keys[1], IValue: ""}, func(src, dst cachestore.Record) bool {
		return src.(*RefreshTokenMock).IValue == dst.(*RefreshTokenMock).IValue
	})
	require.Nil(t, err)
	require.Equal(t, true, done)

	results2, err = List(rc, &RefreshTokenMock{})
	require.Nil(t, err)
	require.Equal(t, 2, len(results2))

	done, err = rc.CompareAndDelete(&RefreshTokenMock{IKey: keyPreview + "NotExist", IValue: ""}, func(src, dst cachestore.Record) bool {
		return src.(*RefreshTokenMock).IValue == dst.(*RefreshTokenMock).IValue
	})
	require.True(t, errcode.IsNotFound(err))
	require.Equal(t, false, done)

}

func TestUpdate(t *testing.T) {
	// Define our test cases
	testingx.Skip(t, "NeedRedis")
	rc, err := NewClient(context.Background(), Config{})
	require.Nil(t, err)

	keyPreview := utils.RandomString(10)
	keys := []string{"key1"}
	//
	defer func() {
		for _, k := range keys {
			require.Nil(t, rc.Del(&RefreshTokenMock{IKey: keyPreview + k}))
		}
	}()

	for _, k := range keys {
		require.Nil(t, rc.Set(&RefreshTokenMock{IKey: keyPreview + k, IValue: "wef"}, 5))
	}

	results, err := List(rc, &RefreshTokenMock{})
	require.Nil(t, err)

	require.Equal(t, 1, len(results))

	time.Sleep(time.Second * 2)

	results, err = List(rc, &RefreshTokenMock{})
	require.Nil(t, err)

	require.Equal(t, 1, len(results))

	fmt.Printf("before results:%v\n", utils.ToString(results))
	for _, k := range keys {
		require.Nil(t, rc.Update(&RefreshTokenMock{IKey: keyPreview + k, IValue: "wef1"}))
	}

	time.Sleep(time.Second * 2)
	results, err = List(rc, &RefreshTokenMock{})
	require.Nil(t, err)

	require.Equal(t, 1, len(results))
	time.Sleep(time.Second * 2)
	results, err = List(rc, &RefreshTokenMock{})
	require.Nil(t, err)
	require.Equal(t, nil, err)

	require.Equal(t, 0, len(results))

}

func TestUpdateTTL(t *testing.T) {
	// Define our test cases
	testingx.Skip(t, "NeedRedis")
	rc, err := NewClient(context.Background(), Config{})
	require.Nil(t, err)

	keyPreview := utils.RandomString(10)
	keys := []string{"key1"}

	defer func() {
		for _, k := range keys {
			require.Nil(t, rc.Del(&RefreshTokenMock{IKey: keyPreview + k}))
		}
	}()

	for _, k := range keys {
		require.Nil(t, rc.Set(&RefreshTokenMock{IKey: keyPreview + k}, 0))
	}

	//TTL = -2
	err = rc.Update(&RefreshTokenMock{IKey: keyPreview + "NotExistKey", IValue: "wef1"})
	require.Truef(t, errcode.IsNotFound(err), "err:%v", err)

	//TTL = -1
	testingx.NilAndLog(t, rc.Update(&RefreshTokenMock{IKey: keyPreview + "key1", IValue: "wef1"}))
	ttl, err := rc.rc.TTL(context.Background(), cachestore.UniqCacheKey(&RefreshTokenMock{IKey: keyPreview + "key1"})).Result()
	require.Nil(t, err)
	require.Equal(t, time.Duration(-1), ttl)
}
