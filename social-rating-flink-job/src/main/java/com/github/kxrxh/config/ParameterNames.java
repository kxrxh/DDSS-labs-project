package com.github.kxrxh.config;

/**
 * Defines constants for configuration parameter names used in the Flink job.
 * These names correspond to the keys expected in the command-line arguments
 * or properties file passed via ParameterTool.
 */
public final class ParameterNames {

    private ParameterNames() { // Prevent instantiation
    }

    // Kafka Source Parameters
    public static final String KAFKA_BROKERS = "kafka.brokers";
    public static final String KAFKA_INPUT_TOPIC = "kafka.input.topic";
    public static final String KAFKA_CONSUMER_GROUP = "kafka.consumer.group";

    // Kafka Sink Parameters
    public static final String KAFKA_SECONDARY_UPDATE_TOPIC = "kafka.secondary.update.topic";

    // MongoDB Parameters
    public static final String MONGO_URI = "mongo.uri";
    public static final String MONGO_DB_NAME = "mongo.db.name";
    public static final String MONGO_POLLING_INTERVAL_SECONDS = "mongo.polling.interval.seconds";

    // Dgraph Parameters
    public static final String DGRAPH_URI = "dgraph.uri";

    // InfluxDB Parameters
    public static final String INFLUXDB_URL = "influxdb.url";
    public static final String INFLUXDB_TOKEN = "influxdb.token";
    public static final String INFLUXDB_ORG = "influxdb.org";
    public static final String INFLUXDB_BUCKET = "influxdb.bucket";

    // ClickHouse Parameters
    public static final String CLICKHOUSE_JDBC_URL = "clickhouse.jdbc.url";
    // Add user/password parameters if needed for ClickHouse:
    public static final String CLICKHOUSE_USER = "clickhouse.user";
    public static final String CLICKHOUSE_PASSWORD = "clickhouse.password";

    // Async Function Parameters (Timeouts in ms, Capacity)
    public static final String ASYNC_DEMO_TIMEOUT_MS = "async.enrich.demographics.timeout.ms";
    public static final String ASYNC_DEMO_CAPACITY = "async.enrich.demographics.capacity";
    public static final String ASYNC_REL_TIMEOUT_MS = "async.enrich.relation.timeout.ms";
    public static final String ASYNC_REL_CAPACITY = "async.enrich.relation.capacity";
    public static final String ASYNC_RELATED_EFFECT_TIMEOUT_MS = "async.related.effect.timeout.ms";
    public static final String ASYNC_RELATED_EFFECT_CAPACITY = "async.related.effect.capacity";

} 