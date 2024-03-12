package httpx

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/madlabx/pkgx/utils"

	"github.com/madlabx/pkgx/log"

	"github.com/labstack/echo"
)

var (
	TagHttpX        = "hx_tag"
	TagFiledPlace   = "hx_place"
	TagFieldName    = "hx_query_name"
	TagFieldMust    = "hx_must"
	TagFieldDefault = "hx_default"
	TagFieldRange   = "hx_range"
)

type httpXTag struct {
	place        string
	name         string
	must         bool
	defaultValue string
	valueRange   string
}

func (ht *httpXTag) isEmpty() bool {
	return ht.name == "" &&
		!ht.must &&
		ht.defaultValue == "" &&
		ht.valueRange == ""
}

func parseTag(t reflect.StructTag) (*httpXTag, error) {
	var (
		place, name, mustStr, defaultValue, valueRange string
		must                                           bool
	)
	tags := t.Get(TagHttpX)
	if len(tags) > 0 {
		tagList := strings.Split(tags, ";")
		if len(tagList) != 5 {
			return nil, fmt.Errorf("invalid "+TagHttpX+":'%v' which should have 5 fields", tags)
		}
		place, name, mustStr, defaultValue, valueRange = tagList[0], tagList[1], tagList[2], tagList[3], tagList[4]
	} else {
		place, name, mustStr, defaultValue, valueRange =
			t.Get(TagFiledPlace),
			t.Get(TagFieldName),
			t.Get(TagFieldMust),
			t.Get(TagFieldDefault),
			t.Get(TagFieldRange)
	}

	if strings.ToLower(mustStr) == "true" {
		must = true
	} else if strings.ToLower(mustStr) == "false" {
		must = false
	} else if len(mustStr) > 0 {
		return nil, fmt.Errorf("invalid must tag:%v", mustStr)
	}

	return &httpXTag{
		place:        place,
		name:         name,
		must:         must,
		defaultValue: defaultValue,
		valueRange:   valueRange,
	}, nil
}

