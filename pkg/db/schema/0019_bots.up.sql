CREATE TABLE "bot" (
    `client_id` UInt64,
    `visitor_id` UInt64,
    `time` DateTime64(3, 'UTC'),
    `user_agent` String,
    `path` String,
    `event_name` String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(time)
ORDER BY time
TTL toDateTime(time) + INTERVAL 1 MONTH
;
