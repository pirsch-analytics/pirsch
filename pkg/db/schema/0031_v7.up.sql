CREATE TABLE session_v7 (
    `sign` Int8,
    `version` UInt16,
    `client_id` UInt64,
    `visitor_id` UInt64,
    `session_id` UInt32,
    `time` DateTime64(3, 'UTC'),
    `start` DateTime64(3, 'UTC'),
    `hostname` String,
    `entry_title` String,
    `exit_title` String,
    `is_bounce` Int8,
    `entry_path` String,
    `exit_path` String,
    `page_views` UInt16,
    `language` LowCardinality(String),
    `country_code` LowCardinality(FixedString(2)),
    `region` String,
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
    `channel` LowCardinality(String),
    `extended` UInt16 DEFAULT 0
)
ENGINE = CollapsingMergeTree(sign)
PARTITION BY toYYYYMM(time)
ORDER BY (client_id, visitor_id, session_id, time)
SAMPLE BY visitor_id
SETTINGS index_granularity = 8192;

CREATE TABLE page_view_v7 (
    `client_id` UInt64,
    `visitor_id` UInt64,
    `session_id` UInt32 DEFAULT 0,
    `time` DateTime64(3, 'UTC'),
    `hostname` String,
    `path` String,
    `title` String,
    `language` LowCardinality(String),
    `country_code` LowCardinality(FixedString(2)),
    `region` LowCardinality(String),
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
    `tags` Map(String, String),
    `channel` LowCardinality(String)
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(time)
ORDER BY (client_id, visitor_id, session_id, time)
SAMPLE BY visitor_id
SETTINGS index_granularity = 8192;

CREATE TABLE event_v7 (
    `client_id` UInt64,
    `visitor_id` UInt64,
    `session_id` UInt32 DEFAULT 0,
    `time` DateTime64(3, 'UTC'),
    `hostname` String,
    `name` String,
    `meta_data` JSON,
    `path` String,
    `title` String,
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
    `desktop` Int8 DEFAULT 0,
    `mobile` Int8 DEFAULT 0,
    `screen_class` LowCardinality(String),
    `utm_source` String,
    `utm_medium` String,
    `utm_campaign` String,
    `utm_content` String,
    `utm_term` String,
    `channel` LowCardinality(String)
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(time)
ORDER BY (client_id, visitor_id, session_id, time)
SAMPLE BY visitor_id
SETTINGS index_granularity = 8192;
