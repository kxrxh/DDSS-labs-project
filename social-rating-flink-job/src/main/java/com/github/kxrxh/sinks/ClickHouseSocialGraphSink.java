package com.github.kxrxh.sinks;

import com.github.kxrxh.model.SocialGraphAggregate;
import com.github.kxrxh.repositories.ClickHouseRepository;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.sink.RichSinkFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

// Import ParameterTool and constants
import org.apache.flink.api.java.utils.ParameterTool;
import com.github.kxrxh.config.ParameterNames;

/**
 * A Flink SinkFunction to write SocialGraphAggregate data to ClickHouse.
 * Reads ClickHouse connection details from global job parameters.
 */
public class ClickHouseSocialGraphSink extends RichSinkFunction<SocialGraphAggregate> {

    private static final Logger LOG = LoggerFactory.getLogger(ClickHouseSocialGraphSink.class);

    private transient ClickHouseRepository repository;
    private transient ParameterTool params;

    // Constructor without parameters
    public ClickHouseSocialGraphSink() {
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        params = (ParameterTool) getRuntimeContext().getExecutionConfig().getGlobalJobParameters();
        if (params == null) {
             throw new RuntimeException("Global job parameters (ParameterTool) not found.");
         }

        final String jdbcUrl = params.getRequired(ParameterNames.CLICKHOUSE_JDBC_URL);
        // Optional: Add user/password retrieval if needed

        try {
            repository = new ClickHouseRepository(jdbcUrl, null);
            LOG.info("ClickHouse Social Graph Sink initialized (URL: {}).", jdbcUrl);
        } catch (Exception e) {
            LOG.error("Failed to initialize ClickHouse connection for SocialGraphSink", e);
            throw new RuntimeException("Failed to initialize ClickHouse connection for SocialGraphSink", e);
        }
    }

    @Override
    public void invoke(SocialGraphAggregate value, Context context) throws Exception {
        if (repository == null) {
            LOG.error("ClickHouse repository is not initialized. Cannot save social graph aggregate.");
            return; // Or throw an exception
        }
        try {
            LOG.debug("Attempting to save social graph aggregate to ClickHouse: {}", value);
            repository.saveSocialGraphAggregate(value);
        } catch (Exception e) {
            LOG.error("Failed to save SocialGraphAggregate to ClickHouse: {}", value, e);
            throw new RuntimeException("Failed to save SocialGraphAggregate to ClickHouse", e);
        }
    }

    @Override
    public void close() throws Exception {
        try {
            if (repository != null) {
                repository.close();
                LOG.info("Closed ClickHouseRepository for SocialGraphSink.");
            }
        } catch (Exception e) {
            LOG.error("Error closing ClickHouseRepository for SocialGraphSink", e);
        } finally {
            repository = null;
        }
        super.close();
    }
} 