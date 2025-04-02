package com.github.kxrxh.sinks; // Assuming a 'sinks' subpackage

// import com.github.kxrxh.model.ScoreUpdate; // Old import
import com.github.kxrxh.model.ScoreSnapshot; // New import
import com.github.kxrxh.repositories.InfluxDBRepository;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.sink.RichSinkFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

// Import ParameterTool and constants
import org.apache.flink.api.java.utils.ParameterTool;
import com.github.kxrxh.config.ParameterNames;

/**
 * Flink SinkFunction to write score changes from ScoreSnapshot objects 
 * to InfluxDB's score_history measurement.
 * Reads InfluxDB connection details from global job parameters.
 */
public class InfluxDBSinkFunction extends RichSinkFunction<ScoreSnapshot> {

    private static final Logger LOG = LoggerFactory.getLogger(InfluxDBSinkFunction.class);

    private transient InfluxDBRepository influxDBRepository;
    private transient ParameterTool params;

    // Constructor without parameters
    public InfluxDBSinkFunction() {
    }

     /**
     * Initializes the InfluxDB connection when the sink starts.
     */
    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        params = (ParameterTool) getRuntimeContext().getExecutionConfig().getGlobalJobParameters();
        if (params == null) {
             throw new RuntimeException("Global job parameters (ParameterTool) not found.");
         }

        // Get InfluxDB config from parameters
        final String influxUrl = params.getRequired(ParameterNames.INFLUXDB_URL);
        final String influxToken = params.getRequired(ParameterNames.INFLUXDB_TOKEN);
        final String influxOrg = params.getRequired(ParameterNames.INFLUXDB_ORG);
        final String influxBucket = params.getRequired(ParameterNames.INFLUXDB_BUCKET);

         // Basic validation
        if (influxToken == null || influxToken.trim().isEmpty()) {
            LOG.error("InfluxDB token parameter '{}' cannot be null or empty.", ParameterNames.INFLUXDB_TOKEN);
            throw new IllegalArgumentException("InfluxDB token is required.");
        }

        try {
            // Pass configuration to the repository constructor
            influxDBRepository = new InfluxDBRepository(influxUrl, influxToken, influxOrg, influxBucket);
            LOG.info("InfluxDB Sink function initialized (URL: {}, Org: {}, Bucket: {}).", influxUrl, influxOrg, influxBucket);
        } catch (Exception e) {
            LOG.error("Failed to initialize InfluxDBRepository in Sink function", e);
            throw new RuntimeException("Failed to initialize InfluxDB connection in Sink", e);
        }
    }

    /**
     * Called for each ScoreSnapshot record.
     * Writes the score *change* and event timestamp to InfluxDB.
     */
    @Override
    public void invoke(ScoreSnapshot snapshot, Context context) throws Exception {
        if (influxDBRepository != null) {
            try {
                LOG.debug("Writing score change from snapshot to InfluxDB for citizen: {}", snapshot.citizenId);
                influxDBRepository.saveScoreHistoryPoint(
                        snapshot.citizenId,
                        snapshot.eventTimestamp, 
                        snapshot.scoreChange    
                );
            } catch (Exception e) {
                LOG.error("Failed to write score change for citizen {} to InfluxDB: {}", snapshot.citizenId, e.getMessage(), e);
            }
        } else {
            LOG.warn("InfluxDBRepository not initialized in invoke method. Skipping write for citizen: {}", snapshot.citizenId);
        }
    }

    /**
     * Closes the InfluxDB connection when the sink stops.
     */
    @Override
    public void close() throws Exception {
        if (influxDBRepository != null) {
            try {
                influxDBRepository.close();
                LOG.info("InfluxDB Sink function closed.");
            } catch (Exception e) {
                LOG.error("Error closing InfluxDBRepository in Sink function", e);
            }
        }
        super.close();
    }
} 