package pirsch

import (
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSessionCacheRedis(t *testing.T) {
	cleanupDB()
	cache := NewSessionCacheRedis(time.Second, &redis.Options{
		Addr: "localhost:6379",
	})
	session := cache.Get(1, "fp", time.Time{})
	assert.Nil(t, session)
	cache.Put(1, "fp", &Hit{Path: "/test"})
	session = cache.Get(1, "fp", time.Time{})
	assert.NotNil(t, session)
	assert.Equal(t, "/test", session.Path)
	cache.Clear()
	session = cache.Get(1, "fp", time.Time{})
	assert.Nil(t, session)
	cache.Put(1, "fp", &Hit{Path: "/test"})
	time.Sleep(time.Second)
	session = cache.Get(1, "fp", time.Time{})
	assert.Nil(t, session)
}
