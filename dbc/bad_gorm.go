package dbc

import (
	"github.com/madlabx/pkgx/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type badDialector struct{}

func (d *badDialector) Name() string {
	return "BadDialector"
}
func (d *badDialector) Initialize(*gorm.DB) error {
	return errors.New("Err notttt")
}
func (d *badDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return nil
}
func (d *badDialector) DataTypeOf(*schema.Field) string {
	return ""
}
func (d *badDialector) DefaultValueOf(*schema.Field) clause.Expression {
	return nil
}
func (d *badDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {}
func (d *badDialector) Explain(sql string, vars ...interface{}) string {
	return ""
}
func (d *badDialector) QuoteTo(writer clause.Writer, str string) {}

func NewBadGorm(err error) *gorm.DB {
	db, _ := gorm.Open(&badDialector{}, &gorm.Config{})
	db.Error = err
	return db
}

func NewBadDbclient(err error) *DbClient {
	return &DbClient{
		db: NewBadGorm(err),
	}
}
