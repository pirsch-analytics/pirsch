package pirsch

import "time"

// RunAtMidnight calls given function on each day of month on midnight.
func RunAtMidnight(f func()) {
	go func() {
		for {
			time.AfterFunc(timeToMidnight(), f)
		}
	}()
}

func timeToMidnight() time.Duration {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 24, 0, 0, 0, now.Location())
	return midnight.Sub(now)
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
