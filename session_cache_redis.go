package pirsch

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"log"
	"sync"
	"time"
)

// SessionCacheRedis caches sessions in Redis.
type SessionCacheRedis struct {
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

// NewSessionCacheRedis creates a new cache for given maximum age and redis connection.
func NewSessionCacheRedis(maxAge time.Duration, log *log.Logger, redisOptions *redis.Options) *SessionCacheRedis {
	if log == nil {
		log = logger
	}

	client := redis.NewClient(redisOptions)
	return &SessionCacheRedis{
		maxAge: maxAge,
		rds:    client,
		rs:     redsync.New(goredis.NewPool(client)),
		logger: log,
	}
}

// Get implements the SessionCache interface.
func (cache *SessionCacheRedis) Get(clientID, fingerprint uint64, _ time.Time) *Session {
	r, err := cache.rds.Get(context.Background(), getSessionKey(clientID, fingerprint)).Result()

	if err != nil {
		if err != redis.Nil {
			cache.logger.Printf("error reading session from cache: %s", err)
		}

		return nil
	}

	var session Session

	if err := json.Unmarshal([]byte(r), &session); err != nil {
		cache.logger.Printf("error unmarshalling session from cache: %s", err)
		return nil
	}

	return &session
}

// Put implements the SessionCache interface.
func (cache *SessionCacheRedis) Put(clientID, fingerprint uint64, session *Session) {
	v, err := json.Marshal(session)

	if err == nil {
		cache.rds.SetEX(context.Background(), getSessionKey(clientID, fingerprint), v, cache.maxAge)
	} else {
		cache.logger.Printf("error storing session in cache: %s", err)
	}
}

// Clear implements the SessionCache interface.
func (cache *SessionCacheRedis) Clear() {
	cache.rds.FlushDB(context.Background())
}

// NewMutex implements the SessionCache interface.
func (cache SessionCacheRedis) NewMutex(clientID, fingerprint uint64) sync.Locker {
	return &RedisMutex{cache.rs.NewMutex(getSessionKey(clientID, fingerprint) + "_lock")}
}
