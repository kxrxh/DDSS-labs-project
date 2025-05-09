package com.github.kxrxh.sinks;

import com.github.kxrxh.model.CreditScoreAggregate;
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
 * A Flink SinkFunction to write CreditScoreAggregate data to ClickHouse.
 * Reads ClickHouse connection details from global job parameters.
 */
public class ClickHouseCreditScoreSink extends RichSinkFunction<CreditScoreAggregate> {

    private static final Logger LOG = LoggerFactory.getLogger(ClickHouseCreditScoreSink.class);

    private transient ClickHouseRepository repository;
    private transient ParameterTool params;

    // Constructor without parameters
    public ClickHouseCreditScoreSink() {
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        // Correct way to get ParameterTool using GlobalJobParameters
        GlobalJobParameters globalJobParameters = getRuntimeContext().getExecutionConfig().getGlobalJobParameters();
        if (globalJobParameters == null) {
            throw new RuntimeException("Global job parameters not found.");
        }
        // Create ParameterTool from the map
        this.params = ParameterTool.fromMap(globalJobParameters.toMap());
        
        if (params == null) {
             // This check might be redundant if toMap() never returns null, but good for safety
             throw new RuntimeException("ParameterTool could not be created from job parameters map.");
        }

        final String jdbcUrl = params.getRequired(ParameterNames.CLICKHOUSE_JDBC_URL);
        final String user = params.getRequired(ParameterNames.CLICKHOUSE_USER);
        final String password = params.getRequired(ParameterNames.CLICKHOUSE_PASSWORD);

        try {
            // Pass URL, user, and password to the repository
            repository = new ClickHouseRepository(jdbcUrl, user, password);
            LOG.info("ClickHouse Credit Score Sink initialized (URL: {}).", jdbcUrl);
        } catch (Exception e) {
            LOG.error("Failed to initialize ClickHouse connection for CreditScoreSink", e);
            throw new RuntimeException("Failed to initialize ClickHouse connection", e);
        }
    }

    @Override
    public void invoke(CreditScoreAggregate value, Context context) throws Exception {
        if (repository == null) {
            LOG.error("ClickHouse repository is not initialized. Cannot save credit score aggregate.");
            return; // Or throw an exception
        }
        try {
            LOG.debug("Attempting to save credit score aggregate to ClickHouse: {}", value);
            repository.saveCreditScoreAggregate(value);
        } catch (Exception e) {
            LOG.error("Failed to save CreditScoreAggregate to ClickHouse: {}", value, e);
            throw new RuntimeException("Failed to save CreditScoreAggregate to ClickHouse", e);
        }
    }

    @Override
    public void close() throws Exception {
        try {
            if (repository != null) {
                repository.close();
                LOG.info("Closed ClickHouseRepository for Credit Score Sink.");
            }
        } catch (Exception e) {
            LOG.error("Error closing ClickHouseRepository for Credit Score Sink", e);
        } finally {
            repository = null;
        }
        super.close();
    }
} 