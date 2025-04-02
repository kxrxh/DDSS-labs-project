package com.github.kxrxh.functions;

import com.github.kxrxh.model.CitizenEvent;
import com.github.kxrxh.model.ScoringRule;
import com.github.kxrxh.model.ScoreUpdate;
import org.apache.flink.api.common.state.MapStateDescriptor;
import org.apache.flink.api.common.state.ReadOnlyBroadcastState;
import org.apache.flink.streaming.api.functions.co.BroadcastProcessFunction;
import org.apache.flink.util.Collector;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;
import java.util.Objects;

/**
 * BroadcastProcessFunction to apply scoring rules to citizen events.
 * Rules are received via broadcast state.
 */
public class ApplyRulesFunction extends BroadcastProcessFunction<CitizenEvent, ScoringRule, ScoreUpdate> {

    private static final Logger LOG = LoggerFactory.getLogger(ApplyRulesFunction.class);

    private final MapStateDescriptor<String, ScoringRule> rulesStateDescriptor;

    /**
     * Constructor accepting the state descriptor.
     * @param rulesStateDescriptor The descriptor for the broadcast state containing scoring rules.
     */
    public ApplyRulesFunction(MapStateDescriptor<String, ScoringRule> rulesStateDescriptor) {
        this.rulesStateDescriptor = Objects.requireNonNull(rulesStateDescriptor, "Rules state descriptor must not be null");
    }

    /**
     * Called for each incoming CitizenEvent from the non-broadcast side.
     */
    @Override
    public void processElement(CitizenEvent event, ReadOnlyContext ctx, Collector<ScoreUpdate> out) throws Exception {
        LOG.debug("Processing event: {}", event.eventId);
        ReadOnlyBroadcastState<String, ScoringRule> rulesState = ctx.getBroadcastState(rulesStateDescriptor);

        for (Map.Entry<String, ScoringRule> entry : rulesState.immutableEntries()) {
            ScoringRule rule = entry.getValue();
            LOG.trace("Checking rule: {} for event {}", rule.rule_id, event.eventId);

            if (rule.isActive != null && rule.isActive && matches(event, rule)) {
                LOG.info("Event {} matched rule {}", event.eventId, rule.rule_id);

                // --- Calculate final score change --- 
                double basePoints = 0.0;
                double multiplier = 1.0; // Default multiplier

                if (rule.scoreChange != null) {
                    if (rule.scoreChange.points != null) {
                        basePoints = rule.scoreChange.points;
                    }
                    // Try to parse multiplierExpression as a double
                    if (rule.scoreChange.multiplierExpression != null) {
                        try {
                            multiplier = Double.parseDouble(rule.scoreChange.multiplierExpression);
                        } catch (NumberFormatException e) {
                            LOG.warn("Rule {}: Could not parse multiplierExpression '{}' as double. Using default 1.0.", 
                                     rule.rule_id, rule.scoreChange.multiplierExpression);
                            multiplier = 1.0;
                        }
                    } else {
                        LOG.trace("Rule {}: multiplierExpression is null. Using default 1.0.", rule.rule_id);
                    }
                } else {
                     LOG.warn("Rule {}: scoreChange details are missing. Using 0.0 base points and 1.0 multiplier.", rule.rule_id);
                }

                double finalChange = basePoints * multiplier;
                // ----------------------------------------

                ScoreUpdate update = new ScoreUpdate(
                    event.citizenId,
                    finalChange, // Use the calculated final change
                    event.timestamp,
                    rule.rule_id,
                    event.eventId
                );
                out.collect(update);

                // Optional: Add logic for rule priority (e.g., only apply highest priority rule)
                // break; // If only one rule should match per event
            }
        }
    }

    /**
     * Called for each incoming ScoringRule from the broadcast side.
     * Updates the broadcast state.
     */
    @Override
    public void processBroadcastElement(ScoringRule rule, Context ctx, Collector<ScoreUpdate> out) throws Exception {
        LOG.info("Updating broadcast state for rule: {}. Active: {}", rule.rule_id, rule.isActive);
        org.apache.flink.api.common.state.BroadcastState<String, ScoringRule> rulesState = ctx.getBroadcastState(rulesStateDescriptor);
        if (rule.rule_id == null) {
            LOG.warn("Received rule with null rule_id. Skipping update.");
            return;
        }
        if (rule.isActive != null && rule.isActive) {
            rulesState.put(rule.rule_id, rule);
        } else {
            rulesState.remove(rule.rule_id);
            LOG.info("Removed rule {} from broadcast state (inactive or null isActive).", rule.rule_id);
        }
    }

    private boolean matches(CitizenEvent event, ScoringRule rule) {
        if (!Objects.equals(event.eventType, rule.eventType)) {
            LOG.trace("Rule {} mismatch: eventType ('{}' vs '{}')", rule.rule_id, event.eventType, rule.eventType);
            return false;
        }

        if (rule.eventSubtype != null && !Objects.equals(event.eventSubtype, rule.eventSubtype)) {
            LOG.trace("Rule {} mismatch: eventSubtype ('{}' vs '{}')", rule.rule_id, event.eventSubtype, rule.eventSubtype);
            return false;
        }

        if (rule.conditions != null && !rule.conditions.isEmpty()) {
            for (Map.Entry<String, Object> condition : rule.conditions.entrySet()) {
                if (!checkCondition(event, rule.rule_id, condition.getKey(), condition.getValue())) {
                    return false;
                }
            }
        }

        return true;
    }

    private boolean checkCondition(CitizenEvent event, String ruleId, String conditionKey, Object expectedValue) {
        Object actualValue = null;

        if ("sourceSystem".equals(conditionKey)) {
            actualValue = event.sourceSystem;
        } else if (conditionKey != null && conditionKey.startsWith("payload.") && event.payload != null) {
            String payloadKey = conditionKey.substring(8);
            actualValue = event.payload.get(payloadKey);
        } else if ("citizenId".equals(conditionKey)) {
            actualValue = event.citizenId;
        }

        boolean match = compareValues(actualValue, expectedValue);
        if (!match) {
            LOG.trace("Rule {} condition mismatch: key='{}', expected='{}'({}), actual='{}'({})",
                    ruleId, conditionKey, expectedValue, expectedValue != null ? expectedValue.getClass().getSimpleName() : "null",
                    actualValue, actualValue != null ? actualValue.getClass().getSimpleName() : "null");
        }
        return match;
    }

    private boolean compareValues(Object actualValue, Object expectedValue) {
        if (Objects.equals(actualValue, expectedValue)) {
            return true;
        }
        if (actualValue == null || expectedValue == null) {
            return false;
        }

        String actualStr = String.valueOf(actualValue);
        String expectedStr = String.valueOf(expectedValue);
        if (actualStr.equals(expectedStr)) {
            return true;
        }

        if (expectedValue instanceof Boolean || expectedStr.equalsIgnoreCase("true") || expectedStr.equalsIgnoreCase("false")) {
            try {
                boolean actualBool = Boolean.parseBoolean(actualStr);
                boolean expectedBool = Boolean.parseBoolean(expectedStr);
                return actualBool == expectedBool;
            } catch (Exception e) { /* Ignore parsing error, proceed to other checks */ }
        }

        if (expectedValue instanceof Number) {
            try {
                double actualDouble = Double.parseDouble(actualStr);
                double expectedDouble = ((Number) expectedValue).doubleValue();
                return actualDouble == expectedDouble;
            } catch (NumberFormatException e) {
                return false;
            }
        }

        return false;
    }
} 