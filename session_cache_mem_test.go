package pirsch

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestSessionCacheMem(t *testing.T) {
	cleanupDB()
	client := NewMockClient()
	cache := NewSessionCacheMem(client, 10)
	session := cache.Get(1, 1, time.Now().Add(-time.Second*10))
	assert.Nil(t, session)
	client.ReturnSession = &Session{
		Time:      time.Now().Add(-time.Second * 15),
		SessionID: rand.Uint32(),
		Path:      "/",
		EntryPath: "/entry",
		PageViews: 3,
	}
	session = cache.Get(1, 1, time.Now().Add(-time.Second*20))
	assert.NotNil(t, session)
	assert.Equal(t, "/", session.Path)
	assert.Equal(t, "/entry", session.EntryPath)
	assert.Equal(t, uint16(3), session.PageViews)
	client.ReturnSession = nil
	cache.Put(1, 1, &Session{
		Path:      session.Path,
		EntryPath: session.EntryPath,
		PageViews: session.PageViews,
		Time:      session.Time,
		SessionID: session.SessionID,
	})
	session = cache.Get(1, 1, time.Now().Add(-time.Second*20))
	assert.NotNil(t, session)
	assert.Equal(t, "/", session.Path)
	assert.Equal(t, "/entry", session.EntryPath)
	assert.Equal(t, uint16(3), session.PageViews)
	cache.Put(1, 1, &Session{
		Path:      session.Path,
		EntryPath: session.EntryPath,
		PageViews: session.PageViews,
		Time:      time.Now().Add(-time.Second * 21),
		SessionID: rand.Uint32(),
	})
	session = cache.Get(1, 1, time.Now().Add(-time.Second*20))
	assert.Nil(t, session)

	for i := 0; i < 9; i++ {
		cache.Put(1, uint64(i+2), &Session{
			SessionID: rand.Uint32(),
			Time:      time.Now(),
			Path:      "/foo",
			EntryPath: "/bar",
			PageViews: 42,
		})
	}

	assert.Len(t, cache.sessions, 10)
	session = cache.Get(1, 1, time.Now().Add(-time.Minute))
	assert.NotNil(t, session)
	assert.Equal(t, "/", session.Path)
	cache.Put(1, 10, &Session{
		Path:      "/foo",
		EntryPath: "/bar",
		PageViews: 42,
		Time:      time.Now(),
		SessionID: rand.Uint32(),
	})
	assert.Len(t, cache.sessions, 1)
	session = cache.Get(1, 1, time.Now().Add(-time.Minute))
	assert.Nil(t, session)
	session = cache.Get(1, 10, time.Now().Add(-time.Minute))
	assert.NotNil(t, session)
	assert.Equal(t, "/foo", session.Path)
	cache.Clear()
	assert.Len(t, cache.sessions, 0)
}
