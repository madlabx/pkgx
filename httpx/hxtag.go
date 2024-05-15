package httpx

import (
	"reflect"
	"strings"

	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/log"
)

var (
	tagHttpX             = "hx_tag"
	tagHttpXFieldPlace   = "hx_place"
	tagHttpXFieldName    = "hx_name"
	tagHttpXFieldMust    = "hx_must"
	tagHttpXFieldDefault = "hx_default"
	tagHttpXFieldRange   = "hx_range"
)

const (
	constHxPlaceBody   = "body"
	constHxPlaceQuery  = "query"
	constHxPlaceHeader = "header"
	constHxPlaceEither = ""
)

type hxTag struct {
	place        string
	name         string
	must         bool
	defaultValue string
	valueRange   string
}

func (ht *hxTag) realName(name string) string {
	if ht.name == "" {
		return name
	} else {
		return ht.name
	}
}

func (ht *hxTag) inQuery() bool {
	return ht.place == constHxPlaceQuery || ht.place == constHxPlaceEither
}

func (ht *hxTag) inBody() bool {
	return ht.place == constHxPlaceBody || ht.place == constHxPlaceEither
}

func (ht *hxTag) inHeader() bool {
	return ht.place == constHxPlaceHeader
}

func (ht *hxTag) isEmpty() bool {
	return ht.name == "" &&
		!ht.must &&
		ht.defaultValue == "" &&
		ht.valueRange == ""
}

func parseHxTag(t reflect.StructTag, paths ...string) (*hxTag, error) {
	var (
		place, name, mustStr, defaultValue, valueRange string
		must                                           bool
	)
	tags := t.Get(tagHttpX)
	if len(tags) > 0 {
		tagList := strings.Split(tags, ";")
		if len(tagList) != 5 {
			return nil, errors.Errorf("invalid "+tagHttpX+":'%v' which should have 5 fields, path:%v", tags, paths)
		}
		place, name, mustStr, defaultValue, valueRange = tagList[0], tagList[1], tagList[2], tagList[3], tagList[4]
	} else {
		place, name, mustStr, defaultValue, valueRange =
			t.Get(tagHttpXFieldPlace),
			t.Get(tagHttpXFieldName),
			t.Get(tagHttpXFieldMust),
			t.Get(tagHttpXFieldDefault),
			t.Get(tagHttpXFieldRange)
	}

	if strings.ToLower(mustStr) == "true" {
		must = true
	} else if strings.ToLower(mustStr) == "false" {
		must = false
	} else if len(mustStr) > 0 {
		return nil, errors.Errorf("invalid must tag:%v", mustStr)
	}

	if must && len(defaultValue) > 0 {
		log.Warn("should not define both of hx_must and hx_defaultValue")
	}

	return &hxTag{
		place:        place,
		name:         name,
		must:         must,
		defaultValue: defaultValue,
		valueRange:   valueRange,
	}, nil
}
