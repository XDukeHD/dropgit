package utils

import "time"

func CurrentDateStr() string {
	return time.Now().Format("2006-01-02")
}

func CurrentDateTimeISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
