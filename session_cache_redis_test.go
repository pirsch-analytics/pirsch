package pirsch

import (
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSessionCacheRedis(t *testing.T) {
	cache := NewSessionCacheRedis(time.Second, nil, &redis.Options{
		Addr: "localhost:6379",
	})
	cache.Clear()
	session := cache.Get(1, 1, time.Time{})
	assert.Nil(t, session)
	cache.Put(1, 1, &Session{ExitPath: "/test"})
	session = cache.Get(1, 1, time.Time{})
	assert.NotNil(t, session)
	assert.Equal(t, "/test", session.ExitPath)
	cache.Clear()
	session = cache.Get(1, 1, time.Time{})
	assert.Nil(t, session)
	cache.Put(1, 1, &Session{ExitPath: "/test"})
	time.Sleep(time.Second * 2)
	session = cache.Get(1, 1, time.Time{})
	assert.Nil(t, session)
}

func TestSessionCacheRedis_PutOlder(t *testing.T) {
	cache := NewSessionCacheRedis(time.Minute, nil, &redis.Options{
		Addr: "localhost:6379",
	})
	cache.Clear()
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
