package com.github.kxrxh.sinks;

import com.github.kxrxh.model.EventAggregate;
import com.github.kxrxh.repositories.ClickHouseRepository;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.sink.RichSinkFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import org.apache.flink.api.java.utils.ParameterTool;
import com.github.kxrxh.config.ParameterNames;

/**
 * Flink SinkFunction to write EventAggregate objects to ClickHouse events_aggregate table.
 * Reads ClickHouse connection details from global job parameters.
 * Implements batching for improved performance by controlling commit frequency.
 */
public class ClickHouseEventAggregateSink extends RichSinkFunction<EventAggregate> {

    private static final Logger LOG = LoggerFactory.getLogger(ClickHouseEventAggregateSink.class);

    private transient ClickHouseRepository clickHouseRepository;
    private transient ParameterTool params;

    // Batching parameters
    private int batchSize = 1000; // Default batch size
    private long batchIntervalMillis = 5000; // Default batch interval (5 seconds)
    private int currentBatchSize = 0;
    private long lastBatchTime = 0L;

    // Constructor without parameters
    public ClickHouseEventAggregateSink() {
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        // Use the specific GlobalJobParameters type
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
        // Get batch parameters from config, with defaults
        batchSize = params.getInt("clickhouse.sink.batch.size", 1000);
        batchIntervalMillis = params.getLong("clickhouse.sink.batch.interval.ms", 5000L);
        currentBatchSize = 0;
        lastBatchTime = System.currentTimeMillis();

        try {
            // Pass null properties for now
            clickHouseRepository = new ClickHouseRepository(jdbcUrl, null /* connectionProperties */);
            LOG.info("ClickHouse Event Aggregate Sink initialized (URL: {}, Batch Size: {}, Batch Interval: {}ms).", 
                     jdbcUrl, batchSize, batchIntervalMillis);
        } catch (Exception e) {
            LOG.error("Failed to initialize ClickHouseRepository in Sink function", e);
            throw new RuntimeException("Failed to initialize ClickHouse connection in Sink", e);
        }
    }

    @Override
    public void invoke(EventAggregate value, Context context) throws Exception {
        if (clickHouseRepository == null) {
            LOG.error("ClickHouse repository is not initialized. Cannot save event aggregate.");
            throw new IllegalStateException("ClickHouse repository is not initialized.");
        }
        try {
            clickHouseRepository.saveEventAggregate(value);
            currentBatchSize++;

            long currentTime = System.currentTimeMillis();
            if (currentBatchSize >= batchSize || (currentTime - lastBatchTime) >= batchIntervalMillis) {
                commitBatch();
            }
        } catch (Exception e) {
            LOG.error("Failed to save EventAggregate to ClickHouse: {}", value, e);
            rollbackBatch();
            throw new RuntimeException("Failed to save EventAggregate to ClickHouse", e);
        }
    }
    
    private void commitBatch() throws Exception {
         if (currentBatchSize > 0) {
            try {
                clickHouseRepository.commit();
                LOG.info("Committed batch of size {}", currentBatchSize);
            } catch (Exception e) {
                 LOG.error("Error committing batch transaction: {}", e.getMessage(), e);
                 rollbackBatch();
                 throw e;
            } finally {
                currentBatchSize = 0;
                lastBatchTime = System.currentTimeMillis();
            }
        }
    }

    private void rollbackBatch() {
        try {
            clickHouseRepository.rollback();
        } catch (Exception e) {
            LOG.error("Exception occurred during ClickHouse rollback", e);
        }
    }

    @Override
    public void close() throws Exception {
        if (clickHouseRepository != null) {
            try {
                LOG.info("Executing final batch commit (pending size: {}) before closing...", currentBatchSize);
                commitBatch();
            } catch(Exception e) {
                LOG.error("Error committing final batch on close: {}", e.getMessage(), e);
            }
        }
        
        if (clickHouseRepository != null) {
            try {
                clickHouseRepository.close();
                LOG.info("ClickHouse Event Aggregate Sink function closed.");
            } catch (Exception e) {
                LOG.error("Error closing ClickHouseRepository in Sink function", e);
            }
        }
        super.close();
    }
} 