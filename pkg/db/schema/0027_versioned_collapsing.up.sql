ALTER TABLE "session" ADD COLUMN version UInt16;
ALTER TABLE "page_view" MODIFY COLUMN region LowCardinality(String);
ALTER TABLE "event" MODIFY COLUMN region LowCardinality(String);

CREATE TABLE IF NOT EXISTS "session_new" (
    sign Int8,
    version UInt16,
    client_id UInt64,
    visitor_id UInt64,
    session_id UInt32,
    time DateTime64(3, 'UTC'),
    start DateTime64(3, 'UTC'),
    duration_seconds UInt32,
    is_bounce Int8,
    entry_path String,
    exit_path String,
    entry_title String,
    exit_title String,
    page_views UInt16,
    language LowCardinality(String),
    country_code LowCardinality(FixedString(2)),
    region LowCardinality(String),
    city String,
    referrer String,
    referrer_name String,
    referrer_icon String,
    os LowCardinality(String),
    os_version LowCardinality(String),
    browser LowCardinality(String),
    browser_version LowCardinality(String),
    desktop Int8,
    mobile Int8,
    screen_class LowCardinality(String),
    utm_source String,
    utm_medium String,
    utm_campaign String,
    utm_content String,
    utm_term String,
    extended UInt16 DEFAULT 0
)
ENGINE = VersionedCollapsingMergeTree(sign, version)
PARTITION BY toYYYYMM(time)
ORDER BY (client_id, visitor_id, session_id, time)
SAMPLE BY visitor_id
SETTINGS index_granularity = 8192;

-- Consider running the following steps manually!
--INSERT INTO "session_new" SELECT sign, page_views version, client_id, visitor_id, session_id, time, start, duration_seconds, is_bounce, entry_path, exit_path, entry_title, exit_title, page_views, language, country_code, region, city, referrer, referrer_name, referrer_icon, os, os_version, browser, browser_version, desktop, mobile, screen_class, utm_source, utm_medium, utm_campaign, utm_content, utm_term, extended FROM "session" WHERE toDate(time) < today();
--RENAME TABLE "session" TO "session_backup";
--RENAME TABLE "session_new" TO "session";
--INSERT INTO "session" SELECT sign, page_views version, client_id, visitor_id, session_id, time, start, duration_seconds, is_bounce, entry_path, exit_path, entry_title, exit_title, page_views, language, country_code, region, city, referrer, referrer_name, referrer_icon, os, os_version, browser, browser_version, desktop, mobile, screen_class, utm_source, utm_medium, utm_campaign, utm_content, utm_term, extended FROM "session_backup" WHERE toDate(time) >= today();
--DROP TABLE "session_backup";
