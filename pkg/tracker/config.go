package tracker

import (
	"github.com/pirsch-analytics/pirsch/v6/internal/util"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/tracker/geodb"
	"github.com/pirsch-analytics/pirsch/v6/pkg/tracker/ip"
	"github.com/pirsch-analytics/pirsch/v6/pkg/tracker/session"
	"log/slog"
	"net"
	"os"
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
	SessionCache        session.Cache
	HeaderParser        []ip.HeaderParser
	AllowedProxySubnets []net.IPNet
	MaxPageViews        uint16
	GeoDB               *geodb.GeoDB
	IPFilter            ip.Filter
	Logger              *slog.Logger
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

	if config.MaxPageViews == 0 {
		config.MaxPageViews = defaultMaxPageViews
	}

	if config.Logger == nil {
		config.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
}
