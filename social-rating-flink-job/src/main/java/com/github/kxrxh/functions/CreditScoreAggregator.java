package com.github.kxrxh.functions;

import com.github.kxrxh.model.CreditScoreAggregate;
import com.github.kxrxh.model.ScoreSnapshot;
import org.apache.flink.api.common.functions.AggregateFunction;

import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;
import java.util.Objects;

/**
 * Aggregates ScoreSnapshots into CreditScoreAggregate metrics within a window.
 * It calculates the average final score and counts distinct citizens for each
 * combination of geographic dimensions (region, city, district) and score level.
 */
public class CreditScoreAggregator implements AggregateFunction<ScoreSnapshot, CreditScoreAggregator.Accumulator, CreditScoreAggregate> {

    // Accumulator stores intermediate aggregation results
    public static class Accumulator {
        // Store sum of final scores and citizen IDs per unique citizen
        // We need the *latest* finalScore for a citizen within the window if they appear multiple times
        Map<String, Double> lastFinalScores = new HashMap<>();
        // Using a Map helps get the latest score, but we need a Set for counting later
        Set<String> distinctCitizens = new HashSet<>();

        // Dimensions (initialize with first element's values)
        String region = null;
        String city = null;
        String district = null;
        String level = null; // The level associated with this group
        String ageGroup = null; // Add demographic dimensions
        String gender = null;
        boolean dimensionsSet = false;

        void setDimensions(ScoreSnapshot snapshot) {
            if (!dimensionsSet) {
                this.region = snapshot.region;
                this.city = snapshot.city;
                this.district = snapshot.district;
                this.level = snapshot.level; // Grouping includes level
                this.ageGroup = snapshot.ageGroup; // Set demographic dimensions
                this.gender = snapshot.gender;
                this.dimensionsSet = true;
            }
        }
    }

    @Override
    public Accumulator createAccumulator() {
        return new Accumulator();
    }

    @Override
    public Accumulator add(ScoreSnapshot value, Accumulator accumulator) {
        // Ensure dimensions are set based on the first element in the group
        accumulator.setDimensions(value);

        // We only add the citizen and their score if their level matches the accumulator's level
        // AND if their demographics match the accumulator's demographics
        // (Since the keying includes all these, this check might be redundant but safe)
        if (accumulator.level != null && accumulator.level.equals(value.level) &&
            Objects.equals(accumulator.ageGroup, value.ageGroup) && // Check demographics too
            Objects.equals(accumulator.gender, value.gender)) 
        {
            // Store/Update the *latest* final score for this citizen in this window
            accumulator.lastFinalScores.put(value.citizenId, value.finalScore);
            accumulator.distinctCitizens.add(value.citizenId);
        }
        // Else: Ignore snapshots that don't match the group for this accumulator instance

        return accumulator;
    }

    @Override
    public CreditScoreAggregate getResult(Accumulator accumulator) {
        CreditScoreAggregate result = new CreditScoreAggregate();

        // Set dimensions
        result.region = accumulator.region;
        result.city = accumulator.city;
        result.district = accumulator.district;
        result.level = accumulator.level;
        result.ageGroup = accumulator.ageGroup; // Set demographic dimensions
        result.gender = accumulator.gender;

        // Calculate metrics
        result.citizenCount = accumulator.distinctCitizens.size();
        double totalScoreSum = accumulator.lastFinalScores.values().stream().mapToDouble(Double::doubleValue).sum();

        // Calculate average score, handle division by zero
        result.avgScore = (result.citizenCount > 0)
                ? totalScoreSum / result.citizenCount
                : 0.0;

        // Window start/end will be set later by the ProcessWindowFunction
        result.windowStart = null;
        result.windowEnd = null;

        return result;
    }

    @Override
    public Accumulator merge(Accumulator a, Accumulator b) {
        // Merge citizen scores, keeping the latest score if a citizen exists in both
        b.lastFinalScores.forEach(a.lastFinalScores::put);
        a.distinctCitizens.addAll(b.distinctCitizens);

        // Dimensions should be the same within a merged group, take from 'a'
        if (!a.dimensionsSet && b.dimensionsSet) {
           a.region = b.region;
           a.city = b.city;
           a.district = b.district;
           a.level = b.level;
           a.ageGroup = b.ageGroup; // Merge demographic dimensions
           a.gender = b.gender;
           a.dimensionsSet = true;
        }

        return a;
    }
} 