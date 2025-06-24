package dbc

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/madlabx/pkgx/errcode"
	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/log"
	"github.com/madlabx/pkgx/memkv"
	"github.com/madlabx/pkgx/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var _ memkv.KvDbClientIf = (*DbClient)(nil)

type SqlStructIf interface {
	TableName() string
}

type DbClient struct {
	config *gorm.Config
	dsn    string
	db     *gorm.DB

	// ext db fields
	extDbPrefix   string
	initCompleted bool
}

type LogContent struct {
	SlowThreshold             int64 `vx_default:"10000"` //in ns
	Colorful                  bool  `vx_default:"true"`
	IgnoreRecordNotFoundError bool  `vx_default:"false"`
	ParameterizedQueries      bool  `vx_default:"false"`
}

type LogConfig struct {
	LogFile log.FileConfig
	Level   string `vx_default:"info"`
}

func (lc LogConfig) ToSqlLogLevel() (gormlogger.LogLevel, error) {
	l, err := logrus.ParseLevel(lc.Level)
	if err != nil {
		return -1, errors.Wrap(err)
	}

	switch l {
	default:
		return -1, errors.New("Invalid log level")
	case logrus.PanicLevel, logrus.FatalLevel:
		return gormlogger.Silent, nil
	case logrus.ErrorLevel:

		return gormlogger.Error, nil
	case logrus.WarnLevel:

		return gormlogger.Warn, nil
	case logrus.InfoLevel, logrus.DebugLevel, logrus.TraceLevel:
		return gormlogger.Info, nil
	}
}

type SqlConfig struct {
	Type         string `vx_range:"oneof=psql mysql sqllite" vx_must:"true"`
	Host         string `json:"-"`
	Port         string `json:"-"`
	User         string `json:"-"`
	Password     string `json:"-"`
	Dbname       string `json:"-" vx_must:"true"`
	Log          LogConfig
	LogContent   LogContent
	MaxIdleConns int    `vx_default:"10"`
	MaxOpenConns int    `vx_default:"100"`
	ExtPrefix    string `vx_default:""`
	ExtDbDir     string `vx_default:"/app/workspace/file_server/ext_db"`
}

func (c *SqlConfig) IsExt() bool {
	return c.ExtPrefix != ""
}

func NewTestDbClient(db *gorm.DB) *DbClient {
	return &DbClient{
		db: db,
	}
}

const (
	ConstDbTypePsql    string = "psql"
	ConstDbTypeMysql   string = "mysql"
	ConstDbTypeSqlLite string = "sqllite"
)

// isDatabaseNotExistError 检查错误是否指示数据库不存在
func isDatabaseNotExistError(err error, dbName string) bool {
	return strings.Contains(err.Error(), dbName+"\" does not exist") || strings.Contains(err.Error(), "Unknown database '"+dbName)
}

func openRawDb(conf SqlConfig) (*gorm.DB, error) {

	var dialector gorm.Dialector
	var dsn string
	switch conf.Type {
	default:
		return nil, errors.Errorf("not support to dropDb for db type:[%v]", conf.Type)
	case ConstDbTypePsql:
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s port=%s sslmode=disable TimeZone=Asia/Shanghai client_encoding=utf8",
			conf.Host, conf.User, conf.Password, conf.Port)
		dialector = postgres.Open(dsn)
	case ConstDbTypeMysql:
		//"user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
		dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
			conf.User, conf.Password, conf.Host, conf.Port)
		dialector = mysql.Open(dsn)
	}

	return gorm.Open(dialector, &gorm.Config{})
}

func dropDb(conf SqlConfig) error {
	db, err := openRawDb(conf)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to db")
	}

	err = db.Exec(fmt.Sprintf("DROP DATABASE %s", conf.Dbname)).Error
	if err != nil {
		return errors.Wrapf(err, "failed to create database")
	}

	return nil
}

