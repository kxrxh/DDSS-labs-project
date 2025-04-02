package com.github.kxrxh.functions;

import com.github.kxrxh.model.ScoreSnapshot;
import com.github.kxrxh.model.ScoreSnapshotWithRelationCount;
import com.github.kxrxh.repositories.DgraphRepository;
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

/**
 * A RichAsyncFunction to enrich ScoreSnapshot with the citizen's relationship count
 * fetched asynchronously from Dgraph.
 * Reads Dgraph connection details from global job parameters.
 */
public class AsyncEnrichWithRelationCountFunction 
    extends RichAsyncFunction<ScoreSnapshot, ScoreSnapshotWithRelationCount> {

    private static final Logger LOG = LoggerFactory.getLogger(AsyncEnrichWithRelationCountFunction.class);

    // Remove final dgraphUri field
    private transient DgraphRepository dgraphRepository;
    private transient ExecutorService executorService;
    private transient ParameterTool params;

    // Constructor without parameters
    public AsyncEnrichWithRelationCountFunction() {
    }

    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        params = (ParameterTool) getRuntimeContext().getExecutionConfig().getGlobalJobParameters();
         if (params == null) {
             throw new RuntimeException("Global job parameters (ParameterTool) not found.");
         }

        final String dgraphUri = params.getRequired(ParameterNames.DGRAPH_URI);
        final int threadPoolSize = params.getInt("async.enrich.relation.pool.size", 10); // Optional

        try {
            dgraphRepository = new DgraphRepository(dgraphUri);
            executorService = Executors.newFixedThreadPool(threadPoolSize); 
            LOG.info("Opened DgraphRepository (URI: {}) and ExecutorService (size: {}) for relation count enrichment.", 
                     dgraphUri, threadPoolSize);
        } catch (Exception e) {
            LOG.error("Failed to open Dgraph repository or executor service", e);
            throw new RuntimeException("Failed to initialize AsyncEnrichWithRelationCountFunction", e);
        }
    }

    @Override
    public void close() throws Exception {
        try {
            if (dgraphRepository != null) {
                dgraphRepository.close();
                LOG.info("Closed DgraphRepository.");
            }
        } catch (Exception e) {
            LOG.error("Error closing DgraphRepository", e);
        }
        try {
            if (executorService != null) {
                executorService.shutdownNow();
                LOG.info("Shutdown ExecutorService for relation count.");
            }
        } catch (Exception e) {
             LOG.error("Error shutting down ExecutorService for relation count", e);
        }
        super.close();
    }

    @Override
    public void asyncInvoke(ScoreSnapshot snapshot, ResultFuture<ScoreSnapshotWithRelationCount> resultFuture) throws Exception {
        if (snapshot == null || snapshot.citizenId == null) {
            LOG.warn("Received null snapshot or citizenId. Skipping enrichment.");
             resultFuture.complete(Collections.emptyList()); 
            return;
        }

        final String citizenId = snapshot.citizenId;

        CompletableFuture.supplyAsync(new Supplier<Integer>() {
            @Override
            public Integer get() {
                try {
                    return dgraphRepository.getCitizenRelationCountByExternalId(citizenId);
                } catch (Exception e) {
                    LOG.error("Error fetching relation count for citizen {}: {}", citizenId, e.getMessage(), e);
                    throw new RuntimeException("Failed to get relation count", e);
                }
            }
        }, executorService).thenAcceptAsync((Integer count) -> {
            // Pass original snapshot demographics to the new object constructor
            ScoreSnapshotWithRelationCount enrichedSnapshot = new ScoreSnapshotWithRelationCount(
                snapshot, 
                count
                // The ScoreSnapshotWithRelationCount constructor already copies fields including demographics
            );
            resultFuture.complete(Collections.singleton(enrichedSnapshot));
            LOG.debug("Successfully enriched snapshot for citizen {} with relation count {}", citizenId, count);
        }, executorService)
        .exceptionally((Throwable throwable) -> {
            LOG.error("Async relation count enrichment failed for citizen {}: {}", citizenId, throwable.getMessage());
            resultFuture.complete(Collections.emptyList()); 
            return null; 
        });
    }
    
    @Override
    public void timeout(ScoreSnapshot input, ResultFuture<ScoreSnapshotWithRelationCount> resultFuture) throws Exception {
        LOG.warn("Async relation count enrichment timed out for citizen {}", input != null ? input.citizenId : "null");
        resultFuture.complete(Collections.emptyList());
    }
} 