-- stores the ingested and recent metrics at a high precision (for 30 days)
CREATE TABLE yamon.metrics (
	`when` DateTime64(9) CODEC(Delta(8), ZSTD(1)),
	`type` Enum8('gauge' = 1, 'counter' = 2) CODEC(ZSTD(1)),
	`host` LowCardinality(String) CODEC(ZSTD(1)),
	`name` LowCardinality(String) CODEC(ZSTD(1)),
	`value` Float64 CODEC(ZSTD(1)),
	`tags` Map(LowCardinality(String), String) CODEC(ZSTD(1)),
	INDEX idx_tag_key mapKeys(tags) TYPE bloom_filter(0.01) GRANULARITY 1,
	INDEX idx_tag_value mapValues(tags) TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = MergeTree
PARTITION BY toDate(when)
ORDER BY (name, host, tags, toUnixTimestamp64Nano(when))
TTL toDateTime(when) + toIntervalDay(30)
SETTINGS 
	index_granularity = 8192,
	ttl_only_drop_parts = 1;


-- stores aggregated (averaged) gauge metrics for a full year at 1 minute resolution
CREATE TABLE yamon.metrics_gauge_lts (
	`when` DateTime64(3) CODEC(Delta, ZSTD(1)),
	`host` LowCardinality(String) CODEC(ZSTD(1)),
	`name` LowCardinality(String) CODEC(ZSTD(1)),
	`value` Float64 CODEC(ZSTD(1)),
	`tags` Map(LowCardinality(String), String) CODEC(ZSTD(1)),
	INDEX idx_tag_key mapKeys(tags) TYPE bloom_filter(0.01) GRANULARITY 1,
	INDEX idx_tag_value mapValues(tags) TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = MergeTree
PARTITION BY toDate(when)
ORDER BY (name, host, tags, toUnixTimestamp(when))
TTL toDateTime(when) + INTERVAL 1 YEAR
SETTINGS 
	index_granularity = 8192,
	ttl_only_drop_parts = 1;

CREATE MATERIALIZED VIEW yamon.metrics_gauge_lts_mv
TO yamon.metrics_gauge_lts
AS
SELECT
	toStartOfInterval(when, INTERVAL 1 minute) as when,
	host,
	name,
	avg(value) as value,
	tags
FROM yamon.metrics
WHERE `type` = 'gauge'
GROUP BY when, host, name, tags;


-- stores aggregated (summed) counter metrics for a full year at 1 minute resolution
CREATE TABLE yamon.metrics_counter_lts (
	`when` DateTime64(3) CODEC(Delta(8), ZSTD(1)),
	`host` LowCardinality(String) CODEC(ZSTD(1)),
	`name` LowCardinality(String) CODEC(ZSTD(1)),
	`value` Float64 CODEC(Delta(8), ZSTD(1)),
	`tags` Map(LowCardinality(String), String) CODEC(ZSTD(1)),
	INDEX idx_tag_key mapKeys(tags) TYPE bloom_filter(0.01) GRANULARITY 1,
	INDEX idx_tag_value mapValues(tags) TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = MergeTree
PARTITION BY toDate(when)
ORDER BY (name, host, tags, toUnixTimestamp(when))
TTL toDateTime(when) + INTERVAL 1 YEAR
SETTINGS 
	index_granularity = 8192,
	ttl_only_drop_parts = 1;

CREATE MATERIALIZED VIEW yamon.metrics_counter_lts_mv
TO yamon.metrics_counter_lts
AS
SELECT
	toStartOfInterval(when, INTERVAL 1 minute) as when,
	host,
	name,
	sum(value) as value,
	tags
FROM yamon.metrics
WHERE `type` = 'counter'
GROUP BY when, host, name, tags;


-- stores logs for 30 days
CREATE TABLE yamon.logs (
	`when` DateTime64(9) CODEC(Delta(8), ZSTD(1)),
	`host` LowCardinality(String) CODEC(ZSTD(1)),
	`service` LowCardinality(String) CODEC(ZSTD(1)),
	`level` LowCardinality(String) CODEC(ZSTD(1)),
	`data` String CODEC(ZSTD(2)),
	`tags` Map(LowCardinality(String), String) CODEC(ZSTD(1)),
	INDEX idx_tag_key mapKeys(tags) TYPE bloom_filter(0.01) GRANULARITY 1,
	INDEX idx_tag_value mapValues(tags) TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = MergeTree
PARTITION BY toDate(when)
ORDER BY (service, host, toUnixTimestamp64Nano(when))
TTL toDateTime(when) + toIntervalDay(30)
SETTINGS 
	index_granularity = 8192,
	ttl_only_drop_parts = 1;


-- stores events for 30 days
CREATE TABLE yamon.events (
	`when` DateTime64(9) CODEC(Delta(8), ZSTD(1)),
	`host` LowCardinality(String) CODEC(ZSTD(1)),
	`type` LowCardinality(String) CODEC(ZSTD(1)),
	`data` String CODEC(ZSTD(2)),
	`tags` Map(LowCardinality(String), String) CODEC(ZSTD(1)),
	INDEX idx_tag_key mapKeys(tags) TYPE bloom_filter(0.01) GRANULARITY 1,
	INDEX idx_tag_value mapValues(tags) TYPE bloom_filter(0.01) GRANULARITY 1
)
ENGINE = MergeTree
PARTITION BY toDate(when)
ORDER BY (type, host, toUnixTimestamp64Nano(when))
TTL toDateTime(when) + toIntervalDay(30)
SETTINGS 
	index_granularity = 8192,
	ttl_only_drop_parts = 1;

