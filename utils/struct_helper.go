package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/madlabx/pkgx/log"
)

func GetFieldValue(sv any, fieldName string) (any, error) {
	v := reflect.ValueOf(sv).Elem() // 获取reflect.Value类型
	t := v.Type()                   // 获取reflect.Type类型

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous {
			result, err := GetFieldValue(v.Field(i).Interface(), fieldName)
			if err == nil {
				return result, nil
			}
		} else if t.Field(i).Name == fieldName {
			fieldValue := v.Field(i)
			return fieldValue.Interface(), nil
		}
	}

	return nil, fmt.Errorf("field '%s' not found", fieldName)
}

func notEmpty(in []string) bool {
	for _, v := range in {
		if len(v) > 0 {
			return true
		}
	}
	return false
}

func BuildSignStringForQueryParams(params map[string][]string) string {
	var buf bytes.Buffer

	signParamKeys := make([]string, 0, len(params))
	for k, v := range params {
		if k != "Sign" && notEmpty(v) {
			signParamKeys = append(signParamKeys, k)
		}
	}
	sort.Strings(signParamKeys)

	for i, key := range signParamKeys {
		if i != 0 {
			buf.WriteString("&")
		}
		for j, value := range params[key] {
			if len(value) > 0 {
				if j != 0 {
					buf.WriteString("&")
				}
				buf.WriteString(key)
				buf.WriteString("=")
				buf.WriteString(value)
			}
		}
	}

	return buf.String()
}

// GetSignString  ignore Sign in r
func GetSignString(r any) string {
	var buf bytes.Buffer

	params := StructToMapStrStr(r)
	signParamKeys := make([]string, 0, len(params))
	for k, v := range params {
		if k != "Sign" && v != "" {
			signParamKeys = append(signParamKeys, k)
		}
	}
	sort.Strings(signParamKeys)

	for i, key := range signParamKeys {
		if i != 0 {
			buf.WriteString("&")
		}
		buf.WriteString(key)
		buf.WriteString("=")
		buf.WriteString(params[key])
	}

	return buf.String()
}

func convertStringToFieldType(fieldType reflect.Kind, filterValue string) (interface{}, error) {
	switch fieldType {
	case reflect.String:
		return filterValue, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(filterValue, 10, 64)
		if err != nil {
			return nil, err
		}
		return intValue, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(filterValue, 10, 64)
		if err != nil {
			return nil, err
		}
		return uintValue, nil
	default:
		return nil, fmt.Errorf("unsupported type: %v", fieldType)
	}
}

func ConvertFilterValueToFieldType(obj interface{}, filterField string, filterValue string) ([]interface{}, error) {
	v := reflect.ValueOf(obj)
	field := v.FieldByName(filterField)
	if !field.IsValid() {
		return nil, fmt.Errorf("field %s is invalid", filterField)
	}

	fieldType := field.Kind()
	vs := strings.Split(filterValue, ",")
	convertedValues := make([]interface{}, len(vs))
	for i, v := range vs {
		//TODO is it necessary to convert??
		convertedValue, err := convertStringToFieldType(fieldType, v)
		if err != nil {
			return nil, err
		}
		convertedValues[i] = convertedValue
	}
	return convertedValues, nil
}

func FormatToSql(data any) string {
	var output strings.Builder

	// 获取结构体数组的反射值
	v := reflect.ValueOf(data)

	// 获取结构体类型
	structType := v.Type().Elem()

	// 构建表头
	fieldNames := make([]string, structType.NumField())
	for i := 0; i < structType.NumField(); i++ {
		fieldNames[i] = structType.Field(i).Name
	}
	output.WriteString(strings.Join(fieldNames, " | "))
	output.WriteString("\n")

	// 输出分隔线
	separator := strings.Repeat("-", output.Len()) + "\n"
	output.WriteString(separator)

	// 输出字段值
	for i := 0; i < v.Len(); i++ {
		// 获取结构体值
		structValue := v.Index(i)

		// 输出每个字段的值
		fieldValues := make([]string, structType.NumField())
		for j := 0; j < structType.NumField(); j++ {
			field := structValue.Field(j)
			fieldValues[j] = fmt.Sprintf("%v", field.Interface())
		}
		output.WriteString(strings.Join(fieldValues, " | "))
		output.WriteString("\n")
	}

	return output.String()
}

