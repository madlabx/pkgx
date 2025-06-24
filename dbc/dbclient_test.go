package dbc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/madlabx/pkgx/cachestore"
	"github.com/madlabx/pkgx/errcode"
	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/log"
	"github.com/madlabx/pkgx/testingx"
	"github.com/madlabx/pkgx/utils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDb() (*DbClient, error) {
	return NewDbClient(context.Background(),
		SqlConfig{
			//Type:     "mysql",
			//Host:     "127.0.0.1",
			//Port:     "3306",
			//User:     "root",
			//Password: "Lzj123456",
			//Dbname:   "nas_cloud_service",
			//Log: LogConfig{
			//	Level: "error",
			//},

			//Type:     "psql",
			//Host:     "127.0.0.1",
			//Port:     "5432",
			//User:     "postgres",
			//Password: "Lzj123456",
			//Dbname:   "nasdb",
			Log: LogConfig{
				Level: "info",
			},
			Type:   "sqllite",
			Dbname: "sqllite.db",
		},
		&RefreshTokenMock{}, &TestFileInfo{}, &TestUserInfo{}, &TestVideoInfo{})
}

// RefreshTokenMock 作为Record接口的一个简单实现，用于测试
type RefreshTokenMock struct {
	ID        uint   `gorm:"primaryKey"`
	IKey      string `gorm:"column:key;"`      // 主键，字段大小，不允许为空
	IValue    string `gorm:"column:value"`     // 令牌值，字段大小，不允许为空
	IName     string `gorm:"column:name"`      // 令牌名称，字段大小，默认值
	IExpireAt int64  `gorm:"column:expire_at"` // 过期时间，不允许为空，使用当前时间戳作为默认值
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

func (r *RefreshTokenMock) TableName() string { return "refresh_token_mock_d11" }

func (r *RefreshTokenMock) save(dbc *DbClient) error {
	return dbc.Save(r)
}

func TestSavePtr(t *testing.T) {
	baseDb, err := newDb()

	require.Nil(t, err)

	a := &RefreshTokenMock{
		IKey: "key1",
	}

	err = baseDb.Save(a)
	require.Nil(t, err)
	require.NotEqual(t, 0, a.ID)

	b := &RefreshTokenMock{
		IKey: "key2",
	}
	err = baseDb.Save(b)
	require.Nil(t, err)
	require.NotEqual(t, 0, b.ID)

	c := &RefreshTokenMock{
		IKey: "key2",
	}
	err = baseDb.Save(c)
	require.Nil(t, err)
	require.NotEqual(t, c.ID, b.ID)
}

func TestStructSave(t *testing.T) {
	baseDb, err := newDb()

	require.Nil(t, err)

	a := RefreshTokenMock{
		IKey: "key0",
	}

	err = a.save(baseDb)
	require.Nil(t, err)

	b := &RefreshTokenMock{
		IKey: "key12",
	}
	err = b.save(baseDb)
	require.Nil(t, err)
}

func TestSave(t *testing.T) {

	baseDb, err := newDb()
	//baseDb, err := NewDbClient(context.Background(),
	//	cfg.Get().BaseSql,
	//	users.UserInfo{})

	require.Nil(t, err)

	a := RefreshTokenMock{
		IKey: "key1",
	}

	err = baseDb.Save(&a)

	require.Nil(t, err)

	b := RefreshTokenMock{
		IKey: "key2",
	}
	err = baseDb.Save(&b)

	require.Nil(t, err)
}

func TestConnectNonExistDb(t *testing.T) {
	testingx.Skip(t, "NeedMySQL")
	conf := SqlConfig{
		Type:     "mysql",
		Host:     "127.0.0.1",
		Port:     "3306",
		User:     "root",
		Password: "Lzj123456",
		Dbname:   "fake_nas_cloud_service",
		Log: LogConfig{
			Level: "error",
		}}
	_ = dropDb(conf)
	err := createDb(conf)

	require.Nil(t, err)

	err = dropDb(conf)

	require.Nil(t, err)

}

// TestFileInfo describes a file.
type TestFileInfo struct {
	FId       uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	Parent    string `gorm:"column:parent"`
	Path      string `gorm:"column:path"`
	Name      string `gorm:"column:name"`
	Length    string
	NewLength string

	// DB record time, just ignore in json
	RecordCreateAt *time.Time `gorm:"column:record_create_at;autoCreateTime" json:"-"`
	RecordModifyAt *time.Time `gorm:"column:record_modify_at;autoUpdateTime" json:"-"`
}

func (fi *TestFileInfo) TableName() string {
	return "test_file_info"
}

func getFileInfoByFId(fid uint64, db *DbClient) (*TestFileInfo, error) {
	fi := &TestFileInfo{}
	return fi, db.GetByPrimary(fi, fid)
}

func getFileInfosByFIds(fids []uint64, db *DbClient) ([]TestFileInfo, error) {
	var fis []TestFileInfo
	return fis, db.ListByPrimaryKeys(&fis, fids)
}

func TestList(t *testing.T) {
	baseDb, err := newDb()
	//baseDb, err := NewDbClient(context.Background(),
	//	cfg.Get().BaseSql,
	//	users.UserInfo{})

	require.Nil(t, err)

	name := "Special123456"
	fi := TestFileInfo{Path: "/file/to/xxx", Name: name}
	err = baseDb.Save(&fi)
	require.Nil(t, err)
	fi4 := TestFileInfo{Path: "/file/to/xxx", Name: name}
	err = baseDb.Save(&fi4)
	require.Nil(t, err)

	fi5 := TestFileInfo{Path: "/file/to/xxx", Name: name}
	err = baseDb.Save(&fi5)
	require.Nil(t, err)

	var fis []TestFileInfo
	err = baseDb.List(&fis, TestFileInfo{Name: name})
	require.Nil(t, err)
	require.Equal(t, 3, len(fis))

	err = baseDb.DeleteByPrimaryKeys(&TestFileInfo{}, []uint64{fi.FId, fi4.FId, fi5.FId})
	require.Nil(t, err)
	err = baseDb.List(&fis, TestFileInfo{Name: name})
	require.True(t, errors.Is(err, errcode.ErrObjectNotExist()))
}

func TestGetByPrimary(t *testing.T) {
	baseDb, err := newDb()
	require.Nil(t, err)

	name := "specialForTestXXX"
	fi := TestFileInfo{Path: "/file/to/xxx", Name: name}

	err = baseDb.Save(&fi)
	require.Nil(t, err)

	dst := TestFileInfo{}
	err = baseDb.GetByPrimary(&dst, fi.FId)

	require.Nil(t, err)
	dst2 := TestFileInfo{}
	err = baseDb.GetByPrimary(&dst2, fi.FId)

	require.Nil(t, err)
	require.Equal(t, dst2, dst)
	fi2, err := getFileInfoByFId(fi.FId, baseDb)
	require.Nil(t, err)
	require.Equal(t, dst2, *fi2)

	fi4 := TestFileInfo{Path: "/file/to/xxx", Name: name}
	err = baseDb.Save(&fi4)
	require.Nil(t, err)
	fi5 := TestFileInfo{Path: "/file/to/xxx", Name: name}
	err = baseDb.Save(&fi5)
	require.Nil(t, err)
	ids := []uint64{fi.FId, fi4.FId, fi5.FId}
	fis, err := getFileInfosByFIds(ids, baseDb)
	require.Nil(t, err)
	require.Equal(t, 3, len(fis))

	require.Nil(t, baseDb.DeleteByPrimaryKeys(&TestFileInfo{}, ids))

	fis, err = getFileInfosByFIds(ids, baseDb)
	require.Nil(t, err)
	require.Equal(t, 0, len(fis))

}
func TestGetByPrimary2(t *testing.T) {
	baseDb, err := newDb()
	//baseDb, err := NewDbClient(context.Background(),
	//	cfg.Get().BaseSql,
	//	users.UserInfo{})

	require.Nil(t, err)

	fi := TestFileInfo{Path: "/file/to/xxx", Name: "123"}

	err = baseDb.Save(&fi)
	require.Nil(t, err)

	dst := TestFileInfo{}
	err = baseDb.GetByPrimary(&dst, fi.FId)
	require.Nil(t, err)

	dst2 := TestFileInfo{}
	err = baseDb.GetByPrimary(&dst2, fi.FId)

	require.Nil(t, err)
	require.Equal(t, dst, dst2)

	fi2, err := getFileInfoByFId(fi.FId, baseDb)
	require.Nil(t, err)
	require.Equal(t, dst, *fi2)

	fi4 := TestFileInfo{Path: "/file/to/xxx", Name: "123"}
	err = baseDb.Save(&fi4)
	require.Nil(t, err)
	fi5 := TestFileInfo{Path: "/file/to/xxx", Name: "123"}
	err = baseDb.Save(&fi5)
	require.Nil(t, err)
	ids := []uint64{fi.FId, fi4.FId, fi5.FId}
	fis, err := getFileInfosByFIds(ids, baseDb)
	require.Nil(t, err)
	require.Equal(t, 3, len(fis))

	require.Nil(t, baseDb.DeleteByPrimaryKeys(&TestFileInfo{}, ids))

	fis, err = getFileInfosByFIds(ids, baseDb)
	require.Nil(t, err)
	require.Equal(t, 0, len(fis))
}

// TestFileInfo describes a file.
type TestUserInfo struct {
	Id                   uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name                 string     `gorm:"column:name;index"`
	CloudUserId          uint64     `gorm:"column:cloud_user_id"`
	Password             string     `gorm:"column:password"`
	EnableSamba          bool       `gorm:"column:samba"`
	SambaPasswdNoEncrypt string     `gorm:"column:samba_passwd"`
	CreateAt             *time.Time `gorm:"column:create_at;autoCreateTime"`
	ModifyAt             *time.Time `gorm:"column:modify_at;autoUpdateTime"`
	Fs                   afero.Fs   `json:"-" gorm:"-"`
}

func (t TestUserInfo) TableName() string {
	return "test_user_info"
}

func TestUpdatesWithZero(t *testing.T) {
	baseDb, err := newDb()
	require.Nil(t, err)

	fi := TestUserInfo{}
	err = baseDb.FirstOrCreate(&fi, TestUserInfo{Name: "test", CloudUserId: 1, Password: "ewfwe", EnableSamba: true})
	require.Nil(t, err)
	user := TestUserInfo{Id: fi.Id}
	err = baseDb.FindUniq(&user)
	require.Nil(t, err)
	log.Errorf("User:%#v", user)

	err = baseDb.UpdatesWithZero(TestUserInfo{Id: fi.Id}, map[string]interface{}{"samba": false})
	assert.Nil(t, err)

	user = TestUserInfo{Id: fi.Id}
	err = baseDb.FindUniq(&user)
	log.Errorf("User:%#v", user)
	require.Nil(t, err)
	require.False(t, user.EnableSamba)
}

func TestFindAndUpdateManyRecords(t *testing.T) {
	baseDb, err := newDb()
	require.Nil(t, err)

	var fiArr []TestFileInfo
	//do clean
	doClean := func(name string) {
		err = baseDb.Delete(&TestFileInfo{Name: name})
		//assert.Nil(t, err)

		err = baseDb.List(&fiArr, TestFileInfo{Name: name})
		require.True(t, errors.Is(err, errcode.ErrObjectNotExist()))
		require.Equal(t, 0, len(fiArr))
	}
	name := utils.RandomString(10)
	doClean(name)
	defer doClean(name)

	newPath := "newPath"
	err = baseDb.UpdatesOmitZero(TestFileInfo{Name: name}, TestFileInfo{Path: newPath})
	assert.Nil(t, err)

	err = baseDb.List(&fiArr, TestFileInfo{Name: name})
	require.True(t, errors.Is(err, errcode.ErrObjectNotExist()))
	require.Equal(t, 0, len(fiArr))

	err = baseDb.FirstOrCreate(&TestFileInfo{}, TestFileInfo{Path: "1", Name: name})
	assert.Nil(t, err)
	err = baseDb.FirstOrCreate(&TestFileInfo{}, TestFileInfo{Path: "2", Name: name})
	assert.Nil(t, err)
	err = baseDb.FirstOrCreate(&TestFileInfo{}, TestFileInfo{Path: "3", Name: name})
	assert.Nil(t, err)

	//err = baseDb.ListWithAscOrder(&fiArr, TestFileInfo{Name: name}, "path")
	//require.Nil(t, err)
	//require.Equal(t, 3, len(fiArr))
	//require.Equal(t, "1", fiArr[0].Path)
	//require.Equal(t, "2", fiArr[1].Path)
	//require.Equal(t, "3", fiArr[2].Path)

	err = baseDb.UpdatesOmitZero(TestFileInfo{Name: name}, TestFileInfo{Path: newPath})
	assert.Nil(t, err)

	err = baseDb.List(&fiArr, TestFileInfo{Name: name})
	require.Nil(t, err)
	require.Equal(t, 3, len(fiArr))
	require.Equal(t, newPath, fiArr[0].Path)
	require.Equal(t, newPath, fiArr[1].Path)
	require.Equal(t, newPath, fiArr[2].Path)
}

func TestGormColumn(t *testing.T) {
	name, err := GormColumn(TestFileInfo{Name: "wefe"})
	require.Nil(t, err)
	require.Equal(t, "name", name)

	name, err = GormColumn(TestFileInfo{FId: 123})
	require.Nil(t, err)
	require.Equal(t, "id", name)

	name, err = GormColumn(TestFileInfo{Length: "123"})
	require.Nil(t, err)
	require.Equal(t, "length", name)

	name, err = GormColumn(TestFileInfo{NewLength: "123"})
	require.Nil(t, err)
	require.Equal(t, "new_length", name)

	name, err = GormColumn(TestFileInfo{Name: "wefe", FId: 1})
	require.NotNil(t, err)
	require.Equal(t, "", name)
}

func TestListWithOneAttr(t *testing.T) {
	baseDb, err := newDb()
	require.Nil(t, err)
	var fiArr []TestFileInfo
	//do clean
	doClean := func(name string) {
		err = baseDb.Delete(&TestFileInfo{Name: name})
		//assert.Nil(t, err)

		err = baseDb.List(&fiArr, TestFileInfo{Name: name})
		require.True(t, errors.Is(err, errcode.ErrObjectNotExist()))
		require.Equal(t, 0, len(fiArr))
	}
	name := utils.RandomString(10)
	doClean(name)
	defer doClean(name)

	var tfi []TestFileInfo
	err = baseDb.FirstOrCreate(&TestFileInfo{}, TestFileInfo{Path: "1", Name: name})
	assert.Nil(t, err)
	err = baseDb.FirstOrCreate(&TestFileInfo{}, TestFileInfo{Path: "2", Name: name})
	assert.Nil(t, err)
	err = baseDb.FirstOrCreate(&TestFileInfo{}, TestFileInfo{Path: "3", Name: name})
	assert.Nil(t, err)

	err = baseDb.ListWithOneAttr(&tfi, TestFileInfo{Path: "wefe"}, []string{"1", "2", "3"})
	require.Nil(t, err)
	log.Errorf("ft:%v", tfi)
}

func TestDelete(t *testing.T) {
	baseDb, err := newDb()
	require.Nil(t, err)
	var fiArr []TestFileInfo
	//do clean
	doClean := func(name string) {
		err = baseDb.Delete(&TestFileInfo{Name: name})
		//assert.Nil(t, err)

		err = baseDb.List(&fiArr, TestFileInfo{Name: name})
		require.True(t, errors.Is(err, errcode.ErrObjectNotExist()))
		require.Equal(t, 0, len(fiArr))
	}
	name := utils.RandomString(10)
	doClean(name)
	defer doClean(name)

	err = baseDb.FirstOrCreate(&TestFileInfo{}, TestFileInfo{Path: "1", Name: name})
	assert.Nil(t, err)

	err = baseDb.List(&fiArr, TestFileInfo{Name: name})
	require.Equal(t, 1, len(fiArr))
	log.Errorf("fiArr[0]:%#v", fiArr[0])

	var tfi []TestFileInfo
	log.Errorf("ft:%#v", tfi)
	err = baseDb.DB().Delete(&tfi, TestFileInfo{Path: "1"}).Error
	require.Nil(t, err)
	log.Errorf("ft:%#v", tfi)

}

type dbRecord interface {
	TableName() string
}

func deleteWildRecords(db *DbClient, in dbRecord) error {
	nonMatchingFids := []uint64{}
	//ftm.dbc.LeftJoin(&VideoBriefInfo{}, &fileinfo.FileInfo{}).Scan(nonMatchingFiles, "v.fid IS NULL")
	fileTable := (&TestFileInfo{}).TableName()

	err := db.Table(fmt.Sprintf("%s LEFT JOIN %s on %s.fid = %s.id", in.TableName(), fileTable, in.TableName(), fileTable)).
		Where(fmt.Sprintf("%s.id IS NULL", fileTable)).
		Select(fmt.Sprintf("%s.fid", in.TableName())).
		Scan(&nonMatchingFids).Error
	if err != nil {
		return errors.Wrap(err)
	}

	log.Errorf("Got records %d", len(nonMatchingFids))
	if len(nonMatchingFids) > 0 {
		log.Warnf("Find %d wild %v records, fid:%v", len(nonMatchingFids), in.TableName(), nonMatchingFids)
		return errors.Wrap(db.DeleteByPrimaryKeys(in, nonMatchingFids))
	}

	return nil
}

// TestFileInfo describes a file.
type TestVideoInfo struct {
	FId                  uint64     `gorm:"column:fid;primaryKey"`
	Name                 string     `gorm:"column:name;index"`
	CloudUserId          uint64     `gorm:"column:cloud_user_id"`
	Password             string     `gorm:"column:password"`
	EnableSamba          bool       `gorm:"column:samba"`
	SambaPasswdNoEncrypt string     `gorm:"column:samba_passwd"`
	CreateAt             *time.Time `gorm:"column:create_at;autoCreateTime"`
	ModifyAt             *time.Time `gorm:"column:modify_at;autoUpdateTime"`
	Fs                   afero.Fs   `json:"-" gorm:"-"`
}

func (t TestVideoInfo) TableName() string {
	return "test_video_info"
}

func TestLeftJoinsAndCount(t *testing.T) {
	baseDb, err := newDb()
	require.Nil(t, err)
	var fiArr []TestFileInfo
	//do clean
	doClean := func(name string) {
		err = baseDb.Delete(&TestFileInfo{Name: name})
		//assert.Nil(t, err)

		err = baseDb.List(&fiArr, TestFileInfo{Name: name})
		require.True(t, errors.Is(err, errcode.ErrObjectNotExist()))
		require.Equal(t, 0, len(fiArr))

		err = baseDb.Delete(&TestVideoInfo{Name: name})
		err = baseDb.List(&fiArr, TestVideoInfo{Name: name})
		require.True(t, errors.Is(err, errcode.ErrObjectNotExist()))
		require.Equal(t, 0, len(fiArr))
	}

	name := utils.RandomString(10)
	doClean(name)
	defer doClean(name)

	tfi1 := TestFileInfo{}
	err = baseDb.FirstOrCreate(&tfi1, TestFileInfo{Path: "2", Name: name})
	assert.Nil(t, err)
	err = baseDb.FirstOrCreate(&TestFileInfo{}, TestFileInfo{Path: "3", Name: name})
	assert.Nil(t, err)
	err = baseDb.FirstOrCreate(&TestFileInfo{}, TestFileInfo{Path: "4", Name: name})
	assert.Nil(t, err)

	assert.Nil(t, baseDb.FirstOrCreate(&TestVideoInfo{}, TestVideoInfo{FId: 1, CloudUserId: 12, Name: name}))
	assert.Nil(t, baseDb.FirstOrCreate(&TestVideoInfo{}, TestVideoInfo{FId: 2, CloudUserId: 12, Name: name}))
	vi := TestVideoInfo{}
	assert.Nil(t, baseDb.FirstOrCreate(&vi, TestVideoInfo{FId: tfi1.FId, CloudUserId: 12, Name: name}))
	assert.Equal(t, vi.FId, tfi1.FId)

	var count int64
	assert.Nil(t, baseDb.GetCount(&count, TestVideoInfo{Name: name}))
	assert.Equal(t, int64(3), count)

	viAtt := []TestVideoInfo{}
	assert.Nil(t, baseDb.List(&viAtt, TestVideoInfo{Name: name}))
	assert.Equal(t, 3, len(viAtt))
	assert.Nil(t, deleteWildRecords(baseDb, &TestVideoInfo{}))

	assert.Nil(t, baseDb.List(&viAtt, TestVideoInfo{Name: name}))
	assert.Equal(t, 1, len(viAtt))

	assert.Nil(t, baseDb.GetCount(&count, TestVideoInfo{Name: name}))
	assert.Equal(t, int64(1), count)
}
