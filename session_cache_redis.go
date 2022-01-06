package pirsch

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

const (
	rdsPutLua = `if redis.call("EXISTS", KEYS[2]) == 0 or tonumber(redis.call("GET", KEYS[2])) <= tonumber(ARGV[3]) then
	redis.call("SETEX", KEYS[2], ARGV[2], ARGV[3])
	return redis.call("SETEX", KEYS[1], ARGV[2], ARGV[1])
end

return ""`
)

// SessionCacheRedis caches sessions in Redis.
type SessionCacheRedis struct {
	maxAge time.Duration
	rds    *redis.Client
	logger *log.Logger
}

// NewSessionCacheRedis creates a new cache for given maximum age and redis connection.
func NewSessionCacheRedis(maxAge time.Duration, log *log.Logger, redisOptions *redis.Options) *SessionCacheRedis {
	if log == nil {
		log = logger
	}

	return &SessionCacheRedis{
		maxAge: maxAge,
		rds:    redis.NewClient(redisOptions),
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
		key := getSessionKey(clientID, fingerprint)
		cmd := cache.rds.Eval(context.Background(), rdsPutLua, []string{key, key + "_time"}, string(v), int(cache.maxAge.Seconds()), session.Time.UnixMilli())

		if err := cmd.Err(); err != nil {
			cache.logger.Printf("error storing session in cache: %s", err)
		}
	} else {
		cache.logger.Printf("error marshalling session: %s", err)
	}
}

// Clear implements the SessionCache interface.
func (cache *SessionCacheRedis) Clear() {
	cache.rds.FlushDB(context.Background())
}