func bindAndValidateStructField(t reflect.Type, vs reflect.Value, c echo.Context) error {
	//1. 判断field的类型,如果是指针，要解开
	if t.Kind() == reflect.Ptr {
		if !vs.IsValid() {
			return nil
		}
		vs = vs.Elem()
	}

	//2. 如果不是结构体，报错
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("%v should be struct, type: %s", t.Name(), t.Kind().String())
	}

	//3. 遍历成员解析
	for index := 0; index < t.NumField(); index++ {
		field := t.Field(index)
		v := vs.Field(index)

		kind := field.Type.Kind()
		if field.Type.Kind() == reflect.Ptr {
			kind = reflect.TypeOf(v.Interface()).Elem().Kind()
			v = v.Elem()

			//kind = v.Type().Kind()
		}

		// 如果是匿名字段，或者非time.Time的结构体
		if field.Anonymous ||
			(field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{})) {
			err := bindAndValidateStructField(field.Type, v, c)
			if err != nil {
				return err
			}
			continue
		}

		//parse httpx tag
		ht, err := parseTag(field.Tag)
		if err != nil {
			return err
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
		case "query":
			//Todo
			value = c.QueryParam(name)
			if value == "" {
				value = c.QueryParam(strings.ToLower(name))
			}

			if ht.must && value == "" {
				return fmt.Errorf("query param %s is required", name)
			}
		case "", "body":
			if v.IsValid() && v.CanInterface() {
				value = fmt.Sprintf("%v", v.Interface())
			}

			if ht.must && value == "" {
				return fmt.Errorf("request param %s is required", name)
			}
		default:
			return fmt.Errorf("invalid "+TagFiledPlace+" tag: %v", ht.place)
		}

		//if value == "" {
		//	value = ht.defaultValue
		//}

		//校验字段值
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
					return fmt.Errorf("invalid value:%s for field %s, must be one of %v", value, name, allowedValues)
				}
			}
			//
			//if ht.defaultValue == value {
			//	if field.Type.Kind() == reflect.Ptr {
			//
			//		//field.Set(reflect.New(field.Type().Elem()))
			//		vs.Field(index).Set(reflect.ValueOf(&value))
			//	} else {
			//		vs.Field(index).SetString(value)
			//	}
			//}
		case reflect.Bool:
			_, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid value:%s for field %s", value, name)
			}
			//
			//if ht.defaultValue == value {
			//	if field.Type.Kind() == reflect.Ptr {
			//		vs.Field(index).Set(reflect.ValueOf(&fieldValue))
			//	} else {
			//		vs.Field(index).SetBool(fieldValue)
			//	}
			//}

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid value:%v for field %s, should be an integer", value, name)
			}

			if ht.valueRange != "" {
				rangeValues := strings.Split(ht.valueRange, "-")
				minVal, e1 := strconv.ParseInt(rangeValues[0], 10, 64)
				maxVal, e2 := strconv.ParseInt(rangeValues[1], 10, 64)
				if e1 != nil || e2 != nil {
					return fmt.Errorf("invalid format for "+TagFieldRange+":%v, field:%v", ht.valueRange, field.Name)
				}
				if fieldValue < minVal || fieldValue > maxVal {
					return fmt.Errorf("invalid value:%s for field %s, must be between %d and %d", value, name, minVal, maxVal)
				}
			}
			//
			//if ht.defaultValue == value {
			//	if field.Type.Kind() == reflect.Ptr {
			//		vs.Field(index).Set(reflect.ValueOf(&fieldValue))
			//	} else {
			//		vs.Field(index).SetInt(fieldValue)
			//	}
			//}

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fieldValue, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid value:%v for field %s, should be an unsigned integer", value, name)
			}
			if ht.valueRange != "" {
				rangeValues := strings.Split(ht.valueRange, "-")
				if len(rangeValues) != 2 {
					return fmt.Errorf("invalid "+TagFieldRange+":%s for field %v", ht.valueRange, name)
				}
				minVal, e1 := strconv.ParseUint(rangeValues[0], 10, 64)
				maxVal, e2 := strconv.ParseUint(rangeValues[1], 10, 64)
				if e1 != nil || e2 != nil {
					return fmt.Errorf("invalid format for "+TagFieldRange+":%v, field:%v", ht.valueRange, field.Name)
				}
				if fieldValue < minVal || fieldValue > maxVal {
					return fmt.Errorf("invalid value:%s for field %s, must be between %d and %d", value, name, minVal, maxVal)
				}
			}

			//if ht.defaultValue == value {
			//	if field.Type.Kind() == reflect.Ptr {
			//		vs.Field(index).Set(reflect.ValueOf(&fieldValue))
			//	} else {
			//		vs.Field(index).SetUint(fieldValue)
			//	}
			//}

		case reflect.Float32, reflect.Float64:
			fieldValue, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid value:%v for field %s, should be a float", value, name)
			}
			if ht.valueRange != "" {
				rangeValues := strings.Split(ht.valueRange, "-")
				if len(rangeValues) != 2 {
					return fmt.Errorf("invalid value range:%s", ht.valueRange)
				}
				minVal, e1 := strconv.ParseFloat(rangeValues[0], 64)
				maxVal, e2 := strconv.ParseFloat(rangeValues[1], 64)
				if e1 != nil || e2 != nil {
					return fmt.Errorf("invalid format for "+TagFieldRange+":%v, field %v", ht.valueRange, field.Name)
				}
				if fieldValue < minVal || fieldValue > maxVal {
					return fmt.Errorf("invalid value:%v for field %v, must be between %v and %v", value, name, minVal, maxVal)
				}
			}

			//if field.Type.Kind() == reflect.Ptr {
			//	vs.Field(index).Set(reflect.ValueOf(&fieldValue))
			//} else {
			//	vs.Field(index).SetFloat(fieldValue)
			//}

		case reflect.Struct:
			if field.Type.String() == "time.Time" {
				layout := "2006-01-02T15:04:05"
				timeValue, err := time.Parse(layout, value)
				if err != nil {
					return fmt.Errorf("invalid time format for field %s, should be in YYYY-MM-DDTHH:MM:SS format", name)
				}

				if ht.valueRange != "" {
					rangeValues := strings.Split(ht.valueRange, "-")
					minUnix, e1 := strconv.ParseInt(rangeValues[0], 10, 64)
					maxUnix, e2 := strconv.ParseInt(rangeValues[1], 10, 64)
					if e1 != nil || e2 != nil {
						return fmt.Errorf("invalid format for "+TagFieldRange+":%v, field %v", ht.valueRange, field.Name)
					}

					unixTime := timeValue.Unix()
					if unixTime < minUnix || unixTime > maxUnix {
						return fmt.Errorf("%s is not within the allowed range for field %s", value, name)
					}
				}
				//if field.Type.Kind() == reflect.Ptr {
				//	vs.Field(index).Set(reflect.ValueOf(&timeValue))
				//} else {
				//	vs.Field(index).Set(reflect.ValueOf(timeValue))
				//}
			} else {
				return fmt.Errorf("unsupported struct field type:%v for field %s", field.Type, name)
			}
		default:
			return fmt.Errorf("unsupported field type:%v", reflect.TypeOf(v))
		}

		//if err := setWithProperType(field.Type.Kind(), value, vs.Field(index)); err != nil {
		//	return err
		//}

	}
	return nil
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

