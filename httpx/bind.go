package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/errors"
	"github.com/madlabx/pkgx/log"
)

var (
	tagHttpX             = "hx_tag"
	tagHttpXFieldPlace   = "hx_place"
	tagHttpXFieldName    = "hx_query_name"
	tagHttpXFieldMust    = "hx_must"
	tagHttpXFieldDefault = "hx_default"
	tagHttpXFieldRange   = "hx_range"
)

const (
	constHxBody   = "body"
	constHxQuery  = "query"
	constHxEither = ""
)

type httpXTag struct {
	place        string
	name         string
	must         bool
	defaultValue string
	valueRange   string
}

func (ht *httpXTag) realName(name string) string {
	if ht.name == "" {
		return name
	} else {
		return ht.name
	}
}

func (ht *httpXTag) inQuery() bool {
	return ht.place == constHxQuery || ht.place == constHxEither
}

func (ht *httpXTag) inBody() bool {
	return ht.place == constHxBody || ht.place == constHxEither
}

func (ht *httpXTag) isEmpty() bool {
	return ht.name == "" &&
		!ht.must &&
		ht.defaultValue == "" &&
		ht.valueRange == ""
}

func parseHttpXDefault(t reflect.StructTag, path string) (string, error) {
	tags := t.Get(tagHttpX)

	//in case of having `hx_tag:";;;default_value;"``
	if len(tags) > 0 {
		tagList := strings.Split(tags, ";")
		if len(tagList) != 5 {
			return "", errors.Errorf("invalid "+tagHttpX+":'%v' which should have 5 fields, path:%v", tags, path)
		}

		return tagList[3], nil
	}

	//in case of having `hx_default:"default_value"`
	return t.Get(tagHttpXFieldDefault), nil
}

func parseHttpXTag(t reflect.StructTag, paths ...string) (*httpXTag, error) {
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

	return &httpXTag{
		place:        place,
		name:         name,
		must:         must,
		defaultValue: defaultValue,
		valueRange:   valueRange,
	}, nil
}

type hxParser struct {
	bodyMap     map[string]any
	bodyParsed  bool
	path        []string
	queryParams url.Values
}

/*
	BindAndValidate 提供从echo.Context里请求里Request的query parameter或者body里取相应的值，赋值给i，并做自校验。使用方式为

	type Trans struct {
		Bandwidth uint64 `hx_place:"query" hx_must:"true" hx_query_name:"bandwidth" hx_default:"default_name" hx_range:"1-2"`
		Loss     float64 `hx_place:"body" hx_must:"false" hx_query_name:"loss" hx_default:"default_name" hx_range:"1.2,3.4"`
	}

	type TusReq struct {
		Name       string `hx_place:"query" hx_must:"true" hx_query_name:"host_name" hx_default:"default_name" hx_range:"alice,bob"`
		TaskId     int64  `hx_place:"body" hx_must:"false" hx_query_name:"task_id" hx_default:"7" hx_range:"0-21"`
		Transfer   Trans
	}

hx_tag自定义如下：

	  easy-to-read style: `hx_place:"query" hx_query_name:"name_in_query" hx_must:"true" hx_default:"def" hx_range:"1-20"
		hx_place: query表示该值从query parameter里取，body从请求body里取。
		hx_query_name： Query Parameters中定义的名称
		hx_must: true表示必须，若未赋值，则报错；false表示可选
		hx_default: 若未赋值，设为该默认值
		hx_range: 根据i的字段的类型来校验range：若为整数，0-21表示0到21是合法的，否则报错；若为字符串，"alice,bob"表示只能为alice或bob，否则报错。

	  compact style:`hx_tag:"f1;f2;f3;f4;f5"`
		f1: same to hx_place
		f2: same to hx_query_name
		f3: same to hx_must
		f4: same to hx_default
		f5: same to hx_range
*/
func BindAndValidate(c echo.Context, i any) error {
	hp := new(hxParser)
	hp.bodyMap = make(map[string]any)
	hp.queryParams = c.QueryParams()

	if c.Request().ContentLength > 0 &&
		strings.HasPrefix(c.Request().Header.Get(echo.HeaderContentType), echo.MIMEApplicationJSON) {
		// Request
		var reqBody []byte
		if c.Request().Body != nil { // Read
			reqBody, _ = io.ReadAll(c.Request().Body)
		}

		c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Reset

		decoder := json.NewDecoder(bytes.NewBuffer(reqBody))
		decoder.UseNumber()
		if err := decoder.Decode(&hp.bodyMap); err != nil {
			return errors.Wrap(err)
		}
	}

	return hp.bindAndValidate(i, hp.bodyMap, []string{}...)
}

