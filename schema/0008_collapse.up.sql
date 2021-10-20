DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS events;

CREATE TABLE session
(
    `sign` Int8,
    `client_id` UInt64,
    `visitor_id` UInt64,
    `session_id` UInt32,
    `time` DateTime('UTC'),
    `start` DateTime('UTC'),
    `duration_seconds` UInt32 DEFAULT 0,
    `path` String,
    `title` String,
    `is_bounce` UInt8 DEFAULT 1,
    `entry_path` String DEFAULT '',
    `page_views` UInt16 DEFAULT 1,
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
    `screen_width` UInt16 DEFAULT 0,
    `screen_height` UInt16 DEFAULT 0,
    `screen_class` LowCardinality(String),
    `utm_source` String DEFAULT '',
    `utm_medium` String DEFAULT '',
    `utm_campaign` String DEFAULT '',
    `utm_content` String DEFAULT '',
    `utm_term` String DEFAULT ''
)
    ENGINE = CollapsingMergeTree(sign)
PARTITION BY toYYYYMM(time)
ORDER BY (client_id, visitor_id, session_id, time)
TTL time + toIntervalMonth(12)
SETTINGS index_granularity = 8192
;
