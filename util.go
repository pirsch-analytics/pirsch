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
		for {
			now := time.Now()
			midnight := time.Date(now.Year(), now.Month(), now.Day(), 24, 0, 0, 0, time.UTC)
			timeToMidnight := midnight.Sub(now)

			select {
			case <-time.After(timeToMidnight):
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

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func shortenString(str string, maxLen int) string {
	if len(str) > maxLen {
		return str[:maxLen]
	}

	return str
}