// validateStruct recursively validates each field of a struct based on the `httpXTag`.
func validateStruct(c echo.Context, vs reflect.Value, paths ...string) error {
	t := vs.Type()
	switch t.Kind() {
	default:
		return errors.Errorf("invalid type:%v, path:%s.%v", t.Kind(), paths, t.Name())
	case reflect.Invalid:
		return nil
	case reflect.Pointer:
		return validateStruct(c, vs.Elem(), paths...)
	case reflect.Struct:
		break
	}

	for index := 0; index < t.NumField(); index++ {
		field := t.Field(index)
		v := vs.Field(index)

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
			v = v.Elem()
		}

		newPaths := append(paths, field.Name)
		if field.Anonymous ||
			(fieldType.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{})) {
			err := validateStruct(c, v, newPaths...)
			if err != nil {
				return err
			}
			continue
		}

		ht, err := parseHttpXTag(field.Tag, newPaths...)
		if err != nil {
			return errors.Wrap(err)
		}

		if ht.isEmpty() {
			continue
		}

		//获取字段值
		var value string

		name := ht.name
		if name == "" {
			name = field.Name
		}

		switch ht.place {
		case constHxQuery:
			value = c.QueryParam(name)

		case "", constHxBody:
			if v.IsValid() {
				value = fmt.Sprintf("%v", v)
			}

		default:
			return errors.Errorf("invalid "+tagHttpXFieldPlace+" tag %v, path:%v", ht.place, newPaths)
		}

		if value == "" {
			continue
		}

		return validateField(fieldType, field, ht, value, name, newPaths...)
	}
	return nil
}

func validateField(fieldType reflect.Type, field reflect.StructField, ht *httpXTag, value, name string, newPaths ...string) error {

	switch fieldType.Kind() {
	case reflect.String:
		if ht.valueRange != "" {
			allowedValues := strings.Split(ht.valueRange, ",")
			validValue := false
			for _, allowed := range allowedValues {
				if value == allowed {
					validValue = true
					break
				}
			}
			if !validValue {
				return errors.Errorf("invalid value \"%s\" for field %s, must be one of %v, path:%v",
					value, name, allowedValues, newPaths)
			}
		}

	case reflect.Bool:
		_, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Errorf("invalid value \"%s\" for field %s, path:%v", value, name, newPaths)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fieldValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return errors.Errorf("invalid value \"%v\" for field %s, should be an integer, path:%v", value, name, newPaths)
		}

		if ht.valueRange != "" {
			rangeValues := strings.Split(ht.valueRange, "-")
			minVal, e1 := strconv.ParseInt(rangeValues[0], 10, 64)
			maxVal, e2 := strconv.ParseInt(rangeValues[1], 10, 64)
			if e1 != nil || e2 != nil {
				return errors.Errorf("invalid format for "+tagHttpXFieldRange+":%v, path:%v", ht.valueRange, newPaths)
			}
			if fieldValue < minVal || fieldValue > maxVal {
				return errors.Errorf("invalid value \"%s\" for field %s, must be between %d and %d, path:%v",
					value, name, minVal, maxVal, newPaths)
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fieldValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return errors.Errorf("invalid value \"%v\" for field %s, should be an unsigned integer, path:%v",
				value, name, newPaths)
		}
		if ht.valueRange != "" {
			rangeValues := strings.Split(ht.valueRange, "-")
			if len(rangeValues) != 2 {
				return errors.Errorf("invalid "+tagHttpXFieldRange+":%s, path:%v", ht.valueRange, newPaths)
			}
			minVal, e1 := strconv.ParseUint(rangeValues[0], 10, 64)
			maxVal, e2 := strconv.ParseUint(rangeValues[1], 10, 64)
			if e1 != nil || e2 != nil {
				return errors.Errorf("invalid format for "+tagHttpXFieldRange+":%v, path:%v", ht.valueRange, newPaths)
			}
			if fieldValue < minVal || fieldValue > maxVal {
				return errors.Errorf("invalid value \"%s\" for field %s, must be between %d and %d, path:%v", value, name, minVal, maxVal, newPaths)
			}
		}

	case reflect.Float32, reflect.Float64:
		fieldValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return errors.Errorf("invalid value \"%v\" for field %s, should be a float, path:%v",
				value, name, newPaths)
		}
		if ht.valueRange != "" {
			rangeValues := strings.Split(ht.valueRange, "-")
			if len(rangeValues) != 2 {
				return errors.Errorf("invalid value range:%s, path:%v", ht.valueRange, newPaths)
			}
			minVal, e1 := strconv.ParseFloat(rangeValues[0], 64)
			maxVal, e2 := strconv.ParseFloat(rangeValues[1], 64)
			if e1 != nil || e2 != nil {
				return errors.Errorf("invalid format for "+tagHttpXFieldRange+":%v, field %v, path:%v",
					ht.valueRange, field.Name, newPaths)
			}
			if fieldValue < minVal || fieldValue > maxVal {
				return errors.Errorf("invalid value \"%v\" for field %v, must be between %v and %v, path:%v",
					value, name, minVal, maxVal, newPaths)
			}
		}

	case reflect.Struct:
		if field.Type.String() == "time.Time" {
			layout := "2006-01-02T15:04:05"
			timeValue, err := time.Parse(layout, value)
			if err != nil {
				return errors.Errorf("invalid time format for field %s, should be in YYYY-MM-DDTHH:MM:SS format, path:%v",
					name, newPaths)
			}

			if ht.valueRange != "" {
				rangeValues := strings.Split(ht.valueRange, "-")
				minUnix, e1 := strconv.ParseInt(rangeValues[0], 10, 64)
				maxUnix, e2 := strconv.ParseInt(rangeValues[1], 10, 64)
				if e1 != nil || e2 != nil {
					return errors.Errorf("invalid format for "+tagHttpXFieldRange+":%v, path:%v",
						ht.valueRange, newPaths)
				}

				unixTime := timeValue.Unix()
				if unixTime < minUnix || unixTime > maxUnix {
					return errors.Errorf("%s is not within the allowed range for field %s, path:%v",
						value, name, newPaths)
				}
			}
		} else {
			return errors.Errorf("unsupported struct field type:%v for field %s,  path:%v",
				field.Type, name, newPaths)
		}
	default:
		return errors.Errorf("unsupported field type:%v, path:%v", fieldType, newPaths)
	}
	return nil
}

