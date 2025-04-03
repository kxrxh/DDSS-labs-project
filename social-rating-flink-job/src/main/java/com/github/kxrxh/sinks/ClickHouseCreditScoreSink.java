package com.github.kxrxh.sinks;

import com.github.kxrxh.model.CreditScoreAggregate;
import com.github.kxrxh.repositories.ClickHouseRepository;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.sink.RichSinkFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

// Import ParameterTool and constants
import org.apache.flink.api.java.utils.ParameterTool;
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
        // Optional: Add user/password retrieval from params if needed

        try {
            // Pass null properties for now
            repository = new ClickHouseRepository(jdbcUrl, null);
            LOG.info("ClickHouse Credit Score Sink initialized (URL: {}).", jdbcUrl);
        } catch (Exception e) {
            LOG.error("Failed to initialize ClickHouse connection", e);
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