# DDSS Labs Project - Distributed Social Credit System

A distributed data storage systems project implementing a simulated social credit system for technical demonstration purposes. This project showcases data integration from diverse sources, real-time scoring updates, distributed data management, and comprehensive reporting capabilities.

## System Architecture

The system employs a polyglot persistence approach using multiple specialized databases to handle diverse data types, query patterns, and scalability requirements across a distributed architecture.

### Core Components

- **Event Generation**: Go applications simulate real-world data streams
- **Message Streaming**: Redpanda provides high-volume, distributed message bus
- **Stream Processing**: Go Streams library handles real-time event processing
- **Polyglot Storage**: Multiple databases optimized for specific use cases
- **Analytics Layer**: InfluxDB for time-series analytics and complex analytical queries

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
+-------+--+ +---+----+ +---+----+
| InfluxDB | | Dgraph | | ScyllaDB|
| - Events | | - Social    | - Archive   |
| - Metrics| |   Graph | - Aggregates|
| - History| | - Relations | - Long-term |
| - Analytics  +--------+ |   Storage   |
| - OLAP   |            | (Columnar)  |
| (Time    |            |             |
|  Series) |            +-------------+
+----------+                     |
      |                          v
      v                          v
+-----+---------------------------+-----+
| Reporting & Analytics Layer                           |
| - Go APIs                                             |
| - Real-time Dashboards                                |
| - Complex Analytics (InfluxDB)                        |
| - Cross-database Queries (via Go APIs)                |
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

### 3. InfluxDB - Time-Series Data & Analytics
- **Raw Events**: Short-term storage of all incoming event data (30-90 days)
- **Derived Metrics**: Score change history, aggregated summaries
- **Real-time Monitoring**: Live metrics and trending analysis
- **Analytical Queries**: Complex OLAP operations using Flux query language
- **Statistical Analysis**: Advanced analytics on citizen behavior patterns
- **Trend Analysis**: Time-windowed aggregations and comparative analysis

### 4. ScyllaDB - Long-term Storage
- **Events Archive**: Long-term event storage (5+ years) with TTL
- **Aggregate Tables**: Pre-calculated analytics optimized for specific queries
- **Operational State**: Rule cooldowns and frequency capping state



## Key Features

- **Real-time Processing**: Go Streams provides lightweight, efficient stream processing
- **Event-driven Architecture**: Redpanda ensures reliable, scalable event distribution
- **Polyglot Persistence**: Each database optimized for specific data patterns
- **Horizontal Scalability**: Distributed design supports high-volume data processing
- **Complex Analytics**: InfluxDB provides powerful time-series analytics and OLAP capabilities
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
5. **Analytics**: InfluxDB performs complex time-series analytics and trend analysis
6. **Reporting**: Go APIs coordinate data from multiple sources and serve unified results

## Technology Stack

- **Stream Processing**: Go Streams library
- **Message Broker**: Redpanda (Kafka-compatible)
- **Document Store**: MongoDB
- **Graph Database**: Dgraph
- **Time-Series**: InfluxDB 2.x/3.x
- **Wide Column**: ScyllaDB
- **Infrastructure**: Terraform, Kubernetes
- **Programming Language**: Go

This architecture demonstrates modern distributed systems principles while showcasing the strengths of different database technologies in a cohesive, real-world applicable system.
