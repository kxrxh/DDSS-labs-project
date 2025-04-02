package com.github.kxrxh.sinks;

import com.github.kxrxh.model.ScoreSnapshot;
import com.github.kxrxh.repositories.DgraphRepository;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.sink.RichSinkFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

// Import ParameterTool and constants
import org.apache.flink.api.java.utils.ParameterTool;
import com.github.kxrxh.config.ParameterNames;

/**
 * Flink SinkFunction to write ScoreSnapshot objects to Dgraph.
 * Reads Dgraph connection details from global job parameters.
 */
public class DgraphSinkFunction extends RichSinkFunction<ScoreSnapshot> {

    private static final Logger LOG = LoggerFactory.getLogger(DgraphSinkFunction.class);

    private transient DgraphRepository dgraphRepository;
    // Remove final dgraphUri field
    private transient ParameterTool params;

    // Constructor without parameters
    public DgraphSinkFunction() {
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        params = (ParameterTool) getRuntimeContext().getExecutionConfig().getGlobalJobParameters();
         if (params == null) {
             throw new RuntimeException("Global job parameters (ParameterTool) not found.");
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