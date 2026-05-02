DROP TABLE IF EXISTS schema_migrations;

{{if .Cluster}}
    CREATE TABLE schema_migrations ON CLUSTER '{{.Cluster}}' (
        `version` Int64,
        `dirty` UInt8,
        `sequence` UInt64
    )
    ENGINE = ReplicatedMergeTree('/clickhouse/tables/schema_migrations/{shard}', '{replica}')
    ORDER BY (version, dirty, sequence);
{{else}}
    CREATE TABLE schema_migrations (
        `version` Int64,
        `dirty` UInt8,
        `sequence` UInt64
    )
    ENGINE = MergeTree
    ORDER BY (version, dirty, sequence);
{{end}}

INSERT INTO schema_migrations (version,dirty,`sequence`) VALUES
    (1,1,1627044237936558313),
    (2,1,1627044237936558314),
    (2,0,1627044237936558315),
    (3,1,1627044443815839664),
    (3,0,1627044443832227059),
    (4,1,1628879857177110063),
    (4,0,1628879857211879922),
    (5,1,1628879857211879923),
    (5,0,1628879857211879924),
    (6,1,1634495187649649615);
INSERT INTO schema_migrations (version,dirty,`sequence`) VALUES
    (6,0,1634495187688630930),
    (7,1,1634495187689307798),
    (7,0,1634495187756461853),
    (8,1,1634760088899694354),
    (8,0,1634760088916237381),
    (9,1,1635025082915362366),
    (9,0,1635025082996626369),
    (10,1,1636307209782745207),
    (10,0,1636307209791457594),
    (11,1,1639328035100645659);
INSERT INTO schema_migrations (version,dirty,`sequence`) VALUES
    (11,0,1639328035113363102),
    (12,1,1641478343943984265),
    (12,0,1641478343987970071),
    (13,1,1642589863670455065),
    (13,0,1642589863676257074),
    (14,1,1648901537194552324),
    (14,0,1648901537202960084),
    (15,1,1656102949738857472),
    (15,0,1656102950838047711),
    (16,1,1669564586889254918);
INSERT INTO schema_migrations (version,dirty,`sequence`) VALUES
    (16,0,1669564586899374777),
    (17,1,1674393693813311436),
    (17,0,1674393749110382037),
    (18,1,1688041412005239275),
    (18,0,1688041433418567512),
    (19,1,1688041433424726093),
    (19,0,1688041433432099907),
    (20,1,1697731664468370121),
    (20,0,1697731664483019345),
    (21,1,1707922044168350235);
INSERT INTO schema_migrations (version,dirty,`sequence`) VALUES
    (21,0,1707922044184696701),
    (22,1,1707922044188086512),
    (22,0,1707922044194909535),
    (23,1,1713201870340537092),
    (23,0,1713201870357926261),
    (24,1,1716288505219745188),
    (24,0,1716288505234836141),
    (25,1,1722338707764296978),
    (25,0,1722338707775632972),
    (26,1,1723822636719854153);
INSERT INTO schema_migrations (version,dirty,`sequence`) VALUES
    (26,0,1723822636774870083),
    (27,1,1724264508398494568),
    (27,0,1724264549610560816),
    (28,1,1727301959603710364),
    (28,0,1727301959628330304),
    (29,1,1727301959628330305);

