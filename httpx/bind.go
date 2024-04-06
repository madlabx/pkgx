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

	"github.com/madlabx/pkgx/log"

	"github.com/madlabx/pkgx/typex"

	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/errors"
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
	ConstHxPlaceBody   = "body"
	ConstHxPlaceQuery  = "query"
	ConstHxPlaceEither = ""
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
	return ht.place == ConstHxPlaceQuery || ht.place == ConstHxPlaceEither
}

func (ht *httpXTag) inBody() bool {
	return ht.place == ConstHxPlaceBody || ht.place == ConstHxPlaceEither
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
	bodyMap     typex.JsonMap
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
	hp.bodyMap = make(typex.JsonMap)
	hp.queryParams = c.QueryParams()
	err := hp.setHttpXDefaultAndCheckMust(c, i, []string{}...)
	if err != nil {
		return err
	}
	//
	//if err = c.Bind(i); err != nil &&
	//	!strings.Contains(err.Error(), "Request body can't be empty") &&
	//	!errors.Is(err, echo.ErrUnsupportedMediaType) {
	//	return err
	//}

	return validate(c, reflect.ValueOf(i), []string{}...)
}

// validate recursively validates each field of a struct based on the `httpXTag`.
func validate(c echo.Context, vs reflect.Value, paths ...string) error {
	t := vs.Type()
	switch t.Kind() {
	default:
		return errors.Errorf("invalid type:%v, path:%s.%v", t.Kind(), paths, t.Name())
	case reflect.Invalid:
		return nil
	case reflect.Pointer:
		return validate(c, vs.Elem(), paths...)
	case reflect.Struct:
		break
	}

	for index := 0; index < t.NumField(); index++ {
		field := t.Field(index)
		v := vs.Field(index)

		kind := field.Type.Kind()
		if field.Type.Kind() == reflect.Ptr {
			kind = reflect.TypeOf(v.Interface()).Elem().Kind()
			v = v.Elem()
		}

		newPaths := append(paths, field.Name)
		if field.Anonymous ||
			(field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{})) {
			err := validate(c, v, newPaths...)
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
		case ConstHxPlaceQuery:
			value = c.QueryParam(name)
			if value == "" {
				value = c.QueryParam(strings.ToLower(name))
			}

			//if ht.must && value == "" {
			//	return errors.Errorf("missing query param %v", name)
			//}
		case "", ConstHxPlaceBody:
			if v.IsValid() && v.CanInterface() {
				value = fmt.Sprintf("%v", v.Interface())
			}

			//if ht.must && value == "" {
			//	return errors.Errorf("missing body paramm %v", newPaths)
			//}
		default:
			return errors.Errorf("invalid "+tagHttpXFieldPlace+" tag %v, path:%v", ht.place, newPaths)
		}

		if value == "" {
			continue
		}

		switch kind {
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
			return errors.Errorf("unsupported field type:%v, path:%v", reflect.TypeOf(v), newPaths)
		}
	}
	return nil
}

func (hp *hxParser) setHttpXDefaultAndCheckMust(c echo.Context, input any, paths ...string) error {
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
		return hp.setHttpXDefaultAndCheckMust(c, v.Interface(), paths...)
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
		if field.Type.Kind() == reflect.Struct ||
			(field.Type.Kind() == reflect.Ptr &&
				field.Type.Elem().Kind() == reflect.Struct &&
				field.Type != reflect.TypeOf(&time.Time{})) {

			if fieldValue.CanSet() {
				if fieldValue.Kind() == reflect.Ptr {
					fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
				}

				err := hp.setHttpXDefaultAndCheckMust(c, fieldValue.Addr().Interface(), newPaths...)
				if err != nil {
					return err
				}
			}
			continue
		}

		hxTags, err := parseHttpXTag(field.Tag, newPaths...)
		if err != nil {
			return err
		}

		var (
			value string
			qv    string
			bv    any
		)
		if hxTags.inQuery() {
			qv = c.QueryParam(hxTags.realName(field.Name))
		}

		if hxTags.inBody() {
			if !hp.bodyParsed {
				hp.bodyParsed = true
				if c.Request().ContentLength > 0 &&
					strings.HasPrefix(c.Request().Header.Get(echo.HeaderContentType), echo.MIMEApplicationJSON) {
					// Request
					var reqBody []byte
					if c.Request().Body != nil { // Read
						reqBody, _ = io.ReadAll(c.Request().Body)
					}
					c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Reset
					if err = json.Unmarshal(reqBody, &hp.bodyMap); err != nil {
						return errors.Wrap(err)
					}
				}
			}
			bv, _ = getMapValue(hp.bodyMap, newPaths...)
		}

		switch {
		case bv != nil:
			value = fmt.Sprintf("%v", bv)
		case qv != "":
			value = qv
		default:
			if hxTags.must {
				if hxTags.place != "" {
					return errors.Errorf("missing %s parameter %s", hxTags.place, strings.Join(newPaths, "."))
				} else {
					return errors.Errorf("missing parameter %s", strings.Join(newPaths, "."))
				}
			}
			value = hxTags.defaultValue
		}

		if value != "" && fieldValue.CanSet() {
			if err = setWithProperType(field.Type, value, fieldValue.Addr()); err != nil {
				return errors.Wrapf(err, "path:%v", newPaths)
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

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		value = "false"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0.0"
	}
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
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

// setWithProperType sets a struct field with a value, ensuring it is of the proper type.
func setWithProperType(t reflect.Type, val string, v reflect.Value) error {
	if v.Elem().Kind() == reflect.Invalid {
		v.Set(reflect.New(t))
	}
	structField := v.Elem()

	//TODO value time.Time
	switch t.Kind() {
	case reflect.Ptr:
		return setWithProperType(t.Elem(), val, structField)
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.Errorf("invalid type:%v", t.Kind())
	}
	return nil
}
