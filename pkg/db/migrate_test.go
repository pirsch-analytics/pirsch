package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigrate(t *testing.T) {
	DropDB(t, dbClient)
	assert.NotNil(t, Migrate(nil))
	assert.NoError(t, Migrate(&ClientConfig{
		Hostnames:     []string{"127.0.0.1"},
		Port:          9000,
		Database:      "pirschtest",
		SSLSkipVerify: true,
		Debug:         true,
	}))
}

func TestParseVersion(t *testing.T) {
	version, err := parseVersion("0001_baseline.up.sql")
	assert.NoError(t, err)
	assert.Equal(t, 1, version)
	version, err = parseVersion("0015_is_bounce.up.sql")
	assert.NoError(t, err)
	assert.Equal(t, 15, version)
	version, err = parseVersion("baseline.up.sql")
	assert.Equal(t, "migration filename needs to start with the version number", err.Error())
	assert.Equal(t, 0, version)
}

func TestParseStatements(t *testing.T) {
	statements, err := parseStatements("0013_remove_backup.up.sql", "")
	assert.NoError(t, err)
	assert.Len(t, statements, 3)
	assert.Equal(t, "DROP TABLE IF EXISTS session_backup", statements[0])
	assert.Equal(t, "DROP TABLE IF EXISTS page_view_backup", statements[1])
	assert.Equal(t, "DROP TABLE IF EXISTS event_backup", statements[2])
}

func TestParseCluster(t *testing.T) {
	statements, err := parseStatements("0029_distributed.up.sql", "pirsch")
	assert.NoError(t, err)
	assert.Len(t, statements, 27)
	assert.Equal(t, "DROP TABLE IF EXISTS schema_migrations", statements[0])
	assert.Equal(t, `CREATE TABLE schema_migrations ON CLUSTER 'pirsch' (
        `+"`version`"+` Int64,
        `+"`dirty`"+` UInt8,
        `+"`sequence`"+` UInt64
    )
    ENGINE = ReplicatedMergeTree
    ORDER BY (version, dirty, sequence)`, statements[1])
	assert.Equal(t, `CREATE TABLE IF NOT EXISTS session ON CLUSTER 'pirsch' (
    `+"`sign`"+` Int8,
    `+"`version`"+` UInt16,
    `+"`client_id`"+` UInt64,
    `+"`visitor_id`"+` UInt64,
    `+"`session_id`"+` UInt32,
    `+"`time`"+` DateTime64(3, 'UTC'),
    `+"`start`"+` DateTime64(3, 'UTC'),
    `+"`duration_seconds`"+` UInt32,
    `+"`is_bounce`"+` Int8,
    `+"`entry_path`"+` String,
    `+"`exit_path`"+` String,
    `+"`entry_title`"+` String,
    `+"`exit_title`"+` String,
    `+"`page_views`"+` UInt16,
    `+"`language`"+` LowCardinality(String),
    `+"`country_code`"+` LowCardinality(FixedString(2)),
    `+"`region`"+` LowCardinality(String),
    `+"`city`"+` String,
    `+"`referrer`"+` String,
    `+"`referrer_name`"+` String,
    `+"`referrer_icon`"+` String,
    `+"`os`"+` LowCardinality(String),
    `+"`os_version`"+` LowCardinality(String),
    `+"`browser`"+` LowCardinality(String),
    `+"`browser_version`"+` LowCardinality(String),
    `+"`desktop`"+` Int8,
    `+"`mobile`"+` Int8,
    `+"`screen_class`"+` LowCardinality(String),
    `+"`utm_source`"+` String,
    `+"`utm_medium`"+` String,
    `+"`utm_campaign`"+` String,
    `+"`utm_content`"+` String,
    `+"`utm_term`"+` String,
    `+"`extended`"+` UInt16 DEFAULT 0,
    `+"`hostname`"+` String
)
ENGINE = ReplicatedVersionedCollapsingMergeTree('/clickhouse/tables/session/{shard}', '{replica}', sign, version)
PARTITION BY toYYYYMM(time)
ORDER BY (client_id, visitor_id, session_id, time)
SAMPLE BY visitor_id
SETTINGS index_granularity = 8192`, statements[8])
	assert.Equal(t, `CREATE TABLE IF NOT EXISTS imported_browser ON CLUSTER 'pirsch' (
    `+"`client_id`"+` UInt64,
    `+"`date`"+` Date,
    `+"`browser`"+` String,
    `+"`visitors`"+` UInt32
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/imported_browser/{shard}', '{replica}')
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192`, statements[12])
}