CREATE TABLE IF NOT EXISTS session {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `sign` Int8,
    `version` UInt16,
    `client_id` UInt64,
    `visitor_id` UInt64,
    `session_id` UInt32,
    `time` DateTime64(3, 'UTC'),
    `start` DateTime64(3, 'UTC'),
    `duration_seconds` UInt32,
    `is_bounce` Int8,
    `entry_path` String,
    `exit_path` String,
    `entry_title` String,
    `exit_title` String,
    `page_views` UInt16,
    `language` LowCardinality(String),
    `country_code` LowCardinality(FixedString(2)),
    `region` LowCardinality(String),
    `city` String,
    `referrer` String,
    `referrer_name` String,
    `referrer_icon` String,
    `os` LowCardinality(String),
    `os_version` LowCardinality(String),
    `browser` LowCardinality(String),
    `browser_version` LowCardinality(String),
    `desktop` Int8,
    `mobile` Int8,
    `screen_class` LowCardinality(String),
    `utm_source` String,
    `utm_medium` String,
    `utm_campaign` String,
    `utm_content` String,
    `utm_term` String,
    `extended` UInt16 DEFAULT 0,
    `hostname` String
)
ENGINE = {{if .Cluster}}ReplicatedVersionedCollapsingMergeTree('/clickhouse/tables/session/{shard}', '{replica}', sign, version){{else}}VersionedCollapsingMergeTree(sign, version){{end}}
PARTITION BY toYYYYMM(time)
ORDER BY (client_id, visitor_id, session_id, time)
SAMPLE BY visitor_id
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS page_view {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `visitor_id` UInt64,
    `session_id` UInt32 DEFAULT 0,
    `time` DateTime64(3, 'UTC'),
    `duration_seconds` UInt32 DEFAULT 0,
    `path` String,
    `title` String,
    `language` LowCardinality(String),
    `country_code` LowCardinality(FixedString(2)),
    `city` String,
    `referrer` String DEFAULT '',
    `referrer_name` String DEFAULT '',
    `referrer_icon` String DEFAULT '',
    `os` LowCardinality(String),
    `os_version` LowCardinality(String),
    `browser` LowCardinality(String),
    `browser_version` LowCardinality(String),
    `desktop` Int8 DEFAULT 0,
    `mobile` Int8 DEFAULT 0,
    `screen_class` LowCardinality(String),
    `utm_source` String DEFAULT '',
    `utm_medium` String DEFAULT '',
    `utm_campaign` String DEFAULT '',
    `utm_content` String DEFAULT '',
    `utm_term` String DEFAULT '',
    `tag_keys` Array(String),
    `tag_values` Array(String),
    `region` LowCardinality(String),
    `hostname` String
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/page_view/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(time)
ORDER BY (client_id, visitor_id, session_id, time)
SAMPLE BY visitor_id
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS event {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `time` DateTime64(3, 'UTC'),
    `duration_seconds` UInt32 DEFAULT 0,
    `path` String,
    `language` LowCardinality(String),
    `country_code` LowCardinality(FixedString(2)),
    `referrer` String,
    `referrer_name` String,
    `referrer_icon` String,
    `os` LowCardinality(String),
    `os_version` LowCardinality(String),
    `browser` LowCardinality(String),
    `browser_version` LowCardinality(String),
    `desktop` Int8 DEFAULT 0,
    `mobile` Int8 DEFAULT 0,
    `screen_class` LowCardinality(String),
    `utm_source` String,
    `utm_medium` String,
    `utm_campaign` String,
    `utm_content` String,
    `utm_term` String,
    `event_name` String,
    `event_meta_keys` Array(String),
    `event_meta_values` Array(String),
    `title` String,
    `session_id` UInt32 DEFAULT 0,
    `city` String,
    `visitor_id` UInt64,
    `region` LowCardinality(String),
    `hostname` String
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/event/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(time)
ORDER BY (client_id, visitor_id, session_id, time)
SAMPLE BY visitor_id
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS request {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `visitor_id` UInt64,
    `time` DateTime64(3, 'UTC'),
    `user_agent` String,
    `path` String,
    `event_name` String,
    `bot` Bool DEFAULT 1,
    `ip` String,
    `referrer` String,
    `utm_source` String,
    `utm_medium` String,
    `utm_campaign` String,
    `bot_reason` String,
    `hostname` String
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/request/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(time)
ORDER BY time
TTL toDateTime(time) + toIntervalMonth(1)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_browser {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `browser` String,
    `visitors` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_browser/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_city {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `city` String,
    `visitors` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_city/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_country {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `country_code` String,
    `visitors` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_country/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_device {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `category` String,
    `visitors` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_device/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_entry_page {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `entry_path` String,
    `visitors` UInt32,
    `sessions` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_entry_page/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_exit_page {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `exit_path` String,
    `visitors` UInt32,
    `sessions` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_exit_page/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_language {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `language` String,
    `visitors` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_language/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_os {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `os` String,
    `visitors` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_os/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_page {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `path` String,
    `visitors` UInt32,
    `views` UInt32,
    `sessions` UInt32,
    `bounces` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_page/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_referrer {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `referrer` String,
    `visitors` UInt32,
    `sessions` UInt32,
    `bounces` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_referrer/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_region {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `region` String,
    `visitors` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_region/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_utm_campaign {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `utm_campaign` String,
    `visitors` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_utm_campaign/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_utm_medium {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `utm_medium` String,
    `visitors` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_utm_medium/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_utm_source {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `utm_source` String,
    `visitors` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_utm_source/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS imported_visitors {{if .Cluster}}ON CLUSTER '{{.Cluster}}'{{end}} (
    `client_id` UInt64,
    `date` Date,
    `visitors` UInt32,
    `views` UInt32,
    `sessions` UInt32,
    `bounces` UInt32,
    `session_duration` UInt32
)
ENGINE = {{if .Cluster}}ReplicatedMergeTree('/clickhouse/tables/imported_visitors/{shard}', '{replica}'){{else}}MergeTree{{end}}
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;
