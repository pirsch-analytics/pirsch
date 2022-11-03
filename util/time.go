package util

import (
	"context"
	"time"
)

// RunAtMidnight calls given function on each day of month on midnight (UTC),
// unless it is cancelled by calling the cancel function.
func RunAtMidnight(f func()) context.CancelFunc {
	ctx, cancelFunc := context.WithCancel(context.Background())

	go func() {
		timer := time.NewTimer(time.Second * 1)
		defer func() {
			if !timer.Stop() {
				<-timer.C
			}
		}()

		for {
			timer.Reset(getTimeToMidnightUTC())

			select {
			case <-timer.C:
				f()
			case <-ctx.Done():
				return
			}
		}
	}()

	return cancelFunc
}

func getTimeToMidnightUTC() time.Duration {
	now := time.Now().UTC()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	return midnight.Sub(now)
}

// Today returns the date for today without time at UTC.
func Today() time.Time {
	return PastDay(0)
}

// PastDay returns the date for today without time minus the given number of days at UTC.
func PastDay(n int) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()-n, 0, 0, 0, 0, time.UTC)
}
