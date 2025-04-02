package com.github.kxrxh.sinks;

import com.github.kxrxh.repositories.MongoRepository;
import com.mongodb.client.MongoCollection;
import com.mongodb.client.model.Filters;
import com.mongodb.client.model.UpdateOptions;
import com.mongodb.client.model.Updates;
import org.apache.flink.api.java.tuple.Tuple2;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.sink.RichSinkFunction;
import org.bson.Document;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

// Import ParameterTool and constants
import org.apache.flink.api.java.utils.ParameterTool;
import com.github.kxrxh.config.ParameterNames;

/**
 * A Flink SinkFunction to write the first seen timestamp for each citizen to MongoDB.
 * Uses upsert with $setOnInsert to ensure only the first timestamp is recorded.
 * Reads MongoDB connection details from global job parameters.
 */
public class MongoDbFirstSeenSinkFunction extends RichSinkFunction<Tuple2<String, Long>> {

    private static final Logger LOG = LoggerFactory.getLogger(MongoDbFirstSeenSinkFunction.class);
    private static final String COLLECTION_NAME = "citizen_first_seen"; // Target collection

    // Remove final fields
    private transient MongoRepository repository;
    private transient MongoCollection<Document> collection;
    private transient ParameterTool params;

    // Constructor without parameters
    public MongoDbFirstSeenSinkFunction() {
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        params = (ParameterTool) getRuntimeContext().getExecutionConfig().getGlobalJobParameters();
         if (params == null) {
             throw new RuntimeException("Global job parameters (ParameterTool) not found.");
         }

        final String mongoUri = params.getRequired(ParameterNames.MONGO_URI);
        final String dbName = params.getRequired(ParameterNames.MONGO_DB_NAME);
        
        try {
            repository = new MongoRepository(mongoUri, dbName);
            collection = repository.getCollection(COLLECTION_NAME);
            LOG.info("MongoDB First Seen Sink function initialized (URI: {}, DB: {}, Collection: {}).", 
                     mongoUri, dbName, COLLECTION_NAME);
        } catch (Exception e) {
            LOG.error("Failed to initialize MongoDB connection or get collection", e);
            throw new RuntimeException("Failed to initialize MongoDB connection for FirstSeenSink", e);
        }
    }

    @Override
    public void invoke(Tuple2<String, Long> value, Context context) throws Exception {
        if (value == null || value.f0 == null || value.f1 == null) {
            LOG.warn("Received null value or components in MongoDbFirstSeenSinkFunction. Skipping.");
            return;
        }
        // Ensure collection is initialized before using it
        if (collection == null) {
             LOG.error("MongoDB collection is null in invoke(). Sink might not have been opened correctly. Skipping.");
             return;
        }

        String citizenId = value.f0;
        long firstSeenTimestamp = value.f1;

        try {
            Document filter = new Document("_id", citizenId);
            Document update = new Document("$setOnInsert", new Document("firstSeenTimestamp", firstSeenTimestamp));
            UpdateOptions options = new UpdateOptions().upsert(true);

            collection.updateOne(filter, update, options);
            LOG.debug("Upserted first seen timestamp for citizen {}: {}", citizenId, firstSeenTimestamp);

        } catch (Exception e) {
            LOG.error("Error upserting first seen timestamp for citizen {} to MongoDB: {}", citizenId, e.getMessage(), e);
        }
    }

    @Override
    public void close() throws Exception {
        // Collection is managed by the repository, no need to close it separately
        try {
            if (repository != null) {
                repository.close();
                LOG.info("MongoDB First Seen Sink function closed.");
            }
        } catch (Exception e) {
            LOG.error("Error closing MongoDB connection for FirstSeenSink", e);
        } finally {
            repository = null;
            collection = null; // Clear reference
        }
        super.close();
    }
} 