package referrer

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestAndroidAppCache(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			name, icon := androidAppCache.get(androidAppPrefix + "com.Slack")
			assert.Equal(t, "Slack", name)
			assert.NotEmpty(t, icon)
			wg.Done()
		}()
	}

	wg.Wait()
}
