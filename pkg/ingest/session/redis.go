package session

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
)

// RedisCache caches sessions in Redis.
type RedisCache struct {
	maxAge time.Duration
	rds    *redis.Client
	rs     *redsync.Redsync
	logger *slog.Logger
}

// RedisMutex wraps a Redis mutex.
// Errors are logged instead of causing a panic, so that the request is processed without a lock rather than being dropped.
type RedisMutex struct {
	m      *redsync.Mutex
	logger *slog.Logger
	locked bool
}

// Lock implements the sync.Locker interface.
func (m *RedisMutex) Lock() {
	if err := m.m.Lock(); err != nil {
		m.locked = false
		m.logger.Warn("error acquiring session lock, continuing without lock", "err", err)
		return
	}

	m.locked = true
}

// Unlock implements the sync.Locker interface.
func (m *RedisMutex) Unlock() {
	if !m.locked {
		return
	}

	m.locked = false

	if _, err := m.m.Unlock(); err != nil {
		m.logger.Warn("error releasing session lock", "err", err)
	}
}

// NewRedisCache creates a new cache for a given maximum age and redis connection.
func NewRedisCache(maxAge time.Duration, log *slog.Logger, redisOptions *redis.Options) *RedisCache {
	if log == nil {
		log = slog.New(slog.NewTextHandler(os.Stdout, nil))
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
func (cache *RedisCache) Get(siteID, fingerprint uint64, _ time.Time) *model.Session {
	r, err := cache.rds.Get(context.Background(), getSessionKey(siteID, fingerprint)).Result()

	if err != nil {
		if !errors.Is(err, redis.Nil) {
			cache.logger.Error("error reading session from cache", "err", err)
		}

		return nil
	}

	var session model.Session

	if err := json.Unmarshal([]byte(r), &session); err != nil {
		cache.logger.Error("error unmarshalling session from cache", "err", err)
		return nil
	}

	return &session
}

// Put implements the Cache interface.
func (cache *RedisCache) Put(siteID, fingerprint uint64, session *model.Session) {
	v, err := json.Marshal(session)

	if err == nil {
		cache.rds.SetEX(context.Background(), getSessionKey(siteID, fingerprint), v, cache.maxAge)
	} else {
		cache.logger.Error("error storing session in cache", "err", err)
	}
}

// Clear implements the Cache interface.
func (cache *RedisCache) Clear() {
	cache.rds.FlushDB(context.Background())
}

// NewMutex implements the Cache interface.
func (cache *RedisCache) NewMutex(siteID, fingerprint uint64) sync.Locker {
	return &RedisMutex{
		m:      cache.rs.NewMutex(getSessionKey(siteID, fingerprint) + "_lock"),
		logger: cache.logger,
	}
}
