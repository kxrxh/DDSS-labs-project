package com.github.kxrxh.functions;

import com.github.kxrxh.model.CitizenEvent;
import com.github.kxrxh.model.RelatedEffect;
import com.github.kxrxh.model.ScoringRule;
import com.github.kxrxh.model.ScoreSnapshot;
import com.github.kxrxh.model.RelatedEffectTrigger;
import org.apache.flink.api.common.state.MapStateDescriptor;
import org.apache.flink.api.common.state.ValueState;
import org.apache.flink.api.common.state.ValueStateDescriptor;
import org.apache.flink.api.common.typeinfo.TypeHint;
import org.apache.flink.api.common.typeinfo.TypeInformation;
import org.apache.flink.api.common.typeinfo.Types;
import org.apache.flink.api.java.tuple.Tuple2;
import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.co.KeyedBroadcastProcessFunction;
import org.apache.flink.util.Collector;
import org.apache.flink.util.OutputTag;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;
import java.util.Objects;

/**
 * KeyedBroadcastProcessFunction to apply scoring rules statefully.
 * Maintains the current score for each citizen in ValueState.
 * Emits ScoreSnapshot containing the final score and level.
 * Emits RelatedEffectTrigger to a side output if a rule has effectOnRelated.
 * Emits the first seen event timestamp for each citizen to a side output.
 */
public class StatefulApplyRulesFunction extends KeyedBroadcastProcessFunction<String, CitizenEvent, ScoringRule, ScoreSnapshot> {

    private static final Logger LOG = LoggerFactory.getLogger(StatefulApplyRulesFunction.class);

    private final MapStateDescriptor<String, ScoringRule> rulesStateDescriptor;
    private transient ValueState<Double> currentScoreState;
    private transient ValueState<Boolean> firstEventSeenState;

    public static final OutputTag<RelatedEffectTrigger> relatedEffectOutputTag =
        new OutputTag<RelatedEffectTrigger>("related-effect-triggers"){};

    public static final OutputTag<Tuple2<String, Long>> firstSeenOutputTag =
        new OutputTag<Tuple2<String, Long>>("first-seen-timestamps"){};

    /**
     * Constructor accepting the broadcast state descriptor.
     * @param rulesStateDescriptor The descriptor for the broadcast state containing scoring rules.
     */
    public StatefulApplyRulesFunction(MapStateDescriptor<String, ScoringRule> rulesStateDescriptor) {
        this.rulesStateDescriptor = Objects.requireNonNull(rulesStateDescriptor, "Rules state descriptor must not be null");
    }

    /**
     * Initialize the ValueState for the citizen's score and the first event flag.
     */
    @Override
    public void open(Configuration parameters) throws Exception {
        super.open(parameters);
        ValueStateDescriptor<Double> scoreDescriptor = new ValueStateDescriptor<>(
                "currentCitizenScore",
                Types.DOUBLE
        );
        currentScoreState = getRuntimeContext().getState(scoreDescriptor);

        ValueStateDescriptor<Boolean> firstSeenDescriptor = new ValueStateDescriptor<>(
                "firstEventSeen",
                Types.BOOLEAN
        );
        firstEventSeenState = getRuntimeContext().getState(firstSeenDescriptor);

        LOG.info("Initialized ValueState for citizen scores and first event flag.");
    }

