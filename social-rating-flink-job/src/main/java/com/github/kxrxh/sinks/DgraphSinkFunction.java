package com.github.kxrxh.sinks;

import com.github.kxrxh.model.ScoreSnapshot;
import com.github.kxrxh.repositories.DgraphRepository;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.sink.RichSinkFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

// Import ParameterTool and constants
import org.apache.flink.api.java.utils.ParameterTool;
import org.apache.flink.api.common.ExecutionConfig.GlobalJobParameters;
import com.github.kxrxh.config.ParameterNames;

/**
 * Flink SinkFunction to write ScoreSnapshot objects to Dgraph.
 * Reads Dgraph connection details from global job parameters.
 */
public class DgraphSinkFunction extends RichSinkFunction<ScoreSnapshot> {

    private static final Logger LOG = LoggerFactory.getLogger(DgraphSinkFunction.class);

    private transient DgraphRepository dgraphRepository;
    private transient ParameterTool params;

    // Constructor without parameters
    public DgraphSinkFunction() {
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        
        // Attempt 1: Get parameters directly from the Configuration object passed to open()
        try {
             this.params = ParameterTool.fromMap(parameters.toMap());
             if (params.getNumberOfParameters() == 0) {
                  LOG.warn("ParameterTool created from Configuration.toMap() is empty. Falling back.");
                  throw new RuntimeException("Empty params from Configuration.toMap()"); // Force fallback
             }
             LOG.info("Successfully obtained parameters using Configuration.toMap().");
        } catch (Exception e) {
             LOG.warn("Failed to get parameters using Configuration.toMap(), trying GlobalJobParameters. Error: {}", e.getMessage());
             // Attempt 2: Fallback to using GlobalJobParameters (the previous method)
             GlobalJobParameters globalJobParameters = getRuntimeContext().getExecutionConfig().getGlobalJobParameters();
             if (globalJobParameters == null) {
                 throw new RuntimeException("Global job parameters not found on fallback.", e);
             }
             try {
                 this.params = ParameterTool.fromMap(globalJobParameters.toMap());
                 LOG.info("Successfully obtained parameters using GlobalJobParameters.toMap() on fallback.");
             } catch (Exception e2) {
                  LOG.error("Failed to obtain parameters using both methods.", e2);
                  // Re-throw the original exception if the fallback also fails badly, 
                  // or the second exception if the first worked but was empty.
                  throw new RuntimeException("Failed to create ParameterTool using both Configuration.toMap and GlobalJobParameters.toMap.", e2);
             }
        }

        // Final check if params is usable
        if (params == null || params.getNumberOfParameters() == 0) {
            throw new RuntimeException("ParameterTool could not be created or is empty after trying both methods.");
        }

        final String dgraphUri = params.getRequired(ParameterNames.DGRAPH_URI);

        try {
            dgraphRepository = new DgraphRepository(dgraphUri);
            LOG.info("Dgraph Sink function initialized (URI: {}).", dgraphUri);
        } catch (Exception e) {
            LOG.error("Failed to initialize DgraphRepository in Sink function", e);
            throw new RuntimeException("Failed to initialize Dgraph connection in Sink", e);
        }
    }

    @Override
    public void invoke(ScoreSnapshot snapshot, Context context) throws Exception {
        if (dgraphRepository != null) {
            try {
                 dgraphRepository.saveScoreChange(snapshot);
            } catch (Exception e) {
                // Log the error but don't fail the sink for one record
                LOG.error("Failed to write ScoreSnapshot for citizen {} to Dgraph: {}", snapshot.citizenId, e.getMessage(), e);
                // Consider retry or dead-letter queue strategy here
            }
        } else {
             LOG.warn("DgraphRepository not initialized in invoke method. Skipping write for citizen: {}", snapshot.citizenId);
        }
    }

    @Override
    public void close() throws Exception {
        if (dgraphRepository != null) {
            try {
                dgraphRepository.close();
                LOG.info("Dgraph Sink function closed.");
            } catch (Exception e) {
                LOG.error("Error closing DgraphRepository in Sink function", e);
            }
        }
        super.close();
    }
} 