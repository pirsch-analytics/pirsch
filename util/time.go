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
