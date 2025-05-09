package com.github.kxrxh.sinks;

import com.github.kxrxh.model.ScoreVelocityMetrics;
import com.github.kxrxh.repositories.ClickHouseRepository;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.sink.RichSinkFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

// Import ParameterTool and constants
import org.apache.flink.api.java.utils.ParameterTool;
import org.apache.flink.api.common.ExecutionConfig.GlobalJobParameters;
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
        
        // Attempt 1: Get parameters directly from the Configuration object passed to open()
        try {
             this.params = ParameterTool.fromMap(parameters.toMap());
             if (params.getNumberOfParameters() == 0) {
                  LOG.warn("ParameterTool created from Configuration.toMap() is empty in ClickHouseVelocitySink. Falling back.");
                  throw new RuntimeException("Empty params from Configuration.toMap()"); // Force fallback
             }
             LOG.info("Successfully obtained parameters using Configuration.toMap() in ClickHouseVelocitySink.");
        } catch (Exception e) {
             LOG.warn("Failed to get parameters using Configuration.toMap() in ClickHouseVelocitySink, trying GlobalJobParameters. Error: {}", e.getMessage());
             // Attempt 2: Fallback to using GlobalJobParameters
             GlobalJobParameters globalJobParameters = getRuntimeContext().getExecutionConfig().getGlobalJobParameters();
             if (globalJobParameters == null) {
                 throw new RuntimeException("Global job parameters not found on fallback in ClickHouseVelocitySink.", e);
             }
             try {
                 this.params = ParameterTool.fromMap(globalJobParameters.toMap());
                 LOG.info("Successfully obtained parameters using GlobalJobParameters.toMap() on fallback in ClickHouseVelocitySink.");
             } catch (Exception e2) {
                  LOG.error("Failed to obtain parameters using both methods in ClickHouseVelocitySink.", e2);
                  throw new RuntimeException("Failed to create ParameterTool using both Configuration.toMap and GlobalJobParameters.toMap in ClickHouseVelocitySink.", e2);
             }
        }

        // Final check if params is usable
        if (params == null || params.getNumberOfParameters() == 0) {
            throw new RuntimeException("ParameterTool could not be created or is empty after trying both methods in ClickHouseVelocitySink.");
        }

        LOG.info("ClickHouseVelocitySink received parameters: {}", params.toMap()); // Log received parameters

        final String jdbcUrl = params.getRequired(ParameterNames.CLICKHOUSE_JDBC_URL);
        final String user = params.getRequired(ParameterNames.CLICKHOUSE_USER);
        final String password = params.getRequired(ParameterNames.CLICKHOUSE_PASSWORD);

        try {
            repository = new ClickHouseRepository(jdbcUrl, user, password);
            repository.createVelocityTableIfNotExists();
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