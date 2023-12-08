package utils

import (
	"fmt"
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

	fmt.Printf("%v\n", FormatToHtml([]string{"a1"}, []AaaT{data, data2}))
}
