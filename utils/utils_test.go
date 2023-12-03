package utils

import (
	"fmt"
	"testing"
)

func TestFormat(t *testing.T) {
	type CccT struct {
		CA int
		CB float64
	}
	type BbbT struct {
		BA int
		BB float64
		CccT
	}
	type AaaT struct {
		AA int
		AB int
		BbbT
	}

	data := AaaT{
		AA: 1,
		AB: 2,
	}
	data.BA = 3
	data.BB = 4.0
	data.CA = 5
	data.CB = 6.0

	data2 := AaaT{
		AA: 11,
		AB: 12,
	}
	data2.BA = 13
	data2.BB = 14.0
	data2.CA = 15
	data2.CB = 16.0

	fmt.Printf("%v\n", FormatToHtml([]string{"a1", "a2"}, []AaaT{data, data2}))
}
