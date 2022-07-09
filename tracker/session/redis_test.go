package session

import (
	"github.com/go-redis/redis/v8"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRedisCache(t *testing.T) {
	cache := NewRedisCache(time.Second, nil, &redis.Options{
		Addr: "localhost:6379",
	})
	cache.Clear()
	session := cache.Get(1, 1, time.Time{})
	assert.Nil(t, session)
	cache.Put(1, 1, &model.Session{ExitPath: "/test"})
	session = cache.Get(1, 1, time.Time{})
	assert.NotNil(t, session)
	assert.Equal(t, "/test", session.ExitPath)
	cache.Clear()
	session = cache.Get(1, 1, time.Time{})
	assert.Nil(t, session)
	cache.Put(1, 1, &model.Session{ExitPath: "/test"})
	time.Sleep(time.Second * 2)
	session = cache.Get(1, 1, time.Time{})
	assert.Nil(t, session)
}