func (hp *hxParser) bindAndValidate(input any, target map[string]any, paths ...string) error {
	v := reflect.ValueOf(input)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.Errorf("invalid type:%v, path:%v", v.Kind(), paths)
	}

	v = v.Elem()
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		if v.Kind() == reflect.Invalid {
			v.Addr().Set(reflect.New(t))
		}
		return hp.bindAndValidate(v.Interface(), target, paths...)
	}

	if t.Kind() != reflect.Struct {
		return errors.Errorf("invalid type:%v, path:%v", v.Kind(), paths)
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		if fieldValue.Kind() == reflect.Invalid {
			fieldValue.Addr().Set(reflect.New(field.Type))
		}

		newPaths := append(paths, field.Name)

		var newTarget any
		if target != nil {
			newTarget = target[field.Name]
		}

		structTarget, _ := newTarget.(map[string]any)

		switch field.Type.Kind() {
		case reflect.Struct:
			err := hp.bindAndValidate(fieldValue.Addr().Interface(), structTarget, newPaths...)
			if err != nil {
				return err
			}
			continue
		case reflect.Ptr:
			if field.Type.Elem().Kind() == reflect.Struct && field.Type.Elem() != reflect.TypeOf(&time.Time{}) {
				if fieldValue.CanSet() {
					fieldValue.Set(reflect.New(fieldValue.Type().Elem()))

					err := hp.bindAndValidate(fieldValue.Addr().Interface(), structTarget, newPaths...)
					if err != nil {
						return err
					}
				}
				continue
			}
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.Struct && field.Type.Elem() != reflect.TypeOf(&time.Time{}) {
				if newTarget != nil {
					sliceTarget, ok := newTarget.([]any)
					if !ok {
						return errors.Errorf("Should be slice %T, path:%v", newTarget, newPaths)
					}

					numElems := len(sliceTarget)
					if numElems > 0 {
						slice := reflect.MakeSlice(field.Type, numElems, numElems)
						for j := 0; j < numElems; j++ {
							structTarget, _ = sliceTarget[j].(map[string]any)
							if err := hp.bindAndValidate(slice.Index(j).Addr().Interface(), structTarget, newPaths...); err != nil {
								return err
							}
						}
						fieldValue.Set(slice)
						continue
					}
				}

				continue

			}
		}

		hxTags, err := parseHttpXTag(field.Tag, newPaths...)
		if err != nil {
			return err
		}

		var (
			value, bv any
			qv        string
		)
		if hxTags.inQuery() {
			qv = hp.queryParams.Get(hxTags.realName(field.Name))
		}

		if hxTags.inBody() {
			bv = target[field.Name]
		}

		if bv != nil {
			value = bv
		} else if qv != "" {
			value = qv
		} else {
			if hxTags.must {
				if hxTags.place != "" {
					return errors.Errorf("missing %s parameter %s", hxTags.place, strings.Join(newPaths, "."))
				} else {
					return errors.Errorf("missing parameter %s", strings.Join(newPaths, "."))
				}
			}
			if hxTags.defaultValue != "" {
				value = hxTags.defaultValue
			}
		}

		if value != nil && fieldValue.CanSet() {
			if err = hp.setFieldAndValidate(fieldValue.Addr(), field.Type, value, hxTags, newPaths...); err != nil {
				return errors.Errorf("%v, path:%v", err, strings.Join(newPaths, "."))
			}
		}
	}

	return nil
}

