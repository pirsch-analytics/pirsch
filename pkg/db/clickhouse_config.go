package db

import (
	"log/slog"
	"os"
)

// ClickHouseConfig is the configuration for the ClickHouse Storage implementation.
type ClickHouseConfig struct {
	// Hostnames is the database hostname.
	Hostnames []string

	// Port is the database port.
	Port int

	// Cluster is the optional database cluster to use.
	Cluster string

	// Database is the database schema.
	Database string

	// Username is the database user.
	Username string

	// Password is the database password.
	Password string

	// Secure enables TLS encryption.
	Secure bool

	// SSLSkipVerify skips the SSL verification if set to true.
	SSLSkipVerify bool

	// MaxOpenConnections sets the number of maximum open connections.
	// If set to <= 0, the default value of 20 will be used.
	MaxOpenConnections int

	// MaxConnectionLifetimeSeconds sets the maximum amount of time a connection will be reused.
	// If set to <= 0, the default value of 1800 will be used.
	MaxConnectionLifetimeSeconds int

	// MaxIdleConnections sets the number of maximum idle connections.
	// If set to <= 0, the default value of 5 will be used.
	MaxIdleConnections int

	// MaxConnectionIdleTimeSeconds sets the maximum amount of time a connection can be idle.
	// If set to <= 0, the default value of 300 will be used.
	MaxConnectionIdleTimeSeconds int

	// Logger is the log.Logger used for logging.
	// The default log will be used printing to os.Stdout with "pirsch" in its prefix in case it is not set.
	Logger *slog.Logger

	// Debug will enable verbose logging.
	Debug bool

	dev bool
}

func (config *ClickHouseConfig) validate() {
	if config.MaxOpenConnections <= 0 {
		config.MaxOpenConnections = defaultMaxOpenConnections
	}

	if config.MaxConnectionLifetimeSeconds <= 0 {
		config.MaxConnectionLifetimeSeconds = defaultMaxConnectionLifetime
	}

	if config.MaxIdleConnections <= 0 {
		config.MaxIdleConnections = defaultMaxIdleConnections
	}

	if config.MaxConnectionIdleTimeSeconds <= 0 {
		config.MaxConnectionIdleTimeSeconds = defaultMaxConnectionIdleTime
	}

	if config.Logger == nil {
		config.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
}
