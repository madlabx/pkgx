package stockx

import (
	"strings"
)

func getStockType(stockCode string) string {
	//	if strings.HasPrefix(stockCode, "sh") || strings.HasPrefix(stockCode, "sz") {
	//		return stockCode[:2]
	//	}
	if strings.HasPrefix(stockCode, "50") || strings.HasPrefix(stockCode, "51") ||
		strings.HasPrefix(stockCode, "60") || strings.HasPrefix(stockCode, "73") ||
		strings.HasPrefix(stockCode, "90") || strings.HasPrefix(stockCode, "110") ||
		strings.HasPrefix(stockCode, "113") || strings.HasPrefix(stockCode, "132") ||
		strings.HasPrefix(stockCode, "204") || strings.HasPrefix(stockCode, "78") {
		return "sh"
	}
	if strings.HasPrefix(stockCode, "00") || strings.HasPrefix(stockCode, "13") ||
		strings.HasPrefix(stockCode, "18") || strings.HasPrefix(stockCode, "15") ||
		strings.HasPrefix(stockCode, "16") || strings.HasPrefix(stockCode, "18") ||
		strings.HasPrefix(stockCode, "20") || strings.HasPrefix(stockCode, "30") ||
		strings.HasPrefix(stockCode, "39") || strings.HasPrefix(stockCode, "115") ||
		strings.HasPrefix(stockCode, "1318") {
		return "sz"
	}
	if strings.HasPrefix(stockCode, "5") || strings.HasPrefix(stockCode, "6") ||
		strings.HasPrefix(stockCode, "9") {
		return "sh"
	}
	return "sz"
}

func FormatToxx123456(tsCode string) string {
	//只含数字
	if len(tsCode) == 6 {
		stockType := getStockType(tsCode)
		return stockType + tsCode
	}

	//格式为 123456.sz
	return tsCode[len(tsCode)-2:] + tsCode[:6]
}
func FormatTo123456(tsCode string) string {
	return tsCode[:6]
}
