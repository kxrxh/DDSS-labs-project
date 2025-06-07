# DDSS Labs Project - Distributed Social Credit System

A distributed data storage systems project implementing a simulated social credit system for technical demonstration purposes. This project showcases data integration from diverse sources, real-time scoring updates, distributed data management, and comprehensive reporting capabilities.

## System Architecture

The system employs a polyglot persistence approach using multiple specialized databases to handle diverse data types, query patterns, and scalability requirements across a distributed architecture.

### Core Components

- **Event Generation**: Go applications simulate real-world data streams
- **Message Streaming**: Redpanda provides high-volume, distributed message bus
- **Stream Processing**: Go Streams library handles real-time event processing
- **Polyglot Storage**: Multiple databases optimized for specific use cases
- **Analytics Layer**: DuckDB for complex analytical queries and data science workloads

## Architecture Diagram

```
                               +-----------------+
                               | Go Event        |<-- (Simulated Data Sources)
                               | Producer        |
                               +-------+---------+
                                       | (JSON Events)
                                       v
                               +-------+---------+
                               | Redpanda        |
                               | (Event Topics)  |
                               +-------+---------+
                                       | (JSON Event Streams)
                                       v
+--------------------------------+-----+-----------------------------------+
| Go Streams Processing          |                                         |
| - Consumes from Redpanda       |<----------------------------------------+
| - Reads Rules from MongoDB     |   (Rules, Config)   +-----------------+
| - Calculates Scores            +-------------------->| MongoDB         |
| - Real-time Updates            |                     | - Citizen Data  |
| - Multi-sink Output            |   (Score Updates)   | - Rules Engine  |
+--------------------------------+                     | - Configuration |
        |        |        |        |                   +-----------------+
        | (Raw   | (Score | (Graph |
        | Events)| History| Updates|
        |        |        |        |
        v        v        v        v
+-------+--+ +---+----+ +---+----+ +-----------------+
| InfluxDB | | Dgraph | | ScyllaDB| | DuckDB          |
| - Events | | - Social    | - Archive   | - Analytics     |
| - Metrics| |   Graph | - Aggregates| - Cross-DB      |
| - History| | - Relations | - Long-term | - Data Science  |
| (Time    | +--------+ |   Storage   | - OLAP Queries  |
|  Series) |            | (Columnar)  | (In-Memory)     |
+----------+            +---------+---+ +---------+-----+
      |                           |               |
      v                           v               v
+-----+---------------------------+---------------+-----+
| Reporting & Analytics Layer                           |
| - Go APIs                                             |
| - Real-time Dashboards                                |
| - Complex Analytics (DuckDB)                          |
| - Cross-database Queries                              |
+-------------------------------------------------------+
```

## Database Roles & Responsibilities

### 1. MongoDB - Core System Data
- **Citizens Collection**: Primary citizen profiles, current scores, demographics
- **Scoring Rules**: Detailed rule definitions with conditions and point values
- **System Configuration**: Global settings, tier definitions, decay policies

### 2. Dgraph - Social Graph
- **CitizenNode**: Relationship-focused storage for social connections
- **Graph Analytics**: Complex relationship queries and social network analysis

### 3. InfluxDB - Time-Series Data
- **Raw Events**: Short-term storage of all incoming event data (30-90 days)
- **Derived Metrics**: Score change history, aggregated summaries
- **Real-time Monitoring**: Live metrics and trending analysis

### 4. ScyllaDB - Long-term Storage
- **Events Archive**: Long-term event storage (5+ years) with TTL
- **Aggregate Tables**: Pre-calculated analytics optimized for specific queries
- **Operational State**: Rule cooldowns and frequency capping state

### 5. DuckDB - Analytics Engine
- **OLAP Capabilities**: Complex analytical queries and aggregations
- **Cross-Database Analytics**: Federated queries across multiple sources
- **Data Science Workloads**: Advanced analytics and trend analysis
- **Reporting Engine**: Powers BI dashboards and analytical reports

## Key Features

- **Real-time Processing**: Go Streams provides lightweight, efficient stream processing
- **Event-driven Architecture**: Redpanda ensures reliable, scalable event distribution
- **Polyglot Persistence**: Each database optimized for specific data patterns
- **Horizontal Scalability**: Distributed design supports high-volume data processing
- **Complex Analytics**: DuckDB enables sophisticated analytical capabilities
- **Social Graph Analysis**: Dgraph optimized for relationship queries

## Workflow Example

1. **Event Generation**: Go producer simulates a "late payment" event
2. **Stream Ingestion**: Event published to Redpanda topic
3. **Real-time Processing**: Go Streams application:
   - Fetches scoring rules from MongoDB
   - Applies business logic to event data
   - Updates citizen score in MongoDB
   - Archives raw event to InfluxDB
   - Records score change history
   - Sends to ScyllaDB for long-term storage
4. **Batch Processing**: Periodic aggregation jobs create summaries
5. **Analytics**: DuckDB performs complex queries across all data sources
6. **Reporting**: APIs serve data from appropriate databases based on query requirements

## Technology Stack

- **Stream Processing**: Go Streams library
- **Message Broker**: Redpanda (Kafka-compatible)
- **Document Store**: MongoDB
- **Graph Database**: Dgraph
- **Time-Series**: InfluxDB 2.x/3.x
- **Wide Column**: ScyllaDB
- **Analytics**: DuckDB
- **Infrastructure**: Terraform, Kubernetes
- **Programming Language**: Go

This architecture demonstrates modern distributed systems principles while showcasing the strengths of different database technologies in a cohesive, real-world applicable system.
