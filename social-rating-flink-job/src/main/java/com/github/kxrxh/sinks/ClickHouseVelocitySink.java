package com.github.kxrxh.sinks;

import com.github.kxrxh.model.ScoreVelocityMetrics;
import com.github.kxrxh.repositories.ClickHouseRepository;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.sink.RichSinkFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

// Import ParameterTool and constants
import org.apache.flink.api.java.utils.ParameterTool;
import com.github.kxrxh.config.ParameterNames;

/**
 * A Flink SinkFunction to write ScoreVelocityMetrics data to ClickHouse.
 * Reads ClickHouse connection details from global job parameters.
 */
public class ClickHouseVelocitySink extends RichSinkFunction<ScoreVelocityMetrics> {

    private static final Logger LOG = LoggerFactory.getLogger(ClickHouseVelocitySink.class);

    private transient ClickHouseRepository repository;
    private transient ParameterTool params;

    public ClickHouseVelocitySink() {
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        // Corrected ParameterTool retrieval using specific GlobalJobParameters type
        org.apache.flink.api.common.ExecutionConfig.GlobalJobParameters jobParameters = 
            getRuntimeContext().getExecutionConfig().getGlobalJobParameters();
            
        if (jobParameters == null) {
            throw new RuntimeException("Required GlobalJobParameters not found.");
        }
        try {
             java.util.Map<String, String> map = jobParameters.toMap();
             if (map == null) {
                  throw new RuntimeException("GlobalJobParameters.toMap() returned null.");
             }
             params = ParameterTool.fromMap(map);
             if (params == null) {
                 // ParameterTool.fromMap might return null/empty if map is empty, check required params later
                 throw new RuntimeException("ParameterTool could not be created from job parameters map.");
             }
        } catch (UnsupportedOperationException e) {
             LOG.error("Could not convert GlobalJobParameters to map.", e);
             throw new RuntimeException("Could not convert GlobalJobParameters to map.", e);
        }

        final String jdbcUrl = params.getRequired(ParameterNames.CLICKHOUSE_JDBC_URL);
        // Optional: Add user/password retrieval if needed

        try {
            repository = new ClickHouseRepository(jdbcUrl, null);
            LOG.info("ClickHouse Velocity Sink initialized (URL: {}).", jdbcUrl);
        } catch (Exception e) {
            LOG.error("Failed to initialize ClickHouse connection for VelocitySink", e);
            throw new RuntimeException("Failed to initialize ClickHouse connection for VelocitySink", e);
        }
    }

    @Override
    public void invoke(ScoreVelocityMetrics value, Context context) throws Exception {
        if (repository == null) {
            LOG.error("ClickHouse repository is not initialized. Cannot save velocity metrics.");
            return; // Or throw an exception
        }
        try {
            LOG.debug("Attempting to save velocity metrics to ClickHouse: {}", value);
            repository.saveVelocityMetrics(value);
        } catch (Exception e) {
            LOG.error("Failed to save ScoreVelocityMetrics to ClickHouse: {}", value, e);
            throw new RuntimeException("Failed to save ScoreVelocityMetrics to ClickHouse", e);
        }
    }

    @Override
    public void close() throws Exception {
        try {
            if (repository != null) {
                repository.close();
                LOG.info("Closed ClickHouseRepository for VelocitySink.");
            }
        } catch (Exception e) {
            LOG.error("Error closing ClickHouseRepository for VelocitySink", e);
        } finally {
            repository = null;
        }
        super.close();
    }
} 