package utils

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	type GGG struct {
		G  int
		GG int
	}
	type TTT struct {
		E *time.Time
		F time.Time
	}
	type CccT struct {
		CA *int
		CB float64
		TTT
		O *GGG
	}
	type BbbT struct {
		BA int
		BB *float64
		CccT
	}
	type AaaT struct {
		AA int
		AB int
		BbbT
	}

	g := GGG{
		G:  1111111111111,
		GG: 2,
	}
	now := time.Now()

	data := AaaT{
		AA: 1,
		AB: 2,
	}
	a := 4.0
	b := 3
	data.BA = 3
	data.BB = &a
	data.CA = &b
	data.CB = 6.0
	data.E = &now
	data.F = now
	data.O = &g

	data2 := AaaT{
		AA: 11,
		AB: 12,
	}
	data2.BA = 13
	data2.BB = &a
	data2.CA = &b
	data2.CB = 16.0
	//data2.E = &now

	d := []AaaT{data, data2}
	// 获取结构体类型
	fieldNames := getStructFieldsNames(reflect.TypeOf(d))

	// 构建表格数据
	tableData := [][]string{fieldNames}
	tableData = append(tableData, getArrayStructFieldsValues(d)...)

	for i, td := range tableData {
		fmt.Printf("%v", i)
		for _, ts := range td {
			fmt.Printf("[%v] ", ts)
		}
		fmt.Printf("\n")
	}

	//fmt.Printf("%v\n", FormatToHtml([]string{"a1"}, []AaaT{data, data2}))
}
