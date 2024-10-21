package timex

import "time"


func UnixTimeToGbLocalTime(sec int64) string {
	tm := time.Unix(sec, 0)
	return tm.Format("2006-01-02T15:04:05")
}

func GbTimeToUnixTime(timeStr string) int64 {
	gbFormat := "2006-01-02T15:04:05"
	p, _ := time.ParseInLocation(gbFormat, timeStr, time.Local)
	return p.Unix()
}

func GbTimeToUnixTime2(timeStr string) int64 {
	gbFormat := "20060102150405"
	p, _ := time.ParseInLocation(gbFormat, timeStr, time.Local)
	return p.Unix()
}

// 时间格式变化 20060102 -> 2006-01-02
func FormatDate(dateStr string) (time.Time, error) {

	// 定义日期格式
	dateFormat := "20060102" // Go语言中的日期格式化字符串，对应 "YYYYMMDD"

	// 解析日期字符串为时间格式
	t, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}
func NowInYYMMDD() string {
	return time.Now().Format("2006-01-02")
}

func NowInHHMMSS() string {
	return time.Now().Format("15:04:05")
}

func NowInISO8601() string {
	return time.Unix(time.Now().Unix(), 0).Format("2006-01-02T15:04:05")
}
func NowInISO8601Var() string {
	return time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
}

func NowInYYMMDDHHMMSS() string {
	return time.Now().Format("20060102150405")
}
