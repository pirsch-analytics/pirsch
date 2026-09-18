{{ if .Cluster }}
    CREATE TABLE session_v7 ON CLUSTER '{{.Cluster}}' (
{{ else }}
   CREATE TABLE session_v7 (
{{ end }}
    `sign` Int8,
    `version` UInt16,
    `site_id` UInt64,
    `visitor_id` UInt64,
    `session_id` UInt32,
    `time` DateTime64(3, 'UTC'),
    `start` DateTime64(3, 'UTC'),
    `duration_seconds` UInt32,
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
    `platform` Int8 DEFAULT 0,
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
ORDER BY (site_id, visitor_id, session_id, time)
SAMPLE BY visitor_id
SETTINGS index_granularity = 8192;

{{ if .Cluster }}
    CREATE TABLE page_view_v7 ON CLUSTER '{{.Cluster}}' (
{{ else }}
    CREATE TABLE page_view_v7 (
{{ end }}
    `site_id` UInt64,
    `visitor_id` UInt64,
    `session_id` UInt32 DEFAULT 0,
    `time` DateTime64(3, 'UTC'),
    `duration_seconds` UInt32,
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
    `platform` Int8 DEFAULT 0,
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
ORDER BY (site_id, visitor_id, session_id, time)
SAMPLE BY visitor_id
SETTINGS index_granularity = 8192;

{{ if .Cluster }}
    CREATE TABLE event_v7 ON CLUSTER '{{.Cluster}}' (
{{ else }}
    CREATE TABLE event_v7 (
{{ end }}
    `site_id` UInt64,
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
    `platform` Int8 DEFAULT 0,
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
ORDER BY (site_id, visitor_id, session_id, time)
SAMPLE BY visitor_id
SETTINGS index_granularity = 8192;

{{ if .Cluster }}
    CREATE TABLE request_v7 ON CLUSTER '{{.Cluster}}' (
{{ else }}
    CREATE TABLE request_v7 (
{{ end }}
    `site_id` UInt64,
    `visitor_id` UInt64,
    `time` DateTime64(3, 'UTC'),
    `hostname` String,
    `path` String,
    `query` String,
    `ip` String,
    `user_agent` String,
    `headers` Map(LowCardinality(String), String),
    `event_name` String,
    `referrer` String,
    `utm_source` String,
    `utm_medium` String,
    `utm_campaign` String,
    `utm_content` String,
    `utm_term` String,
    `bot` Bool DEFAULT 1,
    `bot_reason` String
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(time)
ORDER BY time
TTL toDateTime(time) + toIntervalMonth(1)
SETTINGS index_granularity = 8192;

ALTER TABLE "imported_browser" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_city" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_country" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_device" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_entry_page" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_exit_page" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_language" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_os" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_page" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_referrer" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_region" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_utm_campaign" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_utm_medium" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_utm_source" ADD COLUMN site_id UInt64 ALIAS client_id;
ALTER TABLE "imported_visitors" ADD COLUMN site_id UInt64 ALIAS client_id;
