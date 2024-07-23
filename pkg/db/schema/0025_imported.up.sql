CREATE TABLE "imported_browser"
(
    "client_id" UInt64,
    "date" Date,
    "browser" String,
    "visitors" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_city"
(
    "client_id" UInt64,
    "date" Date,
    "city" String,
    "visitors" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_country"
(
    "client_id" UInt64,
    "date" Date,
    "country_code" String,
    "visitors" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_device"
(
    "client_id" UInt64,
    "date" Date,
    "category" String,
    "visitors" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_entry_page"
(
    "client_id" UInt64,
    "date" Date,
    "entry_path" String,
    "visitors" UInt32,
    "sessions" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_exit_page"
(
    "client_id" UInt64,
    "date" Date,
    "exit_path" String,
    "visitors" UInt32,
    "sessions" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_language"
(
    "client_id" UInt64,
    "date" Date,
    "language" String,
    "visitors" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_os"
(
    "client_id" UInt64,
    "date" Date,
    "os" String,
    "visitors" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_page"
(
    "client_id" UInt64,
    "date" Date,
    "path" String,
    "visitors" UInt32,
    "page_views" UInt32,
    "sessions" UInt32,
    "bounces" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_referrer"
(
    "client_id" UInt64,
    "date" Date,
    "referrer" String,
    "visitors" UInt32,
    "sessions" UInt32,
    "bounces" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_region"
(
    "client_id" UInt64,
    "date" Date,
    "region" String,
    "visitors" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_utm_campaign"
(
    "client_id" UInt64,
    "date" Date,
    "utm_campaign" String,
    "visitors" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_utm_medium"
(
    "client_id" UInt64,
    "date" Date,
    "utm_medium" String,
    "visitors" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_utm_source"
(
    "client_id" UInt64,
    "date" Date,
    "utm_source" String,
    "visitors" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE "imported_visitors"
(
    "client_id" UInt64,
    "date" Date,
    "visitors" UInt32,
    "page_views" UInt32,
    "sessions" UInt32,
    "bounces" UInt32,
    "session_duration" UInt32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (client_id, date)
SETTINGS index_granularity = 8192;