func getStructFieldsNames(t reflect.Type) []string {
	var fieldNames []string

	switch t.Kind() {
	case reflect.Ptr, reflect.Slice:
		fieldNames = append(fieldNames, getStructFieldsNames(t.Elem())...)
	case reflect.Struct:
		// 遍历结构体字段
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Anonymous {
				// 如果是匿名字段（嵌套结构体），则递归调用获取嵌套结构体的字段名
				fieldNames = append(fieldNames, getStructFieldsNames(field.Type)...)
			} else {
				//只接受一层
				fieldNames = append(fieldNames, field.Name)
			}
		}
	default:
		fieldNames = append(fieldNames, t.Name())
	}

	return fieldNames
}

func getStructFieldsValues(structValue reflect.Value) []string {

	var values []string

	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}

	if structValue.Kind() != reflect.Struct {
		log.Errorf("Get invalid value: %#v", structValue)
		return nil
	}
	structType := structValue.Type()
	for j := 0; j < structType.NumField(); j++ {
		field := structValue.Field(j)
		fieldType := structType.Field(j).Type
		if field.Kind() == reflect.Ptr {
			field = field.Elem()
			fieldType = fieldType.Elem()
		}
		if !field.IsValid() {
			values = append(values, "-")
			continue
		}

		if fieldType == reflect.TypeOf(time.Time{}) {
			name := structType.Field(j).Name
			timeValue := field.Interface().(time.Time)
			if name == "Date" {
				values = append(values, timeValue.Format("2006-01-02"))
			} else {
				values = append(values, timeValue.Format("2006-01-02 15:04:05"))
			}
		} else if structType.Field(j).Anonymous {
			log.Errorf("%v\n", fieldType)
			values = append(values, getStructFieldsValues(field)...)
		} else if field.Kind() == reflect.Struct {
			values = append(values, "-")
		} else if field.IsValid() && field.CanInterface() {
			values = append(values, fmt.Sprintf("%v", field.Interface()))
		} else {
			log.Errorf("Get invalid field, j=%v, field=%#v", j, field)
			values = append(values, "-")
		}
	}
	return values
}
func getArrayStructFieldsValues(data interface{}) [][]string {
	valueOf := reflect.ValueOf(data)
	// 处理指针类型
	if valueOf.Kind() == reflect.Ptr {
		valueOf = valueOf.Elem()
	}

	var values [][]string

	for i := 0; i < valueOf.Len(); i++ {
		values = append(values, getStructFieldsValues(valueOf.Index(i)))
	}

	return values
}

func FormatToHtmlTable(titles []string, datas ...any) string {
	if len(titles) != len(datas) {
		log.Errorf("Unmatched number for titles and datas")
		return ""
	}

	// 获取结构体的字段信息

	// 构建表格
	var table strings.Builder
	table.WriteString("<style>\n    table {\n        border-collapse: collapse;\n    white-space: nowrap;\n    }\n    th, td {\n        padding: 8px;\n        text-align: left;\n    }\n    th {\n        background-color: #ddd;\n    }\n    tr:nth-child(even) {\n        background-color: #f2f2f2;\n    }\n</style>")

	for i, data := range datas {
		// 获取结构体类型
		fieldNames := getStructFieldsNames(reflect.TypeOf(data))

		// 构建表格数据
		tableData := [][]string{fieldNames}
		tableData = append(tableData, getArrayStructFieldsValues(data)...)

		table.WriteString(fmt.Sprintf("<h1> %s <h1/>\n", titles[i]))
		table.WriteString("<table border=\"1px\" width=\"300px\">\n")
		for i, row := range tableData {
			table.WriteString("<tr>")
			for _, cell := range row {
				// 表头单元格样式
				if i == 0 {
					table.WriteString(fmt.Sprintf(`<th style="padding: 8px;">%s</th>`, cell))
				} else {
					// 内容单元格样式
					//table.WriteString(fmt.Sprintf(`<td style="padding: 8px; border: 1px solid black;">%s</td>`, cell))
					table.WriteString(fmt.Sprintf(`<td style="padding: 8px;">%s</td>`, cell))
				}
			}
			table.WriteString("</tr>")
		}
		table.WriteString("</table><br/>")
	}

	// 将表格添加到邮件内容中
	return table.String()
}

