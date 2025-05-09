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

    // Constructor without parameters
    public InfluxDBSinkFunction() {
    }

     /**
     * Initializes the InfluxDB connection when the sink starts.
     */
    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        // Get global job parameters as ParameterTool, calling .toMap() first
        ParameterTool jobParams = ParameterTool.fromMap(getRuntimeContext().getExecutionConfig().getGlobalJobParameters().toMap());

        if (jobParams == null) {
             throw new RuntimeException("Global job parameters (ParameterTool) not found.");
         }

        // Get InfluxDB config using ParameterTool's getRequired method
        // This automatically throws an exception if the parameter is missing
        try {
            final String influxUrl = jobParams.getRequired(ParameterNames.INFLUXDB_URL);
            final String influxToken = jobParams.getRequired(ParameterNames.INFLUXDB_TOKEN);
            final String influxOrg = jobParams.getRequired(ParameterNames.INFLUXDB_ORG);
            final String influxBucket = jobParams.getRequired(ParameterNames.INFLUXDB_BUCKET);

            // Validation (trim checks can still be useful if empty strings are possible but invalid)
            if (influxUrl.trim().isEmpty()) {
                 throw new IllegalArgumentException("InfluxDB URL (" + ParameterNames.INFLUXDB_URL + ") cannot be empty.");
            }
            if (influxToken.trim().isEmpty()) {
                LOG.error("InfluxDB token parameter '{}' cannot be empty.", ParameterNames.INFLUXDB_TOKEN);
                throw new IllegalArgumentException("InfluxDB token (" + ParameterNames.INFLUXDB_TOKEN + ") cannot be empty.");
            }
            if (influxOrg.trim().isEmpty()) {
                throw new IllegalArgumentException("InfluxDB Org (" + ParameterNames.INFLUXDB_ORG + ") cannot be empty.");
            }
            if (influxBucket.trim().isEmpty()) {
                throw new IllegalArgumentException("InfluxDB Bucket (" + ParameterNames.INFLUXDB_BUCKET + ") cannot be empty.");
            }

            // Pass configuration to the repository constructor
            influxDBRepository = new InfluxDBRepository(influxUrl, influxToken, influxOrg, influxBucket);
            LOG.info("InfluxDB Sink function initialized (URL: {}, Org: {}, Bucket: {}).", influxUrl, influxOrg, influxBucket);
        } catch (IllegalStateException e) {
            // ParameterTool.getRequired throws IllegalStateException if parameter is missing
            LOG.error("Missing required InfluxDB configuration parameter: {}", e.getMessage(), e);
            throw new RuntimeException("Missing required InfluxDB configuration parameter.", e);
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