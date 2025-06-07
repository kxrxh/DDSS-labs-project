-- ScyllaDB CQL Schema
-- This schema is designed for query patterns and replaces the previous ClickHouse DDL.

CREATE KEYSPACE IF NOT EXISTS social_credit_system
  WITH replication = {'class': 'NetworkTopologyStrategy', 'replication_factor': 3}; -- Adjust RF as needed

USE social_credit_system;

-- ========================================================================
-- Table 1: events_archive
-- Purpose: Long-term storage of citizen events for audit, replay, or batch analytics.
-- Design: Optimized for retrieving events for a citizen over time.
-- Data Source: Written by Arroyo (or an intermediate ETL) from the event stream.
-- ========================================================================
CREATE TABLE IF NOT EXISTS events_archive (
    citizen_id text,
    event_time timestamp,
    event_id text,
    event_type text,
    event_subtype text,
    source_system text,
    region text,
    severity double,
    confidence double,
    payload text,
    PRIMARY KEY ((citizen_id), event_time, event_id)
) WITH CLUSTERING ORDER BY (event_time DESC, event_id ASC)
  AND default_time_to_live = 157680000; -- 5 years in seconds

-- Rule Application State Table
CREATE TABLE IF NOT EXISTS rule_application_state (
    citizen_id text,
    rule_id text,
    last_applied timestamp,
    cooldown_until timestamp,
    application_count counter,
    PRIMARY KEY ((citizen_id), rule_id)
);

-- Credit Score Aggregates by Region and Date
CREATE TABLE IF NOT EXISTS credit_score_aggregates_by_region_date (
    region text,
    date date,
    total_citizens counter,
    average_score double,
    score_variance double,
    tier_distribution map<text, counter>,
    PRIMARY KEY ((region), date)
);

-- Events Summary by Type, Location, and Date
CREATE TABLE IF NOT EXISTS events_summary_by_type_location_date (
    event_type text,
    region text,
    date date,
    total_events counter,
    unique_citizens counter,
    average_severity double,
    PRIMARY KEY ((event_type, region), date)
);

-- Citizen Score History
CREATE TABLE IF NOT EXISTS citizen_score_history (
    citizen_id text,
    date date,
    score double,
    tier text,
    change_reason text,
    rule_id text,
    PRIMARY KEY ((citizen_id), date)
) WITH CLUSTERING ORDER BY (date DESC);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_events_archive_event_type ON events_archive (event_type);
CREATE INDEX IF NOT EXISTS idx_events_archive_source_system ON events_archive (source_system);
CREATE INDEX IF NOT EXISTS idx_events_archive_region ON events_archive (region);

-- Optional: If you frequently need to look up a specific event by its event_id globally:
CREATE TABLE IF NOT EXISTS event_lookup_by_id (
    event_id uuid PRIMARY KEY,
    citizen_id text,
    event_year_month int, // To reconstruct the PK for events_archive
    event_time timestamp
);
// This table would require dual writes or a batch process to populate.

-- ========================================================================
-- Table 2: credit_score_aggregates_by_region_date
-- Purpose: Stores daily aggregated credit score metrics, primarily by region and date.
-- Design: Optimized for querying aggregated scores for a region on a specific day, then by demographics.
-- Data Source: Written by a batch aggregation job (e.g., Spark, or Arroyo if it handles such batching).
-- ========================================================================
CREATE TABLE IF NOT EXISTS credit_score_aggregates_by_region_date (
    region text,
    aggregation_date date,        // YYYY-MM-DD

    -- Demographic dimensions (can be clustering keys or regular columns depending on query patterns)
    age_group text,               // e.g., "18-25", "26-35"
    gender text,
    occupation_category text,     // Optional, if too granular, might lead to wide rows or too many partitions
    income_bracket text,          // Optional

    -- Aggregated metrics
    avg_score float,
    median_score float,
    min_score int,
    max_score int,
    total_citizens_in_aggregate int,

    -- Tier counts (ensure tier names match MongoDB system_configuration.tierDefinitions.tierName)
    tier_exemplary_count int,
    tier_standard_count int,
    tier_caution_count int,
    tier_highrisk_count int,
    // Add other tiers as defined

    score_improvement_count int,  // Number of citizens whose score improved to reach this aggregate state
    score_deterioration_count int,

    last_updated timestamp,       // When this aggregate was last calculated

    PRIMARY KEY ((region, aggregation_date), age_group, gender) // Example PK
    // Alternative PKs depending on query needs:
    // PRIMARY KEY (((region, age_group), aggregation_date), gender) // If region+age_group is a primary filter
    // If city is important and not too high cardinality with region:
    // PRIMARY KEY (((region, city), aggregation_date), age_group, gender)
)
WITH CLUSTERING ORDER BY (age_group ASC, gender ASC);

-- ========================================================================
-- Table 3: events_summary_by_type_location_date
-- Purpose: Stores daily aggregated event counts and impacts by type, location, and date.
-- Design: Optimized for querying event trends.
-- Data Source: Written by a batch aggregation job.
-- ========================================================================
CREATE TABLE IF NOT EXISTS events_summary_by_type_location_date (
    event_type text,
    region text,
    aggregation_date date,

    event_subtype text,           // Optional, can be part of PK or a field
    age_group text,               // Optional demographic dimension for event aggregation
    gender text,                  // Optional

    event_count int,
    total_score_impact int,
    avg_score_impact float,
    distinct_citizens_affected int,

    last_updated timestamp,

    PRIMARY KEY ((event_type, region, aggregation_date), event_subtype, age_group, gender)
)
WITH CLUSTERING ORDER BY (event_subtype ASC, age_group ASC, gender ASC);

-- Example Queries:

-- 1. Get recent events for a citizen
-- SELECT * FROM events_archive 
-- WHERE citizen_id = '12345' 
-- AND event_time > '2024-01-01' 
-- LIMIT 100;

-- 2. Get rule application state
-- SELECT * FROM rule_application_state 
-- WHERE citizen_id = '12345' 
-- AND rule_id = 'rule_001';

-- 3. Get regional score aggregates
-- SELECT * FROM credit_score_aggregates_by_region_date 
-- WHERE region = 'north' 
-- AND date > '2024-01-01';

-- 4. Get event summaries
-- SELECT * FROM events_summary_by_type_location_date 
-- WHERE event_type = 'payment' 
-- AND region = 'north' 
-- AND date > '2024-01-01';

-- 5. Get citizen score history
-- SELECT * FROM citizen_score_history 
-- WHERE citizen_id = '12345' 
-- AND date > '2024-01-01';