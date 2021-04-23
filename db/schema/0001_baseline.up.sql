CREATE TABLE "hit" (
    tenant_id UInt64 NULL,
    fingerprint FixedString(32),
    time DateTime('UTC'),
    session DateTime('UTC') NULL,
    path String,
    url String NULL,
    language FixedString(10) NULL,
    country_code FixedString(2) NULL,
    user_agent String,
    referrer String NULL,
    referrer_name String NULL,
    referrer_icon String NULL,
    os FixedString(20) NULL,
    os_version FixedString(20) NULL,
    browser FixedString(20) NULL,
    browser_version FixedString(20) NULL,
    desktop Boolean DEFAULT 0,
    mobile Boolean DEFAULT 0,
    screen_width UInt16 NULL,
    screen_height UInt16 NULL,
    screen_class FixedString(5) NULL
) ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(time)
ORDER BY (time)
;