// ToString return "null" if a == nil
func ToString(a any) string {
	resultJson, _ := json.Marshal(a)
	return string(resultJson)
}
func ToPrettyString(a any) string {
	// 使用 json.MarshalIndent 进行美化的JSON编码
	resultJson, err := json.MarshalIndent(a, "", "    ") // 第二个参数是前缀，第三个参数是缩进
	if err != nil {
		// 如果发生错误，返回错误信息
		return err.Error()
	}
	return string(resultJson)
}

func MapToStruct(m map[string]interface{}, s interface{}) {
	v := reflect.ValueOf(s).Elem()

	for key, value := range m {
		field := v.FieldByName(key)

		if field.IsValid() {
			if field.CanSet() {
				field.Set(reflect.ValueOf(value))
			}
		}
	}
}
func MarshalJSONIgnoreTag(js any) ([]byte, error) {
	rv := reflect.ValueOf(js)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// 确保传入的是一个struct
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got %s", rv.Kind())
	}

	// 创建一个map来存储原始成员名和值
	data := make(map[string]interface{})
	if err := collectFields(rv, data); err != nil {
		return nil, err
	}

	// 使用json.Marshal序列化map
	return json.Marshal(data)
}

// collectFields 递归收集结构体字段
func collectFields(rv reflect.Value, data map[string]interface{}) error {
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rv.Type().Field(i)

		// 检查是否是匿名字段
		if fieldType.Anonymous {
			// 递归处理匿名字段
			if err := collectFields(field, data); err != nil {
				return err
			}
		} else {
			// 获取字段的名称
			fieldName := fieldType.Name
			// 将字段值添加到map中
			data[fieldName] = field.Interface()
		}
	}
	return nil
}

func StructToMap(obj interface{}) map[string]string {
	m := make(map[string]string)
	j, err := json.Marshal(obj)
	if err != nil {
		log.Errorf("Failed to marshal when StructToMap, err:%v", err)
		return nil
	}
	err = json.Unmarshal(j, &m)

	return m
}

func StructToMapStrStr(input interface{}) map[string]string {
	if reflect.TypeOf(input).Kind() == reflect.Map {
		return input.(map[string]string)
	}

	m := make(map[string]string)
	structToMapStrStrInternal(input, m)
	return m
}

func structToMapStrStrInternal(input interface{}, m map[string]string) {
	objV := reflect.ValueOf(input)
	if objV.Kind() == reflect.Pointer {
		structToMapStrStrInternal(objV.Elem().Interface(), m)
		return
	}
	objT := objV.Type()

	for i := 0; i < objT.NumField(); i++ {
		fieldI := objV.Field(i)
		filedIT := objT.Field(i).Type
		if fieldI.Kind() == reflect.Ptr {
			fieldI = fieldI.Elem()
			filedIT = filedIT.Elem()
		}

		if !fieldI.IsValid() {
			continue
		}
		switch filedIT.Kind() {
		case reflect.Struct:
			structToMapStrStrInternal(fieldI.Interface(), m)
		case reflect.String:
			m[objT.Field(i).Name] = fieldI.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			m[objT.Field(i).Name] = fmt.Sprintf("%v", fieldI.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			m[objT.Field(i).Name] = fmt.Sprintf("%v", fieldI.Uint())
		default:
			if fieldI.IsNil() {
				continue
			}
			m[objT.Field(i).Name] = fmt.Sprintf("%v", fieldI.Interface())
		}
	}
}
