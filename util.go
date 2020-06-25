package pirsch

import "time"

// RunAtMidnight calls given function on each day of month on midnight.
func RunAtMidnight(f func()) {
	go func() {
		for {
			<-time.After(timeToMidnight())
			f()
		}
	}()
}

func timeToMidnight() time.Duration {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 24, 0, 0, 0, time.UTC)
	return midnight.Sub(now)
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
