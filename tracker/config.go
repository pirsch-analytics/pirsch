package tracker

import (
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/tracker/geodb"
	"github.com/pirsch-analytics/pirsch/v4/tracker/ip"
	"github.com/pirsch-analytics/pirsch/v4/tracker/session"
	"github.com/pirsch-analytics/pirsch/v4/util"
	"log"
	"net"
	"runtime"
	"time"
)

const (
	defaultWorkerBufferSize = 500
	defaultWorkerTimeout    = time.Second * 5
	maxWorkerTimeout        = time.Second * 60
	defaultMinDelayMS       = 75
	defaultIsBotThreshold   = 5
)

// Config is the configuration for the Tracker.
type Config struct {
	Store                   db.Store
	Salt                    string
	FingerprintKey0         uint64
	FingerprintKey1         uint64
	Worker                  int
	WorkerBufferSize        int
	WorkerTimeout           time.Duration
	ReferrerDomainBlacklist []string
	SessionCache            session.Cache
	HeaderParser            []ip.HeaderParser
	AllowedProxySubnets     []net.IPNet
	MinDelay                int64
	IsBotThreshold          uint8
	GeoDB                   *geodb.GeoDB
	Logger                  *log.Logger
}

func (config *Config) validate() {
	if config.Salt == "" {
		config.Salt = util.RandString(20)
	}

	if config.FingerprintKey0 == 0 {
		config.FingerprintKey0 = util.RandUint64()
	}

	if config.FingerprintKey1 == 0 {
		config.FingerprintKey1 = util.RandUint64()
	}

	if config.Worker < 1 {
		config.Worker = runtime.NumCPU()
	}

	if config.WorkerBufferSize < 1 {
		config.WorkerBufferSize = defaultWorkerBufferSize
	}

	if config.WorkerTimeout <= 0 {
		config.WorkerTimeout = defaultWorkerTimeout
	} else if config.WorkerTimeout > maxWorkerTimeout {
		config.WorkerTimeout = maxWorkerTimeout
	}

	if config.SessionCache == nil {
		config.SessionCache = session.NewMemCache(config.Store, 0)
	}

	if config.MinDelay <= 0 {
		config.MinDelay = defaultMinDelayMS
	}

	if config.IsBotThreshold == 0 {
		config.IsBotThreshold = defaultIsBotThreshold
	}

	if config.Logger == nil {
		config.Logger = util.GetDefaultLogger()
	}
}
