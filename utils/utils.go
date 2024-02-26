package utils

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/madlabx/pkgx/log"

	uuid "github.com/satori/go.uuid"
)

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
	m := make(map[string]string)
	structToMapStrStrInternal(input, m)
	return m
}

func structToMapStrStrInternal(input interface{}, m map[string]string) {
	objT := reflect.TypeOf(input)
	objV := reflect.ValueOf(input)

	for i := 0; i < objT.NumField(); i++ {
		switch objV.Field(i).Kind() {
		case reflect.Struct:
			structToMapStrStrInternal(objV.Field(i).Interface(), m)
		case reflect.String:
			m[objT.Field(i).Name] = objV.Field(i).String()
		default:
			log.Errorf("%v, Wrong type:%v, Name:%v", i, objV.Field(i).Kind(), objT.Field(i).Name)
		}
	}
}

func ToString(a interface{}) string {
	resultJson, _ := json.Marshal(a)
	return string(resultJson)
}

func CopyFile(sourceFile, destinationFile string) error {
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(destinationFile, input, 0644)
	if err != nil {
		return err
	}

	return err
}

func Md5Sum(strToSign string) string {
	ret := md5.Sum([]byte(strToSign))
	return fmt.Sprintf("%x", ret[:])
}

func NewRequestId() string {
	uuid := uuid.NewV4()
	return strings.ToUpper(uuid.String())
}

/*
var defaultLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandomString returns a random string with a fixed length

	func RandomString(n int, allowedChars ...[]rune) string {
		var letters []rune

		if len(allowedChars) == 0 {
			letters = defaultLetters
		} else {
			letters = allowedChars[0]
		}

		b := make([]rune, n)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}

		return string(b)
	}
*/

var once sync.Once

func RandomString(size int) string {
	var container string
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	once.Do(func() {
		rand.Seed(time.Now().Unix())
	})

	length := len(str)
	//	bigInt := big.NewInt(int64(length))
	for i := 0; i < size; i++ {
		randomInt := rand.Intn(length)
		container += string(str[randomInt])
	}
	return container
}

func ReadCsvFile(fileName string, data interface{}, headerRows int) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// skip header rows
	for i := 0; i < headerRows; i++ {
		_, err = reader.Read()
		if err != nil {
			return err
		}
	}

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	sliceValue := reflect.ValueOf(data).Elem()
	structType := sliceValue.Type().Elem()

	for _, record := range records {
		structValuePtr := reflect.New(structType)
		structValue := structValuePtr.Elem()

		for i, field := range record {
			fieldValue := structValue.Field(i)

			if !fieldValue.CanSet() {
				continue
			}

			switch fieldValue.Kind() {
			// handle different field types here
			case reflect.String:
				fieldValue.SetString(field)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				intValue, err := strconv.ParseInt(field, 10, 64)
				if err != nil {
					continue
				}
				fieldValue.SetInt(intValue)
			case reflect.Float32, reflect.Float64:
				floatValue, err := strconv.ParseFloat(field, 64)
				if err != nil {
					continue
				}
				fieldValue.SetFloat(floatValue)
			default:
				return fmt.Errorf("Wrong type %v", fieldValue.Kind())
			}
		}

		sliceValue.Set(reflect.Append(sliceValue, structValue))
	}

	return nil
}

func IsDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), nil
}

func CamelToSnake(s string) string {
	var res []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && unicode.IsLower(rune(s[i-1])) {
				res = append(res, '_')
			}
			res = append(res, unicode.ToLower(r))
		} else {
			res = append(res, r)
		}
	}
	return string(res)
}

// Round 对 float64 进行四舍五入，保留指定的小数位数
// digits 表示要保留的小数位数
func Round(num float64, digits int) float64 {
	shift := math.Pow(10, float64(digits))
	return math.Round(num*shift) / shift
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

func FormatToSql(data interface{}) string {
	// 初始化字符串，存储最终的输出结果
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
