DROP DATABASE IF EXISTS app;
CREATE DATABASE app;
USE app;

CREATE TABLE events (
    ts DATETIME NOT NULL,
    path TEXT NOT NULL COLLATE "utf8_bin",
    user_id TEXT NOT NULL COLLATE "utf8_bin",

    referrer TEXT,
    page_time_s DOUBLE NOT NULL DEFAULT 0,

    SORT KEY (ts),
    SHARD KEY (user_id),

    KEY (path),
    KEY (user_id),
    KEY (referrer)
);

CREATE PIPELINE events
AS LOAD DATA KAFKA 'redpanda/events'
SKIP DUPLICATE KEY ERRORS
INTO TABLE events
FORMAT JSON (
    @unix_timestamp <- unix_timestamp,
    path <- path,
    user_id <- user_id,
    referrer <- referrer DEFAULT NULL,
    page_time_s <- page_time_seconds DEFAULT 0
)
SET ts = FROM_UNIXTIME(@unix_timestamp);

START PIPELINE events;