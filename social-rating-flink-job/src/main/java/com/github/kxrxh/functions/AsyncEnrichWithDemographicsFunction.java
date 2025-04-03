package com.github.kxrxh.functions;

import com.github.kxrxh.model.CitizenEvent;
import com.github.kxrxh.model.CitizenProfile;
import com.github.kxrxh.repositories.MongoRepository;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.async.ResultFuture;
import org.apache.flink.streaming.api.functions.async.RichAsyncFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

// Import ParameterTool and constants
import org.apache.flink.api.java.utils.ParameterTool;
import com.github.kxrxh.config.ParameterNames;

import java.util.Collections;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.function.Supplier;

// Remove incorrect imports
// import com.mongodb.client.MongoClient; // Keep this if used elsewhere, but not needed for ParameterTool
// import com.mongodb.client.MongoClients; // Keep this if used elsewhere
// import com.mongodb.client.MongoCollection; // Keep this if used elsewhere
// import com.mongodb.client.MongoDatabase; // Keep this if used elsewhere

// Import GlobalJobParameters
import org.apache.flink.api.common.ExecutionConfig.GlobalJobParameters;

/**
 * A RichAsyncFunction to enrich CitizenEvent with demographic data (ageGroup, gender)
 * fetched asynchronously from MongoDB.
 * It modifies the input CitizenEvent object in place.
 * Reads MongoDB connection details from global job parameters.
 */
public class AsyncEnrichWithDemographicsFunction 
    extends RichAsyncFunction<CitizenEvent, CitizenEvent> {

    private static final Logger LOG = LoggerFactory.getLogger(AsyncEnrichWithDemographicsFunction.class);

    // Remove final fields for mongoUri and dbName
    private transient MongoRepository mongoRepository;
    private transient ExecutorService executorService;
    private transient ParameterTool params; // To store global parameters

    // Constructor without parameters
    public AsyncEnrichWithDemographicsFunction() {
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        
        // Correct way to get ParameterTool using GlobalJobParameters
        GlobalJobParameters globalJobParameters = getRuntimeContext().getExecutionConfig().getGlobalJobParameters();
        
        if (globalJobParameters == null) {
            throw new RuntimeException("Global job parameters not found.");
        }
        
        // Create ParameterTool from the map provided by GlobalJobParameters
        this.params = ParameterTool.fromMap(globalJobParameters.toMap());
        
        if (this.params == null) {
             // This check might be redundant if toMap() never returns null, but good for safety
             throw new RuntimeException("ParameterTool could not be created from job parameters map.");
        }

        // Get MongoDB config from parameters
        final String mongoUri = params.getRequired(ParameterNames.MONGO_URI);
        final String dbName = params.getRequired(ParameterNames.MONGO_DB_NAME);
        // Get thread pool size (optional)
        final int threadPoolSize = params.getInt("async.enrich.demographics.pool.size", 10); // Optional param

        try {
            mongoRepository = new MongoRepository(mongoUri, dbName);
            // Use a dedicated thread pool for MongoDB IO
            executorService = Executors.newFixedThreadPool(threadPoolSize); 
            LOG.info("Opened MongoRepository (URI: {}, DB: {}) and ExecutorService (size: {}) for demographic enrichment.", 
                     mongoUri, dbName, threadPoolSize);
        } catch (Exception e) {
            LOG.error("Failed to open Mongo repository or executor service for demographics", e);
            throw new RuntimeException("Failed to initialize AsyncEnrichWithDemographicsFunction", e);
        }
    }

    @Override
    public void close() throws Exception {
        try {
            if (mongoRepository != null) {
                mongoRepository.close();
                LOG.info("Closed MongoRepository for demographics.");
            }
        } catch (Exception e) {
            LOG.error("Error closing MongoRepository for demographics", e);
        }
        try {
            if (executorService != null) {
                executorService.shutdownNow();
                LOG.info("Shutdown ExecutorService for demographics.");
            }
        } catch (Exception e) {
             LOG.error("Error shutting down ExecutorService for demographics", e);
        }
        super.close();
    }

    @Override
    public void asyncInvoke(CitizenEvent event, ResultFuture<CitizenEvent> resultFuture) throws Exception {
        if (event == null || event.citizenId == null) {
            LOG.warn("Received null event or citizenId. Skipping demographic enrichment.");
             resultFuture.complete(Collections.singleton(event)); // Pass through unmodified
            return;
        }

        final String citizenId = event.citizenId;

        // Run the blocking MongoDB query in the executor service
        CompletableFuture.supplyAsync(new Supplier<CitizenProfile>() {
            @Override
            public CitizenProfile get() {
                try {
                    // Repository instance is available here
                    return mongoRepository.getCitizenProfile(citizenId);
                } catch (Exception e) {
                    LOG.error("Error fetching profile for citizen {}: {}", citizenId, e.getMessage(), e);
                    throw new RuntimeException("Failed to get citizen profile", e);
                }
            }
        }, executorService).thenAcceptAsync((CitizenProfile profile) -> {
            // Success: Enrich the original event object
            if (profile != null) {
                event.ageGroup = profile.ageGroup;
                event.gender = profile.gender;
                LOG.debug("Successfully enriched event for citizen {} with profile data.", citizenId);
            } else {
                 LOG.debug("No profile found for citizen {}, proceeding with original event.", citizenId);
            }
            resultFuture.complete(Collections.singleton(event)); // Emit the (potentially) modified event
            
        }, executorService)
        .exceptionally((Throwable throwable) -> {
            // Failure: Log error and emit the original, unmodified event
            LOG.error("Async demographic enrichment failed for citizen {}: {}", citizenId, throwable.getMessage());
            resultFuture.complete(Collections.singleton(event)); 
            return null; 
        });
    }
    
    @Override
    public void timeout(CitizenEvent input, ResultFuture<CitizenEvent> resultFuture) throws Exception {
        LOG.warn("Async demographic enrichment timed out for citizen {}. Emitting original event.", input != null ? input.citizenId : "null");
        resultFuture.complete(Collections.singleton(input));
    }
} 