ALTER TABLE "page_view" MODIFY SAMPLE BY "visitor_id";
ALTER TABLE "session" MODIFY SAMPLE BY "visitor_id";
RENAME TABLE "event" TO "event_backup";

CREATE TABLE "event"
(
    "client_id" UInt64,
    "time" DateTime64(3, 'UTC'),
    "duration_seconds" UInt32 DEFAULT 0,
    "path" String,
    "language" LowCardinality(String),
    "country_code" LowCardinality(FixedString(2)),
    "referrer" String,
    "referrer_name" String,
    "referrer_icon" String,
    "os" LowCardinality(String),
    "os_version" LowCardinality(String),
    "browser" LowCardinality(String),
    "browser_version" LowCardinality(String),
    "desktop" Int8 DEFAULT 0,
    "mobile" Int8 DEFAULT 0,
    "screen_class" LowCardinality(String),
    "utm_source" String,
    "utm_medium" String,
    "utm_campaign" String,
    "utm_content" String,
    "utm_term" String,
    "event_name" String,
    "event_meta_keys" Array(String),
    "event_meta_values" Array(String),
    "title" String,
    "session_id" UInt32 DEFAULT 0,
    "city" String,
    "visitor_id" UInt64
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(time)
ORDER BY (client_id, visitor_id, session_id, time)
SAMPLE BY "visitor_id"
SETTINGS index_granularity = 8192;
