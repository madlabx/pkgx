package dbc_test

import (
	"testing"

	"github.com/madlabx/pkgx/dbc"
	"github.com/madlabx/pkgx/errors"
	"github.com/stretchr/testify/assert"
)

func TestBadGorm(t *testing.T) {
	err := errors.New("test error")

	db := dbc.NewBadGorm(err)
	assert.Equal(t, err, db.Error)

	result := make(map[string]any)
	if err := db.Table("not exist table").Where("path = ?", "test").First(&result).Error; err != nil {
		assert.Equal(t, err, db.Error)
	} else {
		t.Fatalf("should be error")
	}

	var count int64
	if err := db.Table("not exist table").Where("path = ?", "test").Count(&count).Error; err != nil {
		assert.Equal(t, err, db.Error)
	} else {
		t.Fatalf("should be error")
	}

	if err := db.Table("not exist table").Where("path = ?", "test").Update("path", "new").Error; err != nil {
		assert.Equal(t, err, db.Error)
	} else {
		t.Fatalf("should be error")
	}

	if err := db.Table("not exist table").Where("path = ?", "test").Create(result).Error; err != nil {
		assert.Equal(t, err, db.Error)
	} else {
		t.Fatalf("should be error")
	}

	if err := db.Exec("select 1").Error; err != nil {
		assert.Equal(t, err, db.Error)
	} else {
		t.Fatalf("should be error")
	}
}
