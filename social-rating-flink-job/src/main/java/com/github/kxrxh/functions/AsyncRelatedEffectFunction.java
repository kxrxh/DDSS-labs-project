package com.github.kxrxh.functions;

import com.github.kxrxh.model.RelatedEffectTrigger;
import com.github.kxrxh.model.ScoreUpdate; // Emits secondary ScoreUpdate
import com.github.kxrxh.repositories.DgraphRepository;
import com.github.kxrxh.model.RelatedEffect; // Added import
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.async.ResultFuture;
import org.apache.flink.streaming.api.functions.async.RichAsyncFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

// Import ParameterTool and constants
import org.apache.flink.api.java.utils.ParameterTool;
import com.github.kxrxh.config.ParameterNames;

import java.util.Collections;
import java.util.List;
import java.util.Set;
import java.util.Map; // Import Map
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.function.Supplier;
import java.util.ArrayList; // Use ArrayList
import java.util.Collection;
import java.util.concurrent.TimeUnit; // Add import for TimeUnit

/**
 * Async function to handle RelatedEffectTriggers.
 * Finds related citizens in Dgraph and generates ScoreUpdate events for them.
 */
public class AsyncRelatedEffectFunction extends RichAsyncFunction<RelatedEffectTrigger, ScoreUpdate> {

    private static final Logger LOG = LoggerFactory.getLogger(AsyncRelatedEffectFunction.class);

    private transient DgraphRepository dgraphRepository;
    private transient ExecutorService executorService; // For running async tasks

    private final String dgraphUri;

    // Constructor accepting Dgraph URI
    public AsyncRelatedEffectFunction(String dgraphUri) {
        this.dgraphUri = dgraphUri;
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        try {
            dgraphRepository = new DgraphRepository(dgraphUri);
            // Consider making the thread pool size configurable
            executorService = Executors.newFixedThreadPool(10); 
            LOG.info("DgraphRepository and ExecutorService initialized successfully.");
        } catch (Exception e) {
            LOG.error("Failed to initialize DgraphRepository or ExecutorService", e);
            throw new RuntimeException("Failed to initialize async function resources", e);
        }
    }

    @Override
    public void close() throws Exception {
        if (dgraphRepository != null) {
            try {
                dgraphRepository.close();
                LOG.info("DgraphRepository closed successfully.");
            } catch (Exception e) {
                LOG.error("Error closing DgraphRepository", e);
            }
        }
        if (executorService != null) {
            executorService.shutdown();
             try {
                 if (!executorService.awaitTermination(5, TimeUnit.SECONDS)) {
                     executorService.shutdownNow();
                 }
                 LOG.info("ExecutorService shut down successfully.");
             } catch (InterruptedException e) {
                 LOG.warn("ExecutorService shutdown interrupted.", e);
                 executorService.shutdownNow();
                 Thread.currentThread().interrupt();
             }
        }
        super.close();
    }

    @Override
    public void asyncInvoke(RelatedEffectTrigger trigger, ResultFuture<ScoreUpdate> resultFuture) throws Exception {

        CompletableFuture.supplyAsync(() -> {
            try {
                LOG.debug("Processing related effect trigger: {}", trigger);
                
                // Check if effectDetails is present
                if (trigger.effectDetails == null) {
                    LOG.warn("Trigger missing effectDetails: {}", trigger);
                    return Collections.<ScoreUpdate>emptyList();
                }
                
                RelatedEffect effect = trigger.effectDetails; // Extract effect details

                // Call the updated repository method using details from effect
                Map<String, String> relatedCitizenInfo = dgraphRepository.findRelatedCitizens(
                    trigger.originalCitizenId, 
                    effect.relationType, // Use effectDetails
                    effect.maxDistance != null ? effect.maxDistance : 1 // Use effectDetails, provide default
                );

                if (relatedCitizenInfo.isEmpty()) {
                    LOG.debug("No related citizens found for trigger: {}", trigger);
                    return Collections.emptyList(); // Return empty list if no related found
                }

                List<ScoreUpdate> relatedUpdates = new ArrayList<>(relatedCitizenInfo.size());
                
                // Iterate over the map (UID -> citizenId)
                for (Map.Entry<String, String> entry : relatedCitizenInfo.entrySet()) {
                    String relatedUid = entry.getKey(); 
                    String relatedCitizenId = entry.getValue(); 

                    // Create ScoreUpdate for the *related* citizen using their citizenId
                    ScoreUpdate update = new ScoreUpdate(
                        relatedCitizenId, 
                        effect.percentage != null ? effect.percentage : 0.0, 
                        System.currentTimeMillis(), // Correct position for eventTimestamp (long)
                        trigger.triggeringRuleId,   // Correct position for triggeringRuleId (String)
                        trigger.originalSourceEventId + "-related" 
                    );
                    relatedUpdates.add(update); 
                }

                return relatedUpdates; 
            } catch (Exception e) {
                LOG.error("Error processing related effect for trigger {}: {}", trigger, e.getMessage(), e);
                return Collections.<ScoreUpdate>emptyList(); 
            }
        }, executorService).whenCompleteAsync((updates, ex) -> {
            if (ex != null) {
                 LOG.error("Async processing failed for related effect trigger: {}", trigger, ex);
                 resultFuture.completeExceptionally(new RuntimeException("Failed processing related effect", ex));
            } else {
                 resultFuture.complete((Collection<ScoreUpdate>) updates);
                 if (!updates.isEmpty()) {
                    LOG.debug("Successfully processed related effect trigger, emitting {} secondary updates.", updates.size());
                } else {
                     LOG.debug("Completed related effect processing for trigger {} (no updates emitted).", trigger);
                }
            }
        }, executorService);
    }

    @Override
    public void timeout(RelatedEffectTrigger input, ResultFuture<ScoreUpdate> resultFuture) throws Exception {
         LOG.warn("Async Dgraph related effect query timed out for trigger: {}", input);
         resultFuture.complete(Collections.emptyList()); 
    }
} 