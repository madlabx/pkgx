package utils

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func HeadFirstLineFromFile(fileName string) (string, error) {
	// Read the contents of the file.
	data, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	// Split the content into lines.
	lines := strings.Split(string(data), "\n")

	// Get first non-empty line in the file.
	for _, line := range lines {
		if line != "" {
			return line, nil
		}
	}

	// If no non-empty line was found, return an error.
	return "", fmt.Errorf("no valid content found in the file %s", fileName)
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
