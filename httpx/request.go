package httpx

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
)

func ValidateMust(input interface{}, keys ...string) error {
	v := reflect.ValueOf(input)
	t := reflect.TypeOf(input)

	for _, key := range keys {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i)
			if field.Name == key && value.IsZero() {
				return ErrStrResp(http.StatusBadRequest, handleGetECodeBadRequest(), "Need "+key)
			}
		}
	}

	return nil
}

func QueryMustParam(c echo.Context, key string) (string, error) {
	var err error
	value := c.QueryParam(key)
	if len(value) == 0 {
		err = ErrStrResp(http.StatusBadRequest, handleGetECodeBadRequest(), "Missing "+key)
	}

	return value, err
}

func QueryOptionalParam(c echo.Context, key string) (string, bool) {
	value := c.QueryParam(key)
	return value, len(value) != 0
}

type httpxTag struct {
	place        string
	name         string
	must         bool
	defaultValue string
	valueRange   string
}

func (ht *httpxTag) isEmpty() bool {
	return ht.name == "" &&
		!ht.must &&
		ht.defaultValue == "" &&
		ht.valueRange == ""
}

func parseTag(t reflect.StructTag) (*httpxTag, error) {
	ht := httpxTag{
		place:        t.Get(TagFiledPlace),
		name:         t.Get(TagFieldName),
		defaultValue: t.Get(TagFieldDefault),
		valueRange:   t.Get(TagFieldRange),
	}
	mustString := t.Get(TagFieldMust)

	if strings.ToLower(mustString) == "true" {
		ht.must = true
	} else if strings.ToLower(mustString) == "false" {
		ht.must = false
	} else if len(mustString) > 0 {
		return nil, fmt.Errorf("invalid must tag:%v", mustString)
	}

	return &ht, nil
}

/*
	BindAndValidate 提供从echo.Context里请求里Request的queryparemeter或者body里取相应的值，赋值给i，并做自校验。使用方式为

	type Trans struct {
		Bandwidth uint64 `hx_place:"query" hx_must:"true" hx_query_name:"bindwidth" hx_default:"default_name" hx_range:"1-2"`
		Loss     float64 `hx_place:"body" hx_must:"false" hx_query_name:"loss" hx_default:"default_name" hx_range:"1.2,3.4"`
	}

	type TusReq struct {
		Name       string `hx_place:"query" hx_must:"true" hx_query_name:"host_name" hx_default:"default_name" hx_range:"alice,bob"`
		TaskId     int64  `hx_place:"body" hx_must:"false" hx_query_name:"task_id" hx_default:"7" hx_range:"0-21"`
		Transfer   Trans
	}

	var i TusReq
	BindAndValidate(e, &i)

根据自定义的tag做校验：

	hx_place: query表示该值从query parameter里取，body从请求body里取。
	hx_must: true表示必须，若未赋值，则报错；false表示可选
	hx_query_name： Query Parameters中定义的名称
	hx_default: 若未赋值，设为该默认值
	hx_range: 根据i的字段的类型来校验range：若为整数，0-21表示0到21是合法的，否则报错；若为字符串，"alice,bob"表示只能为alice或bob，否则报错。
*/
func bindAndValidateStructField(t reflect.Type, v reflect.Value, c echo.Context) error {
	//1. 判断field的类型,如果是指针，要解开
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	//2. 如果不是结构体，报错
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("%v should be struct, type: %s", t.Name(), t.Kind().String())
	}

	//3. 遍历成员解析
	for index := 0; index < t.NumField(); index++ {
		field := t.Field(index)

		// 如果是匿名字段，或者非time.Time的结构体
		if field.Anonymous ||
			(field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{})) {
			err := bindAndValidateStructField(field.Type, v.Field(index), c)
			if err != nil {
				return err
			}
			continue
		}

		//解析httpx tag
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
		case "body":
			fieldValue := v.Field(index).Interface()
			value = fmt.Sprintf("%v", fieldValue)
			if c.Request().FormValue(name) == "" {
				value = ""
			}
			if ht.must && value == "" {
				return fmt.Errorf("request param %s is required", field.Name)
			}
		default:
			return fmt.Errorf("invalid "+TagFiledPlace+" tag: %v", ht.place)
		}

		if value == "" {
			value = ht.defaultValue
		}

		//校验字段值
		if ht.valueRange == "" || value == "" {
			continue
		}

		switch field.Type.Kind() {
		case reflect.String:
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
			v.Field(index).SetString(value)

		case reflect.Int, reflect.Int64:
			fieldValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid value:%v for field %s, should be an integer", value, name)
			}

			rangeValues := strings.Split(ht.valueRange, "-")
			minVal, e1 := strconv.ParseInt(rangeValues[0], 10, 64)
			maxVal, e2 := strconv.ParseInt(rangeValues[1], 10, 64)
			if e1 != nil || e2 != nil {
				return fmt.Errorf("invalid format for "+TagFieldRange+":%v, field:%v", ht.valueRange, field.Name)
			}
			if fieldValue < minVal || fieldValue > maxVal {
				return fmt.Errorf("invalid value:%s for field %s, must be between %d and %d", value, name, minVal, maxVal)
			}
			v.Field(index).SetInt(fieldValue)

		case reflect.Uint, reflect.Uint64:
			fieldValue, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid value:%v for field %s, should be an unsigned integer", value, name)
			}
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
			v.Field(index).SetUint(fieldValue)

		case reflect.Float64:
			fieldValue, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid value:%v for field %s, should be a float", value, name)
			}
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
				return fmt.Errorf("invalid value:%v or field %v, must be between %v and %v", value, name, minVal, maxVal)
			}
			v.Field(index).SetFloat(fieldValue)

		case reflect.Struct:
			if field.Type.String() == "time.Time" {
				layout := "2006-01-02T15:04:05"
				rangeValues := strings.Split(ht.valueRange, "-")
				minUnix, e1 := strconv.ParseInt(rangeValues[0], 10, 64)
				maxUnix, e2 := strconv.ParseInt(rangeValues[1], 10, 64)
				if e1 != nil || e2 != nil {
					return fmt.Errorf("invalid format for "+TagFieldRange+":%v, field %v", ht.valueRange, field.Name)
				}

				timeValue, err := time.Parse(layout, value)
				if err != nil {
					return fmt.Errorf("invalid time format for field %s, should be in YYYY-MM-DDTHH:MM:SS format", name)
				}
				unixTime := timeValue.Unix()
				if unixTime < minUnix || unixTime > maxUnix {
					return fmt.Errorf("%s is not within the allowed range for field %s", value, name)
				}
				v.Field(index).Set(reflect.ValueOf(timeValue))
			} else {
				return fmt.Errorf("unsupported struct field type:%v for field %s", field.Type, name)
			}
		default:
			return fmt.Errorf("unsupported field type:%v", field.Type)
		}

	}
	return nil
}
func BindAndValidate(c echo.Context, i interface{}) error {
	v := reflect.ValueOf(i).Elem()
	t := v.Type()

	//给i的成员解析赋值
	if err := c.Bind(i); err != nil {
		return err
	}

	return bindAndValidateStructField(t, v, c)
}
