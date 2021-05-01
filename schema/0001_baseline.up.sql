CREATE TABLE "hit" (
    client_id UInt64,
    fingerprint FixedString(32),
    time DateTime('UTC'),
    session DateTime('UTC') NULL,
    previous_time_on_page_seconds UInt32 DEFAULT 0,
    user_agent String,
    path String,
    url String,
    language LowCardinality(String),
    country_code LowCardinality(FixedString(2)),
    referrer String NULL,
    referrer_name String NULL,
    referrer_icon String NULL,
    os LowCardinality(String),
    os_version LowCardinality(String),
    browser LowCardinality(String),
    browser_version LowCardinality(String),
    desktop Boolean DEFAULT 0,
    mobile Boolean DEFAULT 0,
    screen_width UInt16 DEFAULT 0,
    screen_height UInt16 DEFAULT 0,
    screen_class LowCardinality(String),
    utm_source String NULL,
    utm_medium String NULL,
    utm_campaign String NULL,
    utm_content String NULL,
    utm_term String NULL
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(time)
ORDER BY (client_id, time)
TTL time + INTERVAL 13 MONTH
;
