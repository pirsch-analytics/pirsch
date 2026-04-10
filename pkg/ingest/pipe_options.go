package ingest

import (
	"log/slog"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
)

// PipeOptions is the configuration for a Pipe.
type PipeOptions struct {
	// Storage is the data storage to be used for this Pipe.
	Storage db.Storage

	// RequestChannelBufferSize is the channel buffer size for the data ingestion.
	// 0 means unbuffered.
	RequestChannelBufferSize int

	// Worker is the number of workers collecting and storing requests in batch.
	Worker int

	// WorkerBufferSize sets the size for the data ingestion buffer per type (session, page view, event, request) per Worker.
	WorkerBufferSize int

	// WorkerTimeout sets the maximum waiting time before the worker buffers are flushed.
	WorkerTimeout time.Duration

	// Logger is the logger for the Pipe.
	// If not set, the default slog.Logger will be used.
	Logger *slog.Logger

	/* TODO
	defaultMaxPageViews     = uint16(200)

	Salt                string
	FingerprintKey0     uint64
	FingerprintKey1     uint64
	SessionCache        session.Cache

	HeaderParser        []ip.HeaderParser
	AllowedProxySubnets []net.IPNet
	MaxPageViews        uint16
	GeoDB               *geodb.GeoDB
	IPFilter            []ip.Filter
	LogIP               bool
	*/
}

func (options *PipeOptions) validate() {
	if options.RequestChannelBufferSize < 0 {
		options.RequestChannelBufferSize = 0
	}

	if options.Worker < 1 {
		options.Worker = 10
	}

	if options.WorkerBufferSize < 1 {
		options.WorkerBufferSize = 100
	}

	if options.WorkerTimeout <= 0 {
		options.WorkerTimeout = time.Second * 5
	}

	if options.Logger == nil {
		options.Logger = slog.Default()
	}
}
