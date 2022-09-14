package session

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/pirsch-analytics/pirsch/v4/util"
	"log"
	"sync"
	"time"
)

// RedisCache caches sessions in Redis.
type RedisCache struct {
	maxAge time.Duration
	rds    *redis.Client
	rs     *redsync.Redsync
	logger *log.Logger
}

// RedisMutex wraps a redis mutex.
type RedisMutex struct {
	m *redsync.Mutex
}

func (m *RedisMutex) Lock() {
	if err := m.m.Lock(); err != nil {
		panic(err)
	}
}

func (m *RedisMutex) Unlock() {
	if _, err := m.m.Unlock(); err != nil {
		panic(err)
	}
}

// NewRedisCache creates a new cache for given maximum age and redis connection.
func NewRedisCache(maxAge time.Duration, log *log.Logger, redisOptions *redis.Options) *RedisCache {
	if log == nil {
		log = util.GetDefaultLogger()
	}

	client := redis.NewClient(redisOptions)
	return &RedisCache{
		maxAge: maxAge,
		rds:    client,
		rs:     redsync.New(goredis.NewPool(client)),
		logger: log,
	}
}

// Get implements the Cache interface.
func (cache *RedisCache) Get(clientID, fingerprint uint64, _ time.Time) *model.Session {
	r, err := cache.rds.Get(context.Background(), getSessionKey(clientID, fingerprint)).Result()

	if err != nil {
		if err != redis.Nil {
			cache.logger.Printf("error reading session from cache: %s", err)
		}

		return nil
	}

	var session model.Session

	if err := json.Unmarshal([]byte(r), &session); err != nil {
		cache.logger.Printf("error unmarshalling session from cache: %s", err)
		return nil
	}

	return &session
}

// Put implements the Cache interface.
func (cache *RedisCache) Put(clientID, fingerprint uint64, session *model.Session) {
	v, err := json.Marshal(session)

	if err == nil {
		cache.rds.SetEX(context.Background(), getSessionKey(clientID, fingerprint), v, cache.maxAge)
	} else {
		cache.logger.Printf("error storing session in cache: %s", err)
	}
}

// Clear implements the Cache interface.
func (cache *RedisCache) Clear() {
	cache.rds.FlushDB(context.Background())
}

// NewMutex implements the Cache interface.
func (cache *RedisCache) NewMutex(clientID, fingerprint uint64) sync.Locker {
	return &RedisMutex{cache.rs.NewMutex(getSessionKey(clientID, fingerprint) + "_lock")}
}
