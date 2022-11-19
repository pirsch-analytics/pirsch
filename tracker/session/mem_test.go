package session

import (
	"github.com/pirsch-analytics/pirsch/v5/db"
	"github.com/pirsch-analytics/pirsch/v5/model"
	"github.com/pirsch-analytics/pirsch/v5/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMemCache(t *testing.T) {
	client := db.NewMockClient()
	cache := NewMemCache(client, 10)
	session := cache.Get(1, 1, time.Now().Add(-time.Second*10))
	assert.Nil(t, session)
	client.ReturnSession = &model.Session{
		Time:      time.Now().Add(-time.Second * 15),
		SessionID: util.RandUint32(),
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
	cache.Put(1, 1, &model.Session{
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
	cache.Put(1, 1, &model.Session{
		ExitPath:  session.ExitPath,
		EntryPath: session.EntryPath,
		PageViews: session.PageViews,
		Time:      time.Now().Add(-time.Second * 21),
		SessionID: util.RandUint32(),
	})
	session = cache.Get(1, 1, time.Now().Add(-time.Second*20))
	assert.Nil(t, session)

	for i := 0; i < 9; i++ {
		cache.Put(1, uint64(i+2), &model.Session{
			SessionID: util.RandUint32(),
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
	cache.Put(1, 10, &model.Session{
		ExitPath:  "/foo",
		EntryPath: "/bar",
		PageViews: 42,
		Time:      time.Now(),
		SessionID: util.RandUint32(),
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

func TestMemCache_Put(t *testing.T) {
	client := db.NewMockClient()
	cache := NewMemCache(client, 10)
	now := time.Now()
	cache.Put(1, 1, &model.Session{
		EntryPath: "/",
		Time:      now,
	})
	now = now.Add(-time.Second)
	cache.Put(1, 1, &model.Session{
		EntryPath: "/dont-update",
		Time:      now,
	})
	session := cache.Get(1, 1, now.Add(-time.Second*10))
	assert.Equal(t, "/", session.EntryPath)
}
