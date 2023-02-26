package timeutils

import "time"

func GetCurrentYYMM() string {
	return time.Now().Format("0601")
}

func GetCurrentYYMMDD() string {
	return time.Now().Format("060102")
}

func GetCurrentYYYYMM() string {
	return time.Now().Format("200601")
}

func GetCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}

func GetCurrentTime() string {
	return GetTimeFormat(time.Now())
}

func GetTimeFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000000")
}
