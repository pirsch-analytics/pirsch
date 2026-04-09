package referrer

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAndroidAppCache(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(100)

	for range 100 {
		go func() {
			name, icon := androidAppCache.get(androidAppPrefix + "com.Slack")
			assert.Equal(t, "Slack", name)
			assert.NotEmpty(t, icon)
			wg.Done()
		}()
	}

	wg.Wait()
}
