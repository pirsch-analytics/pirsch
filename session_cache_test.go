package pirsch

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestSessionCache(t *testing.T) {
	cleanupDB()
	client := NewMockClient()
	cache := NewSessionCache(client, 10)
	session := cache.get(1, "fp", time.Now().Add(-time.Second*10))
	assert.Nil(t, session)
	client.ReturnSession = &Hit{
		Time:      time.Now().Add(-time.Second * 15),
		SessionID: rand.Uint32(),
		Path:      "/",
		EntryPath: "/entry",
		PageViews: 3,
	}
	session = cache.get(1, "fp", time.Now().Add(-time.Second*20))
	assert.NotNil(t, session)
	assert.Equal(t, "/", session.Path)
	assert.Equal(t, "/entry", session.EntryPath)
	assert.Equal(t, uint16(3), session.PageViews)
	client.ReturnSession = nil
	cache.put(1, "fp", &Hit{
		Path:      session.Path,
		EntryPath: session.EntryPath,
		PageViews: session.PageViews,
		Time:      session.Time,
		SessionID: session.SessionID,
	})
	session = cache.get(1, "fp", time.Now().Add(-time.Second*20))
	assert.NotNil(t, session)
	assert.Equal(t, "/", session.Path)
	assert.Equal(t, "/entry", session.EntryPath)
	assert.Equal(t, uint16(3), session.PageViews)
	cache.put(1, "fp", &Hit{
		Path:      session.Path,
		EntryPath: session.EntryPath,
		PageViews: session.PageViews,
		Time:      time.Now().Add(-time.Second * 21),
		SessionID: rand.Uint32(),
	})
	session = cache.get(1, "fp", time.Now().Add(-time.Second*20))
	assert.Nil(t, session)

	for i := 0; i < 9; i++ {
		cache.put(1, fmt.Sprintf("fp%d", i), &Hit{
			Path:      "/foo",
			EntryPath: "/bar",
			PageViews: 42,
			Time:      time.Now(),
			SessionID: rand.Uint32(),
		})
	}

	assert.Len(t, cache.sessions, 10)
	session = cache.get(1, "fp", time.Now().Add(-time.Minute))
	assert.NotNil(t, session)
	assert.Equal(t, "/", session.Path)
	cache.put(1, "fp10", &Hit{
		Path:      "/foo",
		EntryPath: "/bar",
		PageViews: 42,
		Time:      time.Now(),
		SessionID: rand.Uint32(),
	})
	assert.Len(t, cache.sessions, 1)
	session = cache.get(1, "fp", time.Now().Add(-time.Minute))
	assert.Nil(t, session)
	session = cache.get(1, "fp10", time.Now().Add(-time.Minute))
	assert.NotNil(t, session)
	assert.Equal(t, "/foo", session.Path)
	cache.clear()
	assert.Len(t, cache.sessions, 0)
}
