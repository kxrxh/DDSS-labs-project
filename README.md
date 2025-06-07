# DDSS-labs-project
Project for distributed data storage systems course


```
                               +-----------------+
                               | Go Event        |<-- (Manual Input/Simulation Logic)
                               | Producer        |
                               +-------+---------+
                                       | (Avro Events)
                                       v
                               +-------+---------+
                               | Redpanda        |
                               | (Kafka Topics)  |
                               +-------+---------+
                                       | (Avro Events)
                                       v
+--------------------------------+-----+-----------------------------------+
| Arroyo (Stream Processing)     |                                         |
| - Consumes from Redpanda       |<----------------------------------------+
| - Reads Rules from MongoDB     |   (Rules, Config)   +-----------------+
| - Calculates Scores            +-------------------->| MongoDB         |
| - Writes to multiple sinks     |                     | - Profiles      |
|                                |   (Score Snapshot)  | - Rules         |
+--------------------------------+                     | - Config        |
        |        |        |        |                   +-----------------+
        | (Raw   | (Score | (Graph | (Data for Batch/
        | Events)| History| Update |  Archive to ScyllaDB)
        |        |        | Event?) |
        v        v        v        v
+-------+--+ +---+----+ +---+----+ +-----------------+
| InfluxDB | | Dgraph | | ScyllaDB        |
| - Raw Ev.| | - Graph|<--| - Event Archive | (Batch Aggregation)
| - Score  | | Updates|   | - Aggregates    |<------>+ (Spark/Flink/Arroyo Batch)
|   History| +--------+   +-------+---------+       ^
| - Metrics|                        |                 |
| (Flux    |<-----------------------+                 | (Data from various sources)
|  Tasks)  |   (Flux Task Input)                     |
+----------+-----------------------------------------+
      |           |            |
      v           v            v
+-----+-----------+------------+-----+
| Reporting/Access Layer             |
| (APIs, Dashboards)                 |
+------------------------------------+
```
