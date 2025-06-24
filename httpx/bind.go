package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/madlabx/pkgx/errors"
)

type hxParser struct {
	bodyMap     map[string]any
	bodyParsed  bool
	path        []string
	queryParams url.Values
	headers     http.Header
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
	hp.headers = c.Request().Header

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

<<<<<<< HEAD
func (hp *hxParser) bindAndValidate(input any, target map[string]any, paths ...string) error {
	v := reflect.ValueOf(input)
=======
func (hp *hxParser) bindStruct(fieldValue reflect.Value, value any, newPaths ...string) error {

	if done, err := hp.tryUnmarshalX(fieldValue, value); err != nil {
		return errors.Wrapf(err, "path:%v", newPaths)
	} else if done {
		return nil
	}

	if fieldValue.Kind() == reflect.Pointer {
		fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
	} else {

		fieldValue = fieldValue.Addr()
	}
	structTarget, _ := value.(map[string]any)
	return hp.bindAndValidate(fieldValue.Interface(), structTarget, newPaths...)
}

func (hp *hxParser) parseValue(fieldType reflect.StructField, target map[string]any, hxTags *hxTag, newPaths ...string) (value any, err error) {

	var (
		bv     any
		qv, hv string
	)

	if hxTags.inQuery() {
		qv = hp.queryParams.Get(hxTags.realName(fieldType.Name))
	}

	if hxTags.inBody() {
		bv = target[fieldType.Name]
	}

	if hxTags.inHeader() {
		hv = hp.headers.Get(hxTags.realName(fieldType.Name))
	}

	// apply body in first
	if bv != nil {
		vv := reflect.ValueOf(bv)
		if hxTags.must {
			if vv.Kind() == reflect.Slice || vv.Kind() == reflect.Map || vv.Kind() == reflect.String {
				// 返回数据的长度
				if vv.Len() == 0 {
					err = errors.Errorf("missing parameter %s", strings.Join(newPaths, "."))
					return
				}
			}
		}
		value = bv
	} else if qv != "" {
		value = qv
	} else if hv != "" {
		value = hv
	} else {
		if hxTags.must {
			if hxTags.place != "" {
				err = errors.Errorf("missing %s parameter %s", hxTags.place, strings.Join(newPaths, "."))
			} else {
				err = errors.Errorf("missing parameter %s", strings.Join(newPaths, "."))
			}
			return

		}
		if hxTags.defaultValue != "" {
			value = hxTags.defaultValue
		}
	}

	return
}

func (hp *hxParser) bindAndValidate(input any, target map[string]any, paths ...string) error {

	v := reflect.ValueOf(input)

>>>>>>> 491ef3b (do clean)
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
<<<<<<< HEAD
		field := t.Field(i)
		fieldValue := v.Field(i)
		if fieldValue.Kind() == reflect.Invalid {
			fieldValue.Addr().Set(reflect.New(field.Type))
		}

		if field.Anonymous {
=======
		fieldType := t.Field(i)
		fieldValue := v.Field(i)
		if fieldValue.Kind() == reflect.Invalid {
			fieldValue.Addr().Set(reflect.New(fieldType.Type))
		}

		if fieldType.Anonymous {
>>>>>>> 491ef3b (do clean)
			err := hp.bindAndValidate(fieldValue.Addr().Interface(), target, paths...)
			if err != nil {
				return err
			}
			continue
		}

<<<<<<< HEAD
		newPaths := append(paths, field.Name)

		var newTarget any
		if target != nil {
			newTarget = target[field.Name]
		}

		{
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
			case reflect.Interface:
				fieldValue.Set(reflect.ValueOf(structTarget))
				continue
			default:
				//do nothing
			}
		}

		hxTags, err := parseHxTag(field.Tag, newPaths...)
=======
		newPaths := append(paths, fieldType.Name)

		hxTags, err := parseHxTag(fieldType.Tag, newPaths...)
>>>>>>> 491ef3b (do clean)
		if err != nil {
			return err
		}

<<<<<<< HEAD
		var (
			value, bv any
			qv, hv    string
		)

		if hxTags.inQuery() {
			qv = hp.queryParams.Get(hxTags.realName(field.Name))
		}

		if hxTags.inBody() {
			bv = target[field.Name]
		}

		if hxTags.inHeader() {
			hv = hp.headers.Get(hxTags.realName(field.Name))
		}

		// apply body in first
		if bv != nil {
			vv := reflect.ValueOf(bv)
			if hxTags.must {
				if vv.Kind() == reflect.Slice || vv.Kind() == reflect.Map || vv.Kind() == reflect.String {
					// 返回数据的长度
					if vv.Len() == 0 {
						return errors.Errorf("missing parameter %s", strings.Join(newPaths, "."))
					}
				}
			}
			value = bv
		} else if qv != "" {
			value = qv
		} else if hv != "" {
			value = hv
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
=======
		value, err := hp.parseValue(fieldType, target, hxTags, newPaths...)
		if err != nil {
			return err
		}

		// parse struct member
		switch fieldType.Type.Kind() {
		case reflect.Struct:
			if err = hp.bindStruct(fieldValue, value, newPaths...); err != nil {
				return err
			}
			continue
		case reflect.Ptr:
			if fieldType.Type.Elem().Kind() == reflect.Struct && fieldValue.CanSet() && fieldType.Type.Elem() != reflect.TypeOf(&time.Time{}) {
				if err = hp.bindStruct(fieldValue, value, newPaths...); err != nil {
					return err
				}

				continue
			}
		case reflect.Slice:
			if fieldType.Type.Elem() != reflect.TypeOf(&time.Time{}) {
				if value == nil {
					continue
				}
				sliceTarget, ok := value.([]any)
				if !ok {
					return errors.Errorf("Should be slice %T, path:%v", value, newPaths)
				}

				numElems := len(sliceTarget)
				if numElems > 0 {
					slice := reflect.MakeSlice(fieldType.Type, numElems, numElems)
					for j := 0; j < numElems; j++ {
						item := slice.Index(j)
						switch item.Type().Kind() {
						case reflect.Struct:

							if err = hp.bindStruct(item, sliceTarget[j], newPaths...); err != nil {
								return err
							}
						case reflect.Ptr:
							if item.CanSet() {
								if err = hp.bindStruct(item, sliceTarget[j], newPaths...); err != nil {
									return err
								}
							}
						default:
							if err = hp.setFieldAndValidate(item, item.Type(), sliceTarget[j], hxTags, newPaths...); err != nil {
								return errors.Errorf("%v, path:%v", err, strings.Join(newPaths, "."))
							}
						}
					}
					fieldValue.Set(slice)
					continue
				}

				continue

			}
		case reflect.Interface:
			structTarget, _ := value.(map[string]any)
			fieldValue.Set(reflect.ValueOf(structTarget))
			continue
		default:
			//do nothing
		}

		// parse not struct
		if value != nil && fieldValue.CanSet() {
			if err = hp.setFieldAndValidate(fieldValue, fieldType.Type, value, hxTags, newPaths...); err != nil {
>>>>>>> 491ef3b (do clean)
				return errors.Errorf("%v, path:%v", err, strings.Join(newPaths, "."))
			}
		}
	}

	return nil
}

<<<<<<< HEAD
=======
func (hp *hxParser) tryUnmarshalX(fieldValue reflect.Value, value any) (bool, error) {

	switch fieldValue.Kind() {

	case reflect.Ptr:
		if _, ok := fieldValue.Interface().(Unmarshaler); !ok {
			return false, nil
		}

		if value == nil {
			return true, nil
		}

		fieldValue.Set(reflect.New(fieldValue.Type().Elem()))

		u, _ := fieldValue.Interface().(Unmarshaler)
		return true, u.UnmarshalJSONX([]byte(fmt.Sprint(value)))
	case reflect.Struct:
		if u, ok := fieldValue.Addr().Interface().(Unmarshaler); ok {
			if value == nil {
				return true, nil
			}
			return true, u.UnmarshalJSONX([]byte(fmt.Sprint(value)))
		}

		return false, nil
	default:
		return false, errors.Errorf("Wrong type[%v] for tryUnmarshalX", fieldValue.Kind())
	}
}

>>>>>>> 491ef3b (do clean)
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
<<<<<<< HEAD
	for k, _ := range hp.queryParams {
=======
	for k := range hp.queryParams {
>>>>>>> 491ef3b (do clean)
		if strings.ToLower(k) == key {
			return true
		}
	}

	return false
}

func setIntField(value string, bitSize int, field reflect.Value, ht *hxTag) error {

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

func setUintField(value string, bitSize int, field reflect.Value, ht *hxTag) error {
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

func setBoolField(value string, field reflect.Value, ht *hxTag) error {
	if value == "" {
		value = "false"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value, ht *hxTag) error {
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

func setStringField(value string, field reflect.Value, ht *hxTag) error {
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

// setFieldAndValidate sets a struct field with a value, ensuring it is of the proper type.
<<<<<<< HEAD
func (hp *hxParser) setFieldAndValidate(structFieldPtr reflect.Value, objT reflect.Type, oriV any, ht *hxTag, paths ...string) error {
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
=======
func (hp *hxParser) setFieldAndValidate(structField reflect.Value, objT reflect.Type, oriV any, ht *hxTag, paths ...string) error {
	val := fmt.Sprintf("%v", oriV)

	//TODO value time.Time
	switch objT.Kind() {
	case reflect.Pointer:
		if objT.Elem().Kind() == reflect.Pointer {
			return errors.Errorf("invalid type for field, type:%v, value:%#v, value type:%T", objT.Kind(), oriV, oriV)
		}
		structField.Set(reflect.New(objT.Elem()))
		return hp.setFieldAndValidate(structField.Elem(), objT.Elem(), val, ht, paths...)
>>>>>>> 491ef3b (do clean)
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
<<<<<<< HEAD
	//case reflect.Struct:
	//	return hp.bindAndValidate(structField.Addr().Interface(), paths...)
	default:
		return errors.Errorf("invalid type:%v, value:%#v, value type:%T", objT.Kind(), oriV, oriV)
=======
	default:
		return errors.Errorf("invalid type for field, type:%v, value:%#v, value type:%T", objT.Kind(), oriV, oriV)
>>>>>>> 491ef3b (do clean)
	}
}