func (hp *hxParser) existInQueryParam(key string) bool {
	_, exists := hp.queryParams[key]
	if exists {
		return true
	}

	// Go json.Unmarshal supports case insensitive binding.  However the
	// url params are bound case sensitive which is inconsistent.  To
	// fix this we must check all of the map values in a
	// case-insensitive search.
	key = strings.ToLower(key)
	for k, _ := range hp.queryParams {
		if strings.ToLower(k) == key {
			return true
		}
	}

	return false
}

func getMapValue(data map[string]any, path ...string) (any, bool) {
	if len(path) == 0 {
		return "", false
	}

	// 递归遍历map的路径
	cursor := data
	for _, key := range path {
		value, ok := cursor[key]
		if !ok {
			return nil, false
		}
		cursor, ok = value.(map[string]any)
		if !ok {
			return value, true
		}
	}

	return nil, false
}

func isMapValueExist(data map[string]interface{}, path ...string) bool {
	if len(path) == 0 {
		return false
	}

	// 递归遍历map的路径
	cursor := data
	for _, key := range path {
		value, ok := cursor[key]
		if !ok {
			return false
		}
		cursor, ok = value.(map[string]interface{})
		if !ok {
			return true
		}
	}

	// 检查最终路径对应的值是否存在
	return reflect.ValueOf(cursor).IsValid()
}

func setIntField(value string, bitSize int, field reflect.Value, ht *httpXTag) error {

	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err != nil {
		return err
	}

	if ht.valueRange != "" {
		rangeValues := strings.Split(ht.valueRange, "-")
		minVal, e1 := strconv.ParseInt(rangeValues[0], 10, bitSize)
		maxVal, e2 := strconv.ParseInt(rangeValues[1], 10, bitSize)
		if e1 != nil || e2 != nil {
			return errors.Errorf("invalid format for "+tagHttpXFieldRange+":%v", ht.valueRange)
		}
		if intVal < minVal || intVal > maxVal {
			return errors.Errorf("invalid value \"%s\", must be between %d and %d",
				value, minVal, maxVal)
		}
	}

	field.SetInt(intVal)
	return nil
}

func setUintField(value string, bitSize int, field reflect.Value, ht *httpXTag) error {
	if value == "" {
		value = "0"
	}
	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil {
		return err
	}

	if ht.valueRange != "" {
		rangeValues := strings.Split(ht.valueRange, "-")
		if len(rangeValues) != 2 {
			return errors.Errorf("invalid "+tagHttpXFieldRange+":%s", ht.valueRange)
		}
		minVal, e1 := strconv.ParseUint(rangeValues[0], 10, bitSize)
		maxVal, e2 := strconv.ParseUint(rangeValues[1], 10, bitSize)
		if e1 != nil || e2 != nil {
			return errors.Errorf("invalid format for "+tagHttpXFieldRange+":%v", ht.valueRange)
		}
		if uintVal < minVal || uintVal > maxVal {
			return errors.Errorf("invalid value \"%s\", must be between %d and %d", value, minVal, maxVal)
		}
	}

	field.SetUint(uintVal)

	return nil
}

