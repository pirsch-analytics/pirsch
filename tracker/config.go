package tracker

import (
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/tracker_/geodb"
	"github.com/pirsch-analytics/pirsch/v4/tracker_/ip"
	"github.com/pirsch-analytics/pirsch/v4/tracker_/session"
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
