package pirsch

import (
	"context"
	"database/sql"
	"time"
)

// RunAtMidnight calls given function on each day of month on midnight (UTC),
// unless it is cancelled by calling the cancel function.
func RunAtMidnight(f func()) context.CancelFunc {
	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() {
		timer := time.NewTimer(0)
		defer timer.Stop()

		for {
			now := time.Now().UTC()
			midnight := time.Date(now.Year(), now.Month(), now.Day(), 24, 0, 0, 0, time.UTC)
			timeToMidnight := midnight.Sub(now)
			timer.Reset(timeToMidnight)

			select {
			case <-timer.C:
				f()
			case <-ctx.Done():
				return // stop loop
			}
		}
	}()

	return cancelFunc
}

// NewTenantID is a helper function to return a sql.NullInt64.
// The ID is considered valid if greater than 0.
func NewTenantID(id int64) sql.NullInt64 {
	return sql.NullInt64{Int64: id, Valid: id > 0}
}

func containsString(list []string, str string) bool {
	for _, item := range list {
		if item == str {
			return true
		}
	}

	return false
}

func today() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func hourInTimezone(hour int, timezone *time.Location) int {
	return time.Date(2020, 1, 1, hour, 0, 0, 0, time.UTC).In(timezone).Hour()
}