func createDb(conf SqlConfig) error {
	db, err := openRawDb(conf)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to db")
	}

	err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", conf.Dbname)).Error
	if err != nil {
		return errors.Wrapf(err, "failed to create database")
	}

	return nil
}

func NewDbClient(pCtx context.Context, conf SqlConfig, tables ...any) (*DbClient, error) {
	var dialector gorm.Dialector
	newDbC := &DbClient{}
	switch conf.Type {
	default:
		return nil, errors.Errorf("wrong db type:[%v]", conf.Type)
	case ConstDbTypePsql:
		//"user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
		newDbC.dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai client_encoding=utf8",
			conf.Host, conf.User, conf.Password, conf.Dbname, conf.Port)

		dialector = postgres.Open(newDbC.dsn)
	case ConstDbTypeMysql:
		//"user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
		newDbC.dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			conf.User, conf.Password, conf.Host, conf.Port, conf.Dbname)

		dialector = mysql.Open(newDbC.dsn)

	case ConstDbTypeSqlLite:
		newDbC.dsn = fmt.Sprintf("file:%s", conf.Dbname)
		dialector = sqlite.Open(newDbC.dsn)
	}

	lg := log.SetLoggerOutput(nil, pCtx, conf.Log.LogFile)
	log.SetLoggerFormatter(lg, &log.TextFormatter{
		QuoteEmptyFields: true,
		SkipFrames:       1,
		DisableSorting:   true})

	isTerminal := func(w io.Writer) bool {
		switch v := w.(type) {
		case *os.File:
			return terminal.IsTerminal(int(v.Fd()))
		default:
			return false
		}
	}(lg.Out)

	loglevel, err := conf.Log.ToSqlLogLevel()
	if err != nil {
		return nil, err
	}

	newDbC.config = &gorm.Config{
		Logger: gormlogger.New(
			lg,
			gormlogger.Config{
				SlowThreshold:             time.Duration(conf.LogContent.SlowThreshold), // 慢查询阈值，单位为毫秒
				LogLevel:                  loglevel,                                     // 日志级别
				Colorful:                  conf.LogContent.Colorful && isTerminal,       // 禁用控制台日志的颜色
				IgnoreRecordNotFoundError: conf.LogContent.IgnoreRecordNotFoundError,    // 忽略记录未找到的错误
				ParameterizedQueries:      conf.LogContent.ParameterizedQueries,
			},
		),
	}

	// 使用GORM连接数据库

	connectOrCreateDb := func() (*gorm.DB, error) {
		db, err := gorm.Open(dialector, newDbC.config)
		if err != nil {
			if isDatabaseNotExistError(err, conf.Dbname) {
				log.Infof("database not exist, try to create")
				if err1 := createDb(conf); err1 != nil {
					log.Errorf("failed to createDb, err:%v", err1)
				} else {
					db, err = gorm.Open(dialector, newDbC.config)
				}
			}
		}
		return db, err
	}

	db, err := connectOrCreateDb()
	if err != nil {
		log.Errorf("failed to connect db, err:%v", err)
		ctx, _ := context.WithCancel(pCtx)
		timer := time.NewTimer(time.Second)
		for {
			select {
			case <-ctx.Done():
				log.Infof("context done, return")
				return nil, errors.New("failed to connect db for context done")
			case <-timer.C:
				if db, err = connectOrCreateDb(); err != nil {
					log.Errorf("failed to connect db, err:%v", err)
				}
				timer.Reset(time.Second)
			}
		}
	}

	if conf.Type == ConstDbTypeSqlLite {
		if conf.IsExt() {
			// https://alidocs.dingtalk.com/i/nodes/XPwkYGxZV3y4NARwi0QN4QaX8AgozOKL?corpId=dingff6a1f72f162ef7235c2f4657eb6378f&utm_medium=im_card&cid=1566763%3A3737526235&iframeQuery=utm_medium%3Dim_card%26utm_source%3Dim&utm_scene=team_space&utm_source=im
			// 不追求过高的可靠性，使用 WAL 模式
			// 由 scan、startLoopToCleanInvalidRecords 来确保数据最终一致性
			err = db.Exec("PRAGMA journal_mode = WAL").Error
			if err != nil {
				log.Errorf("Failed to run special action for SqlLite, err:%v", err)
			}
			newDbC.extDbPrefix = conf.ExtPrefix
		} else {
			//保证强同步
			//TODO 可能会加剧db lock问题，后面结合WAL做测试
			err = db.Exec("PRAGMA synchronous = FULL").Error
			if err != nil {
				log.Errorf("Failed to run special action for SqlLite, err:%v", err)
			}

			// sqlite 改为 1 避免 db lock，ext 采用了 WAL 模式，所以不限制单线程操作
			conf.MaxOpenConns = 1
			conf.MaxIdleConns = 1
		}

		//if err == nil {
		//	err = db.Exec("PRAGMA locking_mode = EXCLUSIVE").Error // 设置锁定模式
		//}
		//if err == nil {
		//	err = db.Exec("PRAGMA journal_mode = WAL").Error // 使用 Write-Ahead Logging 模式
		//} else {
		//	log.Errorf("Failed to run special action for SqlLite, err:%v", err)
		//}
		//TODO add retry for sql busy
	}

	rdb, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	rdb.SetMaxIdleConns(conf.MaxIdleConns)
	rdb.SetMaxOpenConns(conf.MaxOpenConns)
	newDbC.db = db

	if err = newDbC.initTable(tables...); err != nil {
		return nil, errors.Wrap(err)
	}

	return newDbC, nil
}

