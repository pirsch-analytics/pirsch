package util

import (
	"time"
)

// Today returns the date for today without time at UTC.
func Today() time.Time {
	return PastDay(0)
}

// PastDay returns the date for today without time minus the given number of days at UTC.
func PastDay(n int) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()-n, 0, 0, 0, 0, time.UTC)
}
