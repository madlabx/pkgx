package memkv

import (
	"reflect"
	"strings"
)

func parsePrimaryKey(in any) (string, bool) {
	v := reflect.ValueOf(in)
	t := reflect.TypeOf(in)

	if t.Kind() == reflect.Pointer {
		return parsePrimaryKey(v.Elem().Interface())
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