func (c *DbClient) DB() *gorm.DB {
	return c.db
}

func (c *DbClient) RawCmd(sql string) ([]map[string]any, error) {
	rows, err := c.db.Raw(sql).Rows()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		for i := range columns {
			values[i] = new(any)
		}
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		m := make(map[string]any)
		for i, colName := range columns {
			val := *(values[i].(*any))
			switch v := val.(type) {
			case []byte:
				m[colName] = string(v)
			default:
				m[colName] = v
			}
		}
		result = append(result, m)
	}
	return result, nil
}

func (c *DbClient) initTable(ts ...any) error {
	if err := c.db.AutoMigrate(ts...); err != nil {
		return err
	}
	for _, t := range ts {
		if to := reflect.TypeOf(t); to.Kind() != reflect.Ptr { // 如果不是指针，则不允许操作
			return errcode.ErrInternalServerError().WithErrorf("table should be pointer")
		}
		u, ok := t.(UpgradeTable)
		if ok {
			if c.db.Migrator().HasTable(u.OldTableName()) {
				if err := u.Upgrade(c); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *DbClient) InitCompleted() {
	c.initCompleted = true
}

func (c *DbClient) IsInitCompleted() bool {
	return c.initCompleted
}

type OrCondition struct {
	Wheres []WhereCondition
}

func (o *OrCondition) Or(w WhereCondition) {
	o.Wheres = append(o.Wheres, w)
}

func (o *OrCondition) ToWhere() *WhereCondition {
	return &WhereCondition{
		Query: o,
	}
}

func (o *OrCondition) Apply(db *gorm.DB) *gorm.DB {
	if len(o.Wheres) == 0 {
		return db
	}
	db = db.Where(o.Wheres[0].Query, o.Wheres[0].Args)
	for _, w := range o.Wheres[1:] {
		db = db.Or(w.Query, w.Args)
	}
	return db
}

type WhereCondition struct {
	Query any
	Args  any
}

func (w *WhereCondition) Valid() bool {
	return w.Query != ""
}

func (w *WhereCondition) IsOr() bool {
	_, ok := w.Query.(*OrCondition)
	return ok
}

type OrderCondition struct {
	Field string
	Order string
}

func (o *OrderCondition) Valid() bool {
	return o.Field != "" && o.Order != ""
}

func (o *OrderCondition) ToString() string {
	return utils.ToSnakeString(o.Field) + " " + o.Order
}

// GetArrayCondition retrieves a paginated list of records from the database and also returns the total count of records matching the conditions.
// The conditions parameter can be a string with arguments, or a map or struct used to build the WHERE clause.
// It returns the total record count and an error if the operation fails.
func (c *DbClient) GetArrayCondition(records any,
	whereConditions []WhereCondition, orderConditions []OrderCondition,
	pageSize int, pageNum int,
) (int64, error) {
	var count int64
	res := c.db.Model(records)

	// Apply WHERE conditions
	for _, whereCondition := range whereConditions {
		if whereCondition.Valid() {
			res = res.Where(whereCondition.Query, whereCondition.Args)
		}
	}

	// Summary the total number of records matching the conditions
	if err := res.Count(&count).Error; err != nil {
		return 0, err
	}

	for _, cdt := range orderConditions {
		// Apply ORDER BY if an order string is present
		if cdt.Valid() {
			res = res.Order(cdt.ToString())
		}
	}

	if pageSize > 0 && pageNum > 0 {
		// Calculate the number of records to skip (offset)
		offset := (pageNum - 1) * pageSize
		res = res.Offset(offset).Limit(pageSize)
	}

	// Retrieve the paginated records
	if err := res.Find(records).Error; err != nil {
		return 0, err
	}

	// Return the total record count and nil error on success
	return count, nil
}

func (c *DbClient) Save(records any) error {
	return errors.Wrap(c.db.Save(records).Error)
}

func (c *DbClient) Delete(records any) error {
	rst := c.db.Delete(records, records)
	if rst.RowsAffected == 0 {
		return errcode.ErrObjectNotExist()
	}

	return errors.Wrap(rst.Error)
}

func (c *DbClient) List(dest any, cond any) error {
	if reflect.TypeOf(dest).Kind() != reflect.Pointer {
		return errors.New("dest should be pointer")
	}

	if reflect.TypeOf(dest).Elem().Kind() != reflect.Slice {
		return errors.New("dest should be pointer to slice")
	}

	err := c.db.Find(dest, cond).Error
	if err != nil {
		return errors.Wrap(err)
	}

	array := reflect.ValueOf(dest).Elem()
	if array.Kind() != reflect.Slice {
		return errors.New("dest should be pointer to slice")
	}
	if array.Len() == 0 {
		return errcode.ErrObjectNotExist()
	}

	return nil
}

func (c *DbClient) ListWithKeyPrefix(dest any, filter any, keyFieldName, keyPrefix string) error {
	if reflect.TypeOf(dest).Kind() != reflect.Pointer {
		return errors.New("dest should be pointer")
	}

	if reflect.TypeOf(dest).Elem().Kind() != reflect.Slice {
		return errors.New("dest should be pointer to slice")
	}

	err := c.db.Model(filter).Find(dest, keyFieldName+" like ?", keyPrefix+"%").Error
	if err != nil {
		return errors.Wrap(err)
	}
	//err = c.db.Find(dest, keyFieldName+" like ?", keyPrefix+"%").Error
	//if err != nil {
	//	return errors.Wrap(err)
	//}

	//err = c.db.Table(filter.TableName()).Where(keyFieldName+" like ?", keyPrefix+"%").Find(dest).Error

	if err != nil {
		return errors.Wrap(err)
	}

	array := reflect.ValueOf(dest).Elem()
	if array.Kind() != reflect.Slice {
		return errors.New("dest should be pointer to slice")
	}
	if array.Len() == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func ParsePrimaryKey(in any) (string, bool) {
	v := reflect.ValueOf(in)
	t := reflect.TypeOf(in)

	if t.Kind() == reflect.Pointer {
		return ParsePrimaryKey(v.Elem().Interface())
	}

	var field reflect.StructField
	foundPrimaryKey := false
	primaryKeyName := ""
	for i := 0; i < t.NumField(); i++ {
		field = t.Field(i)
		// 分割标签来找到primaryKey
		tags := strings.Split(field.Tag.Get("gorm"), ";")
		for _, tag := range tags {
			innerTags := strings.Split(tag, ":")
			switch innerTags[0] {
			default:
			//do nothing
			case "primaryKey":
				foundPrimaryKey = true
			case "column":
				primaryKeyName = innerTags[1]
			}
		}

		if foundPrimaryKey {
			return primaryKeyName, foundPrimaryKey
		}
	}

	return "", false
}

func (c *DbClient) GetByPrimary(dest any, id any) error {
	err := c.db.First(dest, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrObjectNotExist()
	}
	return errors.Wrap(err)
}

func (c *DbClient) ListByPrimaryKeys(dest any, keys any) error {
	if reflect.TypeOf(keys).Kind() != reflect.Slice {
		return errors.New("keys should be slice")
	}
	return c.db.Find(dest, keys).Error
}

// 获取filterWithAttr中的非零成员作为
func (c *DbClient) ListWithOneAttr(dest any, modelWithOneColumn any, attrValue any) error {

	attrName, err := GormColumn(modelWithOneColumn)
	if err != nil {
		return err
	}

	rst := c.db.Where(attrName+" in ?", attrValue).Find(dest)
	if rst.Error != nil {
		return errors.Wrap(rst.Error)
	}
	if rst.RowsAffected == 0 {
		return errcode.ErrObjectNotExist()
	}

	return nil
}

func (c *DbClient) First(dest any, conds ...any) error {
	return c.db.First(dest, conds...).Error
}

func (c *DbClient) Last(dest any, conds ...any) error {
	return c.db.Last(dest, conds...).Error
}

func (c *DbClient) Update(dest any, column string, value any) error {
	// Model(dest) to let gorm know which table to operate
	return c.db.Model(dest).Update(column, value).Error
}

func (c *DbClient) DeleteByPrimaryKeys(dest any, keys any) error {
	if reflect.TypeOf(keys).Kind() != reflect.Slice {
		return errors.New("keys should be slice")
	}
	return c.db.Delete(dest, keys).Error
}

// db.Table("users").Select("users.*, profiles.*")
// FirstInJoinQuery supports join ops like
// .Joins("LEFT JOIN profiles ON users.id = profiles.user_id")
// .Where("users.username = ?", "john")
// .Find(&usersWithProfiles) or .Take(&usersWithProfiles)
func (c *DbClient) FirstInJoinQuery(table, fields string,
	joins []string, dest any, conds ...any) error {
	chainedDB := c.db.Table(table).Select(fields)
	for _, join := range joins {
		chainedDB.Joins(join)
	}
	return chainedDB.Take(dest, conds...).Error
}

// 注意： 如果newAttrs为结构体，会忽略0值, 比如EnableSamba: false
func (c *DbClient) UpdatesOmitZero(cond any, newAttrs any) error {
	return errors.Wrap(c.db.Model(cond).Where(cond).Updates(newAttrs).Error)
}

func (c *DbClient) UpdatesWithZero(cond any, newAttrs any) error {
	if reflect.TypeOf(newAttrs).Kind() != reflect.Map {
		return errcode.ErrNotImplemented().WithErrorf("Only accept map")
	}
	return errors.Wrap(c.db.Model(cond).Where(cond).Updates(newAttrs).Error)
}

func (c *DbClient) ReActivate(in any) error {
	return c.db.Model(in).Updates(map[string]interface{}{"delete_at": nil}).Error
}

// WARMING: zero in filterAndDest is omited
func (c *DbClient) FindUniq(filterAndDest any) error {
	err := c.Where(filterAndDest).First(filterAndDest)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrObjectNotExist()
	}
	return errors.Wrap(err)
}

func (c *DbClient) Updates(dest any, values any, conds ...any) error {
	if len(conds) < 2 {
		return errors.New("Update conditions invalid")
	}
	return c.db.Model(dest).Where(conds[0], conds[1:]...).Updates(values).Error
}

func (c *DbClient) clone() *DbClient {
	n := *c
	return &n
}

func (c *DbClient) FirstOrCreate(dest any, cond any) error {
	return c.db.FirstOrCreate(dest, cond).Error
}
func (c *DbClient) Where(query any, args ...any) *DbClient {
	clo := c.clone()
	clo.db = clo.db.Where(query, args...)
	return clo
}

func (c *DbClient) GetCount(count *int64, filter any) error {
	return errors.Wrap(c.db.Model(filter).Count(count).Error)
}

// enable chainning operations
type Accessor interface {
	Count(count *int64) *gorm.DB
	Model(value any) *gorm.DB
	Joins(query string, args ...any) *gorm.DB
	Table(name string, args ...any) *gorm.DB
	Where(query any, args ...any) *gorm.DB
	Limit(limit int) *gorm.DB
	Offset(offset int) *gorm.DB
	Find(dest any, conds ...any) *gorm.DB
	Select(query any, args ...any) *gorm.DB
	Order(value any) *gorm.DB
	First(dest any, conds ...any) *gorm.DB
}

func (c *DbClient) Count(count *int64) Accessor {
	return c.db.Count(count)
}

func (c *DbClient) Model(value any) Accessor {
	return c.db.Model(value)
}
func (c *DbClient) Joins(query string, args ...any) Accessor {
	return c.db.Joins(query, args...)
}
func (c *DbClient) Table(name string, args ...any) Accessor {
	return c.db.Table(name, args...)
}

func (c *DbClient) Limit(count int) Accessor {
	return c.db.Limit(count)
}
func (c *DbClient) Offset(offset int) Accessor {
	return c.db.Offset(offset)
}

func (c *DbClient) Select(query any, args ...any) Accessor {
	return c.db.Select(query, args...)
}

func (c *DbClient) Order(value any) Accessor {
	return c.db.Order(value)
}

func (c *DbClient) DeleteExpired(filter any) error {
	return c.db.Delete(filter, "expire_at < ?", time.Now().Unix()).Error
}
func (c *DbClient) Set(input any) error {
	return c.Save(input)
}

func (c *DbClient) Get(filterAndDest any) error {
	err := c.Where(filterAndDest).First(filterAndDest)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrObjectNotExist()
	}
	return err
}

func GormColumn(record interface{}) (string, error) {
	// 获取record的反射值
	recv := reflect.ValueOf(record)

	// 确保record是一个结构体
	if recv.Kind() != reflect.Struct {
		return "", errors.Errorf("expected a struct")
	}

	// 遍历结构体的所有字段
	attryFieldName := ""
	columnName := ""
	for i := 0; i < recv.NumField(); i++ {
		field := recv.Field(i)

		// 跳过空值字段
		if field.IsZero() {
			continue
		}

		if attryFieldName != "" {
			return "", errors.Errorf("multiple non-empty fields found: %v, %v", attryFieldName, recv.Type().Field(i).Name)
		}

		attryFieldName = recv.Type().Field(i).Name
		// 获取字段的gorm标签
		gormTag := recv.Type().Field(i).Tag.Get("gorm")

		// 解析gorm标签以获取列名
		columnName = parseGormTag(gormTag)
		if columnName == "" {
			columnName = utils.ToSnakeString(attryFieldName)
		}
	}

	// 返回第一个非空字段的列名
	return columnName, nil
}

// parseGormTag 解析gorm标签并返回列名
func parseGormTag(tag string) string {
	// 使用空格分割标签内容
	parts := strings.Split(tag, ";")

	// 遍历标签的各个部分，找到以"column:"开头的部分
	for _, part := range parts {
		if strings.HasPrefix(part, "column:") {
			// 返回列名，即"column:"后面的内容
			return strings.TrimPrefix(part, "column:")
		}
	}

	// 如果没有找到"column:"，则返回空字符串
	return ""
}
