CREATE TABLE "event" (
    client_id UInt64,
    fingerprint FixedString(32),
    time DateTime('UTC'),
    session DateTime('UTC'),
    previous_time_on_page_seconds UInt32 DEFAULT 0,
    user_agent String,
    path String,
    url String,
    language LowCardinality(String),
    country_code LowCardinality(FixedString(2)),
    referrer String,
    referrer_name String,
    referrer_icon String,
    os LowCardinality(String),
    os_version LowCardinality(String),
    browser LowCardinality(String),
    browser_version LowCardinality(String),
    desktop Boolean DEFAULT 0,
    mobile Boolean DEFAULT 0,
    screen_width UInt16 DEFAULT 0,
    screen_height UInt16 DEFAULT 0,
    screen_class LowCardinality(String),
    utm_source String,
    utm_medium String,
    utm_campaign String,
    utm_content String,
    utm_term String,
    event_name String,
    event_duration_seconds UInt32 DEFAULT 0,
    event_meta_keys Array(String),
    event_meta_values Array(String)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(time)
ORDER BY (client_id, time)
TTL time + INTERVAL 13 MONTH
;
