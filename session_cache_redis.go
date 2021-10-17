package pirsch

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"time"
)

// SessionCacheRedis caches sessions in Redis.
type SessionCacheRedis struct {
	maxAge time.Duration
	rds    *redis.Client
}

// NewSessionCacheRedis creates a new cache for given maximum age and redis connection.
func NewSessionCacheRedis(maxAge time.Duration, redisOptions *redis.Options) *SessionCacheRedis {
	return &SessionCacheRedis{
		maxAge: maxAge,
		rds:    redis.NewClient(redisOptions),
	}
}

// Get implements the SessionCache interface.
func (cache *SessionCacheRedis) Get(clientID, fingerprint uint64, maxAge time.Time) *Hit {
	r, err := cache.rds.Get(context.Background(), getSessionKey(clientID, fingerprint)).Result()

	if err != nil {
		return nil
	}

	var hit Hit

	if err := json.Unmarshal([]byte(r), &hit); err != nil {
		return nil
	}

	return &hit
}

// Put implements the SessionCache interface.
func (cache *SessionCacheRedis) Put(clientID, fingerprint uint64, hit *Hit) {
	v, err := json.Marshal(hit)

	if err == nil {
		cache.rds.SetEX(context.Background(), getSessionKey(clientID, fingerprint), v, cache.maxAge)
	}
}

// Clear implements the SessionCache interface.
func (cache *SessionCacheRedis) Clear() {
	cache.rds.FlushDB(context.Background())
}
