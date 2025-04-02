package com.github.kxrxh.sinks;

// import com.github.kxrxh.model.ScoreUpdate; // Old import
import com.github.kxrxh.model.ScoreSnapshot; // New import
import com.github.kxrxh.repositories.MongoRepository;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.sink.RichSinkFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

// Import ParameterTool and constants
import org.apache.flink.api.java.utils.ParameterTool;
import com.github.kxrxh.config.ParameterNames;

/**
 * Flink SinkFunction to write ScoreSnapshot objects to MongoDB 'citizens' collection.
 * Reads MongoDB connection details from global job parameters.
 */
public class MongoDbSnapshotSinkFunction extends RichSinkFunction<ScoreSnapshot> {

    private static final Logger LOG = LoggerFactory.getLogger(MongoDbSnapshotSinkFunction.class);

    private transient MongoRepository mongoRepository;
    // Remove final mongoUri and mongoDbName fields
    private transient ParameterTool params;

    // Constructor without parameters
    public MongoDbSnapshotSinkFunction() {
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        params = (ParameterTool) getRuntimeContext().getExecutionConfig().getGlobalJobParameters();
        if (params == null) {
             throw new RuntimeException("Global job parameters (ParameterTool) not found.");
        }

        final String mongoUri = params.getRequired(ParameterNames.MONGO_URI);
        final String mongoDbName = params.getRequired(ParameterNames.MONGO_DB_NAME);
        
        try {
            mongoRepository = new MongoRepository(mongoUri, mongoDbName);
            LOG.info("MongoDB Snapshot Sink function initialized (URI: {}, DB: {}).", mongoUri, mongoDbName);
        } catch (Exception e) {
            LOG.error("Failed to initialize MongoRepository in Sink function", e);
            throw new RuntimeException("Failed to initialize MongoDB connection in Sink", e);
        }
    }

    @Override
    public void invoke(ScoreSnapshot snapshot, Context context) throws Exception { 
        if (mongoRepository != null) {
            try {
                LOG.debug("Writing ScoreSnapshot to MongoDB for citizen: {}", snapshot.citizenId);
                // Call the repository method to update/insert the snapshot
                // TODO: Enhance updateCitizenSnapshot to accept full snapshot or use a more specific method
                mongoRepository.updateCitizenSnapshot(
                        snapshot.citizenId,
                        snapshot.finalScore, 
                        snapshot.level      
                );
            } catch (Exception e) {
                LOG.error("Failed to write ScoreSnapshot for citizen {} to MongoDB: {}", snapshot.citizenId, e.getMessage(), e);
            }
        } else {
             LOG.warn("MongoRepository not initialized in invoke method. Skipping write for citizen: {}", snapshot.citizenId);
        }
    }

    @Override
    public void close() throws Exception {
        if (mongoRepository != null) {
            try {
                mongoRepository.close();
                LOG.info("MongoDB Snapshot Sink function closed.");
            } catch (Exception e) {
                LOG.error("Error closing MongoRepository in Sink function", e);
            }
        }
        super.close();
    }
} 