package util

import "testing"

func TestRunAtMidnight(t *testing.T) {
	cancel := RunAtMidnight(func() {
		panic("Function must not be called")
	})
	cancel()
}
