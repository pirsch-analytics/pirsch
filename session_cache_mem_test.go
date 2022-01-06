package pirsch

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestSessionCacheMem(t *testing.T) {
	client := NewMockClient()
	cache := NewSessionCacheMem(client, 10)
	session := cache.Get(1, 1, time.Now().Add(-time.Second*10))
	assert.Nil(t, session)
	client.ReturnSession = &Session{
		Time:      time.Now().Add(-time.Second * 15),
		SessionID: rand.Uint32(),
		ExitPath:  "/",
		EntryPath: "/entry",
		PageViews: 3,
	}
	session = cache.Get(1, 1, time.Now().Add(-time.Second*20))
	assert.NotNil(t, session)
	assert.Equal(t, "/", session.ExitPath)
	assert.Equal(t, "/entry", session.EntryPath)
	assert.Equal(t, uint16(3), session.PageViews)
	client.ReturnSession = nil
	cache.Put(1, 1, &Session{
		ExitPath:  session.ExitPath,
		EntryPath: session.EntryPath,
		PageViews: session.PageViews,
		Time:      session.Time,
		SessionID: session.SessionID,
	})
	session = cache.Get(1, 1, time.Now().Add(-time.Second*20))
	assert.NotNil(t, session)
	assert.Equal(t, "/", session.ExitPath)
	assert.Equal(t, "/entry", session.EntryPath)
	assert.Equal(t, uint16(3), session.PageViews)
	cache.Clear()
	cache.Put(1, 1, &Session{
		ExitPath:  session.ExitPath,
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
			ExitPath:  "/foo",
			EntryPath: "/bar",
			PageViews: 42,
		})
	}

	assert.Len(t, cache.sessions, 10)
	session = cache.Get(1, 1, time.Now().Add(-time.Minute))
	assert.NotNil(t, session)
	assert.Equal(t, "/", session.ExitPath)
	cache.Put(1, 10, &Session{
		ExitPath:  "/foo",
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
	assert.Equal(t, "/foo", session.ExitPath)
	cache.Clear()
	assert.Len(t, cache.sessions, 0)
}

func TestSessionCacheMem_Put(t *testing.T) {
	client := NewMockClient()
	cache := NewSessionCacheMem(client, 10)
	now := time.Now()
	cache.Put(1, 1, &Session{
		EntryPath: "/",
		Time:      now,
	})
	now = now.Add(-time.Second)
	cache.Put(1, 1, &Session{
		EntryPath: "/dont-update",
		Time:      now,
	})
	session := cache.Get(1, 1, now.Add(-time.Second*10))
	assert.Equal(t, "/", session.EntryPath)
}
