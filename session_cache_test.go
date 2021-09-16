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
	cache := NewSessionCache(client, 10)
	session := cache.get(1, "fp", time.Now().Add(-time.Second*10))
	assert.Nil(t, session)
	client.ReturnSession = &Session{
		Time:      time.Now().Add(-time.Second * 15),
		Session:   time.Now().Add(-time.Second * 20),
		Path:      "/",
		EntryPath: "/entry",
		PageViews: 3,
	}
	session = cache.get(1, "fp", time.Now().Add(-time.Second*20))
	assert.NotNil(t, session)
	assert.Equal(t, "/", session.Path)
	assert.Equal(t, "/entry", session.EntryPath)
	assert.Equal(t, 3, session.PageViews)
	client.ReturnSession = nil
	cache.put(1, "fp", session.Path, session.EntryPath, session.PageViews, session.Time, session.Session)
	session = cache.get(1, "fp", time.Now().Add(-time.Second*20))
	assert.NotNil(t, session)
	assert.Equal(t, "/", session.Path)
	assert.Equal(t, "/entry", session.EntryPath)
	assert.Equal(t, 3, session.PageViews)
	cache.put(1, "fp", session.Path, session.EntryPath, session.PageViews, time.Now().Add(-time.Second*21), time.Now().Add(-time.Second*21))
	session = cache.get(1, "fp", time.Now().Add(-time.Second*20))
	assert.Nil(t, session)

	for i := 0; i < 9; i++ {
		cache.put(1, fmt.Sprintf("fp%d", i), "/foo", "/bar", 42, time.Now(), time.Now())
	}

	assert.Len(t, cache.sessions, 10)
	session = cache.get(1, "fp", time.Now().Add(-time.Minute))
	assert.NotNil(t, session)
	assert.Equal(t, "/", session.Path)
	cache.put(1, "fp10", "/foo", "/bar", 42, time.Now(), time.Now())
	assert.Len(t, cache.sessions, 1)
	session = cache.get(1, "fp", time.Now().Add(-time.Minute))
	assert.Nil(t, session)
	session = cache.get(1, "fp10", time.Now().Add(-time.Minute))
	assert.NotNil(t, session)
	assert.Equal(t, "/foo", session.Path)
}
