package timex

import "time"

func GetTimeString3() string {
	return time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
}

func GetTimeStringT() string {
	return time.Unix(time.Now().Unix(), 0).Format("2006-01-02T15:04:05")
}

func GetTimeString2() string {
	return time.Unix(time.Now().Unix(), 0).Format("20060102150405")
}

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