    /**
     * Process incoming CitizenEvent for a specific citizen (key).
     * Checks if it's the first event seen for the citizen and emits to side output.
     * Applies rules, updates the citizen's score state, calculates level and emits ScoreSnapshot.
     */
    @Override
    public void processElement(CitizenEvent event, ReadOnlyContext ctx, Collector<ScoreSnapshot> out) throws Exception {
        LOG.debug("Processing event: {} for citizen key: {}", event.eventId, ctx.getCurrentKey());

        Boolean seenBefore = firstEventSeenState.value();
        if (seenBefore == null || !seenBefore) {
            if (event.timestamp != null) {
                ctx.output(firstSeenOutputTag, Tuple2.of(ctx.getCurrentKey(), event.timestamp));
                LOG.info("First event seen for citizen {}. Emitting timestamp {}.", ctx.getCurrentKey(), event.timestamp);
            } else {
                 LOG.warn("First event for citizen {} has null timestamp. Cannot emit to firstSeenOutputTag.", ctx.getCurrentKey());
            }
            firstEventSeenState.update(true);
        }

        Double currentScore = currentScoreState.value();
        if (currentScore == null) {
            currentScore = 0.0;
            LOG.debug("Initializing score state for citizen {}", ctx.getCurrentKey());
        }

        boolean ruleMatched = false;

        for (Map.Entry<String, ScoringRule> entry : ctx.getBroadcastState(rulesStateDescriptor).immutableEntries()) {
            ScoringRule rule = entry.getValue();
            LOG.trace("Checking rule: {} for event {} (Citizen: {})", rule.rule_id, event.eventId, ctx.getCurrentKey());

            if (rule.isActive != null && rule.isActive && matches(event, rule)) {
                ruleMatched = true;
                LOG.info("Event {} matched rule {} for citizen {}", event.eventId, rule.rule_id, ctx.getCurrentKey());

                double basePoints = 0.0;
                double multiplier = 1.0;
                if (rule.scoreChange != null) {
                    basePoints = (rule.scoreChange.points != null) ? rule.scoreChange.points : 0.0;
                    String expression = rule.scoreChange.multiplierExpression;
                    if (expression != null) {
                        try {
                            String evaluatedExpression = expression.replace("currentScore", String.valueOf(currentScore));
                            multiplier = Double.parseDouble(evaluatedExpression);
                            LOG.trace("Rule {}: Evaluated multiplier expression '{}' to {}", rule.rule_id, evaluatedExpression, multiplier);
                        } catch (Exception e) {
                            LOG.warn("Rule {}: Error evaluating multiplier expression '{}'. Using 1.0.", rule.rule_id, expression, e);
                            multiplier = 1.0;
                        }
                    }
                } else {
                     LOG.warn("Rule {}: scoreChange details missing. Using 0 points.", rule.rule_id);
                }
                double ruleScoreChange = basePoints * multiplier;
                double newScore = currentScore + ruleScoreChange;

                currentScoreState.update(newScore);
                LOG.debug("Updated score state for citizen {}: {} -> {}", ctx.getCurrentKey(), currentScore, newScore);
                currentScore = newScore;

                String level = calculateLevel(newScore);

                ScoreSnapshot snapshot = new ScoreSnapshot(
                    event.citizenId,
                    newScore,
                    level,
                    ruleScoreChange,
                    event.timestamp,
                    event.eventId,
                    rule.rule_id,
                    event.eventType,
                    event.eventSubtype,
                    event.location != null ? event.location.region : null,
                    event.location != null ? event.location.city : null,
                    event.location != null ? event.location.district : null,
                    event.ageGroup,
                    event.gender
                );
                out.collect(snapshot);

                if (rule.effectOnRelated != null && 
                    rule.effectOnRelated.percentage != null && rule.effectOnRelated.percentage != 0 && 
                    rule.effectOnRelated.relationType != null && !rule.effectOnRelated.relationType.isEmpty()) 
                {
                    RelatedEffectTrigger trigger = new RelatedEffectTrigger(
                        event.citizenId,
                        ruleScoreChange,
                        event.eventId,
                        rule.effectOnRelated,
                        rule.rule_id
                    );
                    ctx.output(relatedEffectOutputTag, trigger);
                     LOG.info("Emitted RelatedEffectTrigger for event {}, rule {}", event.eventId, rule.rule_id);
                }
            }
        }
        
        if (!ruleMatched) {
             LOG.debug("Event {} did not match any active rules for citizen {}", event.eventId, ctx.getCurrentKey());
        }
    }

    /**
     * Process incoming ScoringRule from the broadcast side.
     * Updates the broadcast state.
     */
    @Override
    public void processBroadcastElement(ScoringRule rule, Context ctx, Collector<ScoreSnapshot> out) throws Exception {
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

    private String calculateLevel(double score) {
        if (score >= 500) return "A";
        if (score >= 300) return "B";
        if (score >= 100) return "C";
        if (score >= 0) return "D";
        return "E";
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
             } catch (Exception e) { /* Ignore */ }
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