package utils

import "time"

func TruncDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func FormatDDMMYYYY(t time.Time) string {
	return t.Format("02/01/2006")
}

func MonthKey(t time.Time) string { // "YYYY-MM"
	return t.Format("2006-01")
}

func MonthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}