func setBoolField(value string, field reflect.Value, ht *httpXTag) error {
	if value == "" {
		value = "false"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value, ht *httpXTag) error {
	if value == "" {
		value = "0.0"
	}
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err != nil {
		return err
	}

	if ht.valueRange != "" {
		rangeValues := strings.Split(ht.valueRange, "-")
		if len(rangeValues) != 2 {
			return errors.Errorf("invalid value range:%s", ht.valueRange)
		}
		minVal, e1 := strconv.ParseFloat(rangeValues[0], bitSize)
		maxVal, e2 := strconv.ParseFloat(rangeValues[1], bitSize)
		if e1 != nil || e2 != nil {
			return errors.Errorf("invalid format for "+tagHttpXFieldRange+":%v",
				ht.valueRange)
		}
		if floatVal < minVal || floatVal > maxVal {
			return errors.Errorf("invalid value \"%v\" , must be between %v and %v",
				value, minVal, maxVal)
		}
	}

	field.SetFloat(floatVal)
	return nil
}

func setStringField(value string, field reflect.Value, ht *httpXTag) error {
	if ht.valueRange != "" {
		allowedValues := strings.Split(ht.valueRange, ",")
		validValue := false
		for _, allowed := range allowedValues {
			if value == allowed {
				validValue = true
				break
			}
		}
		if !validValue {
			return errors.Errorf("invalid value \"%s\" must be one of %v",
				value, allowedValues)
		}
	}
	field.SetString(value)
	return nil
}

func setValue(t reflect.Type, val reflect.Value, v reflect.Value) error {
	if v.Elem().Kind() == reflect.Invalid {
		v.Set(reflect.New(t))
	}

	structField := v.Elem()

	//TODO value time.Time
	switch t.Kind() {
	case reflect.Ptr:
		return setValue(t.Elem(), val, structField)
	default:
		structField.Set(reflect.ValueOf(val))
	}
	return nil
}

// setFieldAndValidate sets a struct field with a value, ensuring it is of the proper type.
func (hp *hxParser) setFieldAndValidate(structFieldPtr reflect.Value, objT reflect.Type, oriV any, ht *httpXTag, paths ...string) error {
	if structFieldPtr.Elem().Kind() == reflect.Invalid {
		structFieldPtr.Set(reflect.New(objT))
	}

	structField := structFieldPtr.Elem()

	if objT.Kind() == reflect.Slice {
		if reflect.TypeOf(oriV).Kind() != reflect.Slice {
			return errors.Errorf("unmatch type, field type:%v, value:%v, type of value:%T", objT.Kind(), oriV, oriV)
		}

		val := oriV.([]any)

		numElems := len(val)
		if numElems >= 0 {
			sliceOf := structField.Type().Elem()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for j := 0; j < numElems; j++ {
				if err := hp.setFieldAndValidate(slice.Index(j).Addr(), sliceOf, val[j], ht, paths...); err != nil {
					return err
				}
			}
			structField.Set(slice)
			return nil
		}
	}

	val := fmt.Sprintf("%v", oriV)
	//TODO value time.Time
	switch objT.Kind() {
	case reflect.Pointer:
		return hp.setFieldAndValidate(structField, objT.Elem(), val, ht, paths...)
	case reflect.Int:
		return setIntField(val, 0, structField, ht)
	case reflect.Int8:
		return setIntField(val, 8, structField, ht)
	case reflect.Int16:
		return setIntField(val, 16, structField, ht)
	case reflect.Int32:
		return setIntField(val, 32, structField, ht)
	case reflect.Int64:
		return setIntField(val, 64, structField, ht)
	case reflect.Uint:
		return setUintField(val, 0, structField, ht)
	case reflect.Uint8:
		return setUintField(val, 8, structField, ht)
	case reflect.Uint16:
		return setUintField(val, 16, structField, ht)
	case reflect.Uint32:
		return setUintField(val, 32, structField, ht)
	case reflect.Uint64:
		return setUintField(val, 64, structField, ht)
	case reflect.Bool:
		return setBoolField(val, structField, ht)
	case reflect.Float32:
		return setFloatField(val, 32, structField, ht)
	case reflect.Float64:
		return setFloatField(val, 64, structField, ht)
	case reflect.String:
		return setStringField(val, structField, ht)
	//case reflect.Struct:
	//	return hp.bindAndValidate(structField.Addr().Interface(), paths...)
	default:
		return errors.Errorf("invalid type:%v, value:%#v, value type:%T", objT.Kind(), oriV, oriV)
	}
}