httpx tag自定义如下：

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

// 根据字符串值和成员变量的类型将字符串转换为恰当的类型
func strToType(strValue string, valueType reflect.Type) (reflect.Value, error) {
	switch valueType.Kind() {
	case reflect.Ptr:
		v, err := strToType(strValue, valueType.Elem())
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		return reflect.ValueOf(&v), nil
	case reflect.String:
		if valueType.Kind() == reflect.Ptr {
			temp := strValue
			return reflect.ValueOf(&temp), nil
		}
		return reflect.ValueOf(strValue), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(strValue, 10, valueType.Bits())
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		if valueType.Kind() == reflect.Ptr {
			return reflect.ValueOf(&intVal).Convert(valueType), nil
		}
		return reflect.ValueOf(intVal).Convert(valueType), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(strValue, 10, valueType.Bits())
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		if valueType.Kind() == reflect.Ptr {
			return reflect.ValueOf(&uintVal).Convert(valueType), nil
		}
		return reflect.ValueOf(uintVal).Convert(valueType), nil
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(strValue, valueType.Bits())
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		if valueType.Kind() == reflect.Ptr {
			return reflect.ValueOf(&floatVal).Convert(valueType), nil
		}
		return reflect.ValueOf(floatVal).Convert(valueType), nil
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(strValue)
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		if valueType.Kind() == reflect.Ptr {
			return reflect.ValueOf(&boolVal).Convert(valueType), nil
		}
		return reflect.ValueOf(boolVal), nil
	default:
		return reflect.ValueOf(nil), fmt.Errorf("unsupported type %s", valueType.Name())
	}
}
func SetDefaults(data interface{}) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("data must be a non-nil pointer to a struct, %v", v)
	}

	v = v.Elem()
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		if v.Kind() == reflect.Invalid {
			v.Addr().Set(reflect.New(t))
		}
		return SetDefaults(v.Interface())
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a pointer to a struct, current:%v", t.Kind())
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		if fieldValue.Kind() == reflect.Invalid {
			fieldValue.Addr().Set(reflect.New(field.Type))
		}

		// 处理嵌套结构体或指向结构体的指针
		if field.Type.Kind() == reflect.Struct ||
			(field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {

			if fieldValue.CanAddr() {
				//if fieldValue.IsNil() && fieldValue.Kind() == reflect.Ptr {
				if fieldValue.Kind() == reflect.Ptr {
					fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
				}
				err := SetDefaults(fieldValue.Addr().Interface())
				if err != nil {
					return err
				}
			}
			continue
		}

		// 获取 hx_default 标签
		defaultValue, hasDefault := field.Tag.Lookup("hx_default")
		if !hasDefault || !fieldValue.CanSet() {
			continue
		}

		if err := setWithProperType(field.Type, defaultValue, fieldValue.Addr()); err != nil {
			log.Errorf("failed to setWithProperType %v %v to %v", field.Name, field.Type.Name(), defaultValue)
			return err
		}
		//// 转换类型并设置默认值
		//val, err := strToType(defaultValue, field.Type)
		//if err != nil {
		//	log.Errorf("failed to set %v to %v", field.Name, defaultValue)
		//	return err
		//}
		//
		//if field.Type.Kind() == reflect.Ptr && fieldValue.IsNil() {
		//	fieldValue.Set(reflect.New(field.Type.Elem()))
		//}
		//
		//fieldValue.Set(val.Convert(fieldValue.Type()))
	}

	return nil
}

func BindAndValidate(c echo.Context, i interface{}) error {
	err := SetDefaults(i)
	if err != nil {
		log.Errorf("XXXX err:%v", err)
		return err
	}

	log.Errorf("TETSTSTTX: %v", utils.ToString(i))

	v := reflect.ValueOf(i).Elem()
	t := v.Type()

	if err := c.Bind(i); err != nil && !strings.Contains(err.Error(), "Request body can't be empty") {
		return err
	}

	log.Errorf("AfterBind: %v", utils.ToString(i))

	return bindAndValidateStructField(t, v, c)
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

func setWithProperType(t reflect.Type, val string, v reflect.Value) error {
	if v.Elem().Kind() == reflect.Invalid {
		v.Set(reflect.New(t))
	}
	structField := v.Elem()

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
		return fmt.Errorf("unknown type:%v", t.Kind())
	}
	return nil
}
