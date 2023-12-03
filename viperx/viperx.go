package viperx

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strconv"
)

func GetString(name string, def string) string {

	if !viper.IsSet(name) {
		value := os.Getenv(name)
		if len(value) > 0 {
			return value
		}
		return def
	}
	return viper.GetString(name)
}

func GetStrings(name string, def []string) []string {

	if !viper.IsSet(name) {
		return def
	}
	return viper.GetStringSlice(name)
}

func GetInt(name string, def int) int {

	if !viper.IsSet(name) {
		value, ok := os.LookupEnv(name)
		if ok {
			valueF64, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				fmt.Printf("Failed GetFloat64(%v, %v), err:%v", name, def, err)
			} else {
				return int(valueF64)
			}
		}
		return def
	}
	return viper.GetInt(name)
}

func GetInt64(name string, def int64) int64 {

	if !viper.IsSet(name) {
		value, ok := os.LookupEnv(name)
		if ok {
			valueF64, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				fmt.Printf("Failed GetFloat64(%v, %v), err:%v", name, def, err)
			} else {
				return valueF64
			}
		}
		return def
	}
	return viper.GetInt64(name)
}

func GetBool(name string, def bool) bool {

	if !viper.IsSet(name) {
		value, ok := os.LookupEnv(name)
		if ok {
			if value == "true" {
				return true
			}
			if value == "false" {
				return false
			}
			fmt.Printf("Failed GetBool(%v, %v), invalid value:%v", name, def, value)
		}
		return def
	}
	return viper.GetBool(name)
}

func GetFloat64(name string, def float64) float64 {

	if !viper.IsSet(name) {
		value, ok := os.LookupEnv(name)
		if ok {
			valueF64, err := strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Printf("Failed GetFloat64(%v, %v), err:%v", name, def, err)
			} else {
				return valueF64
			}
		}
		return def
	}
	return viper.GetFloat64(name)
}
