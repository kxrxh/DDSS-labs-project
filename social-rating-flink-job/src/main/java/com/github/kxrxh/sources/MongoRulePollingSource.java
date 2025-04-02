package com.github.kxrxh.sources; // Новый пакет для источников

import com.github.kxrxh.model.ScoringRule;
import com.github.kxrxh.repositories.MongoRepository;
import org.apache.flink.streaming.api.functions.source.RichSourceFunction; // Используем Rich для доступа к контексту (хотя здесь не строго нужно)
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.TimeUnit;

/**
 * A SourceFunction that periodically polls a MongoDB collection for scoring rules
 * and emits new/updated rules or inactivation markers downstream.
 */
public class MongoRulePollingSource extends RichSourceFunction<ScoringRule> {

    private static final Logger LOG = LoggerFactory.getLogger(MongoRulePollingSource.class);

    private volatile boolean isRunning = true;

    private final String mongoUri;
    private final String mongoDbName;
    private final long pollingIntervalMillis;

    // State to keep track of emitted rules to detect changes and deletions
    private Map<String, ScoringRule> lastEmittedRules = new HashMap<>();

    public MongoRulePollingSource(String mongoUri, String mongoDbName, long pollingIntervalSeconds) {
        this.mongoUri = mongoUri;
        this.mongoDbName = mongoDbName;
        this.pollingIntervalMillis = TimeUnit.SECONDS.toMillis(pollingIntervalSeconds);
        if (pollingIntervalMillis <= 0) {
            throw new IllegalArgumentException("Polling interval must be positive.");
        }
    }

    @Override
    public void run(SourceContext<ScoringRule> ctx) throws Exception {
        LOG.info("Starting MongoDB rule polling source with interval {} ms", pollingIntervalMillis);

        while (isRunning) {
            try (MongoRepository repo = new MongoRepository(mongoUri, mongoDbName)) {
                LOG.debug("Polling MongoDB for active rules...");
                List<ScoringRule> currentActiveRulesList = repo.getActiveScoringRules();
                Map<String, ScoringRule> currentActiveRulesMap = new HashMap<>();
                for (ScoringRule rule : currentActiveRulesList) {
                     if (rule != null && rule.rule_id != null) { // Ensure rule and ID are valid
                        currentActiveRulesMap.put(rule.rule_id, rule);
                     } else {
                        LOG.warn("Fetched a null rule or rule with null rule_id from MongoDB. Skipping.");
                    }
                }
                LOG.debug("Found {} active rules in current poll.", currentActiveRulesMap.size());

                // --- Detect and emit changes ---

                // 1. Emit new or updated rules
                for (Map.Entry<String, ScoringRule> entry : currentActiveRulesMap.entrySet()) {
                    String ruleId = entry.getKey();
                    ScoringRule currentRule = entry.getValue();
                    ScoringRule previousRule = lastEmittedRules.get(ruleId);

                    // Emit if rule is new or has changed (using equals method)
                    if (previousRule == null || !currentRule.equals(previousRule)) {
                        synchronized (ctx.getCheckpointLock()) { // Synchronize emission with checkpointing
                            ctx.collect(currentRule);
                        }
                        LOG.info("Emitted updated/new rule: {}", ruleId);
                    }
                }

                // 2. Detect and emit inactivations/deletions
                Set<String> previousRuleIds = new HashSet<>(lastEmittedRules.keySet());
                previousRuleIds.removeAll(currentActiveRulesMap.keySet()); // IDs that are no longer active

                for (String inactiveRuleId : previousRuleIds) {
                    ScoringRule inactiveRuleMarker = lastEmittedRules.get(inactiveRuleId);
                    if (inactiveRuleMarker != null) {
                        inactiveRuleMarker.isActive = false; // Mark as inactive
                        synchronized (ctx.getCheckpointLock()) {
                            ctx.collect(inactiveRuleMarker); // Emit the rule marked as inactive
                        }
                        LOG.info("Emitted inactivation for rule: {}", inactiveRuleId);
                    }
                }

                // --- Update state for next iteration ---
                lastEmittedRules = currentActiveRulesMap; // Store current rules for next comparison

            } catch (Exception e) {
                LOG.error("Error during MongoDB rule polling: {}", e.getMessage(), e);
                // Avoid busy-looping in case of persistent errors
            }

            // Wait for the next polling interval
            try {
                Thread.sleep(pollingIntervalMillis);
            } catch (InterruptedException e) {
                if (!isRunning) {
                    // Expected interruption on cancel
                    LOG.info("Polling source interrupted and cancelled.");
                    break; // Exit loop
                } else {
                    // Unexpected interruption
                    LOG.warn("Polling source sleep interrupted unexpectedly.", e);
                    Thread.currentThread().interrupt(); // Re-interrupt thread
                }
            }
        }
         LOG.info("MongoDB rule polling source finished.");
    }

    @Override
    public void cancel() {
        isRunning = false;
        LOG.info("Cancelling MongoDB rule polling source.");
    }
} 