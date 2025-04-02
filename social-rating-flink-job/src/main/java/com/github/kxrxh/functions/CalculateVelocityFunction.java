package com.github.kxrxh.functions;

import com.github.kxrxh.model.ScoreSnapshot;
import com.github.kxrxh.model.ScoreVelocityMetrics;
import org.apache.flink.api.common.state.MapState;
import org.apache.flink.api.common.state.MapStateDescriptor;
import org.apache.flink.api.common.typeinfo.Types;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.KeyedProcessFunction;
import org.apache.flink.util.Collector;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Iterator;
import java.util.Map;

/**
 * Calculates score velocity metrics over 1-hour and 24-hour rolling windows
 * using processing time timers and MapState to store recent score changes.
 * Keyed by location string (e.g., "region|city|district").
 */
public class CalculateVelocityFunction extends KeyedProcessFunction<String, ScoreSnapshot, ScoreVelocityMetrics> {

    private static final Logger LOG = LoggerFactory.getLogger(CalculateVelocityFunction.class);

    // Configuration for calculation intervals and state retention
    private static final long CALCULATION_INTERVAL_MS = 5 * 60 * 1000L; // Calculate every 5 minutes
    private static final long ONE_HOUR_MS = 60 * 60 * 1000L;
    private static final long TWENTY_FOUR_HOURS_MS = 24 * ONE_HOUR_MS;
    // Keep state slightly longer than needed for calculations to handle clock skew/timer delays
    private static final long STATE_RETENTION_MS = TWENTY_FOUR_HOURS_MS + ONE_HOUR_MS; 

    // State to store recent score changes: Timestamp -> ScoreChange
    private transient MapState<Long, Double> recentScoreChanges;

    @Override
    public void open(Configuration parameters) throws Exception {
        MapStateDescriptor<Long, Double> descriptor = new MapStateDescriptor<>(
                "recentScoreChangesState",
                Types.LONG,  // Timestamp (ms)
                Types.DOUBLE // Score Change
        );
        recentScoreChanges = getRuntimeContext().getMapState(descriptor);
        LOG.info("Initialized MapState for recent score changes.");
    }

    @Override
    public void processElement(ScoreSnapshot snapshot, Context ctx, Collector<ScoreVelocityMetrics> out) throws Exception {
        // Use eventTimestamp from the snapshot as the time associated with the score change
        long eventTimestamp = snapshot.eventTimestamp; 
        long currentProcessingTime = ctx.timerService().currentProcessingTime();

        if (eventTimestamp <= 0) { // Check for invalid timestamp
             LOG.warn("Snapshot for citizen {} has invalid timestamp ({}). Cannot process for velocity. Event ID: {}", 
                      snapshot.citizenId, eventTimestamp, snapshot.sourceEventId);
             return; // Cannot process without a valid timestamp
        }

        // 1. Add current score change to state
        recentScoreChanges.put(eventTimestamp, snapshot.scoreChange);
        LOG.trace("Added score change {} at event time {} for key {}", snapshot.scoreChange, eventTimestamp, ctx.getCurrentKey());

        // 2. Clean up old state (based on event time)
        long cleanupThreshold = eventTimestamp - STATE_RETENTION_MS; // Clean based on *event* time
        Iterator<Map.Entry<Long, Double>> iterator = recentScoreChanges.iterator();
        int removedCount = 0;
        while (iterator.hasNext()) {
            Map.Entry<Long, Double> entry = iterator.next();
            if (entry.getKey() < cleanupThreshold) { 
                 iterator.remove();
                 removedCount++;
            }
        }
         if (removedCount > 0) {
             LOG.debug("Cleaned up {} old score change entries (older than {}) for key {}", 
                       removedCount, cleanupThreshold, ctx.getCurrentKey());
         }

        // 3. Register a timer based on processing time to perform the calculation periodically
        // Fire CALCULATION_INTERVAL_MS *after* the current processing time
        long timerTimestamp = currentProcessingTime + CALCULATION_INTERVAL_MS;
        ctx.timerService().registerProcessingTimeTimer(timerTimestamp);
        LOG.trace("Registered processing time timer for {} for key {}", timerTimestamp, ctx.getCurrentKey());
    }

    @Override
    public void onTimer(long timestamp, OnTimerContext ctx, Collector<ScoreVelocityMetrics> out) throws Exception {
        // Use the time the timer *actually* fires as the reference point 'now'
        long calculationTime = ctx.timerService().currentProcessingTime(); 
        LOG.debug("Processing timer fired at {} (scheduled for {}) for key {}", 
                  calculationTime, timestamp, ctx.getCurrentKey());

        // Define time boundaries for calculations based on calculationTime
        long oneHourAgo = calculationTime - ONE_HOUR_MS;
        long twentyFourHoursAgo = calculationTime - TWENTY_FOUR_HOURS_MS;

        // Variables to calculate metrics
        double sum1h = 0.0;
        int count1h = 0;
        long positive1h = 0;
        long negative1h = 0;
        double sum24h = 0.0;
        int count24h = 0;

        // Iterate through the state (event timestamps)
        for (Map.Entry<Long, Double> entry : recentScoreChanges.entries()) {
            long eventTime = entry.getKey();
            double scoreChange = entry.getValue();

            // Check if event time is within 24-hour window ending now (calculationTime)
            if (eventTime >= twentyFourHoursAgo && eventTime < calculationTime) {
                sum24h += scoreChange;
                count24h++;

                // Check if also within 1-hour window ending now
                if (eventTime >= oneHourAgo) {
                    sum1h += scoreChange;
                    count1h++;
                    if (scoreChange > 0) {
                        positive1h++;
                    } else if (scoreChange < 0) {
                        negative1h++;
                    }
                }
            }
        }

        // Calculate averages, handle division by zero
        double avgChange1h = (count1h > 0) ? sum1h / count1h : 0.0;
        double avgChange24h = (count24h > 0) ? sum24h / count24h : 0.0;

        // Create the result object
        ScoreVelocityMetrics metrics = new ScoreVelocityMetrics();
        // Extract dimensions from the key (expecting "region|city|district")
        String[] keyParts = ctx.getCurrentKey().split("\\|"); // Escape pipe for regex
        metrics.region = keyParts.length > 0 && !"NULL".equals(keyParts[0]) ? keyParts[0] : null;
        metrics.city = keyParts.length > 1 && !"NULL".equals(keyParts[1]) ? keyParts[1] : null;
        metrics.district = keyParts.length > 2 && !"NULL".equals(keyParts[2]) ? keyParts[2] : null;
        
        metrics.calculationTimestamp = calculationTime; // Use timer firing time as reference
        metrics.scoreChange1h = avgChange1h;
        metrics.scoreChange24h = avgChange24h;
        metrics.positiveEvents1h = positive1h;
        metrics.negativeEvents1h = negative1h;

        // Only emit if there were events in the last 24 hours for this key
        if (count24h > 0) {
            out.collect(metrics);
            LOG.debug("Emitted velocity metrics for key {}: {}", ctx.getCurrentKey(), metrics);
        } else {
             LOG.debug("No score changes found in the last 24 hours for key {}. Skipping emission.", ctx.getCurrentKey());
        }
        
        // No need to register the next timer explicitly here.
        // Timers are registered in processElement, ensuring they only exist
        // when data is flowing for the key.
    }
} 