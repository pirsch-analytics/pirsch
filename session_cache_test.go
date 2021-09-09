package pirsch

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSessionCache(t *testing.T) {
	cleanupDB()
	client := NewMockClient()
	cache := newSessionCache(10)
	session := cache.get(client, 1, "fp", time.Now().Add(-time.Second*10))
	assert.Empty(t, session.Path)
	client.ReturnSession = &Session{
		Path:    "/",
		Time:    time.Now().Add(-time.Second * 15),
		Session: time.Now().Add(-time.Second * 20),
	}
	session = cache.get(client, 1, "fp", time.Now().Add(-time.Second*20))
	assert.Equal(t, "/", session.Path)
	client.ReturnSession = nil
	cache.put(1, "fp", session.Path, session.Time, session.Session)
	session = cache.get(client, 1, "fp", time.Now().Add(-time.Second*20))
	assert.Equal(t, "/", session.Path)
	cache.put(1, "fp", session.Path, time.Now().Add(-time.Second*21), time.Now().Add(-time.Second*21))
	session = cache.get(client, 1, "fp", time.Now().Add(-time.Second*20))
	assert.Empty(t, session.Path)

	for i := 0; i < 9; i++ {
		cache.put(1, fmt.Sprintf("fp%d", i), "/foo", time.Now(), time.Now())
	}

	assert.Len(t, cache.sessions, 10)
	session = cache.get(client, 1, "fp", time.Now().Add(-time.Minute))
	assert.Equal(t, "/", session.Path)
	cache.put(1, "fp10", "/foo", time.Now(), time.Now())
	assert.Len(t, cache.sessions, 1)
	session = cache.get(client, 1, "fp", time.Now().Add(-time.Minute))
	assert.Empty(t, session.Path)
	session = cache.get(client, 1, "fp10", time.Now().Add(-time.Minute))
	assert.Equal(t, "/foo", session.Path)
}
