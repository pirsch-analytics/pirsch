package tracker

import (
	util2 "github.com/pirsch-analytics/pirsch/v6/internal/util"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/logger"
	"github.com/pirsch-analytics/pirsch/v6/pkg/tracker/geodb"
	ip2 "github.com/pirsch-analytics/pirsch/v6/pkg/tracker/ip"
	session2 "github.com/pirsch-analytics/pirsch/v6/pkg/tracker/session"
	"log"
	"net"
	"runtime"
	"time"
)

const (
	defaultWorkerBufferSize = 500
	defaultWorkerTimeout    = time.Second * 5
	maxWorkerTimeout        = time.Second * 60
	defaultMaxPageViews     = uint16(200)
)

// Config is the configuration for the Tracker.
type Config struct {
	Store               db.Store
	Salt                string
	FingerprintKey0     uint64
	FingerprintKey1     uint64
	Worker              int
	WorkerBufferSize    int
	WorkerTimeout       time.Duration
	SessionCache        session2.Cache
	HeaderParser        []ip2.HeaderParser
	AllowedProxySubnets []net.IPNet
	MaxPageViews        uint16
	GeoDB               *geodb.GeoDB
	IPFilter            ip2.Filter
	Logger              *log.Logger
}

func (config *Config) validate() {
	if config.Salt == "" {
		config.Salt = util2.RandString(20)
	}

	if config.FingerprintKey0 == 0 {
		config.FingerprintKey0 = util2.RandUint64()
	}

	if config.FingerprintKey1 == 0 {
		config.FingerprintKey1 = util2.RandUint64()
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
		config.SessionCache = session2.NewMemCache(config.Store, 0)
	}

	if config.MaxPageViews == 0 {
		config.MaxPageViews = defaultMaxPageViews
	}

	if config.Logger == nil {
		config.Logger = logger.GetDefaultLogger()
	}
}
