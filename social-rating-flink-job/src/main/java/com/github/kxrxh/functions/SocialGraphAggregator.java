package com.github.kxrxh.functions;

import com.github.kxrxh.model.ScoreSnapshotWithRelationCount;
import com.github.kxrxh.model.SocialGraphAggregate;
import org.apache.flink.api.common.functions.AggregateFunction;

import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;
import java.util.Objects;

/**
 * Aggregates enriched ScoreSnapshots (with relation count) into SocialGraphAggregate metrics.
 * Calculates total relations and average relations per citizen within a geographic group and window.
 */
public class SocialGraphAggregator implements AggregateFunction<
        ScoreSnapshotWithRelationCount, 
        SocialGraphAggregator.Accumulator, 
        SocialGraphAggregate> {

    // Accumulator stores intermediate results
    public static class Accumulator {
        // Keep track of the latest relation count per citizen within the window
        Map<String, Integer> lastRelationCounts = new HashMap<>(); 
        Set<String> distinctCitizens = new HashSet<>();

        // Dimensions
        String region = null;
        String city = null;
        String district = null;
        String ageGroup = null;
        String gender = null;
        boolean dimensionsSet = false;

        void setDimensions(ScoreSnapshotWithRelationCount snapshot) {
            if (!dimensionsSet) {
                this.region = snapshot.region;
                this.city = snapshot.city;
                this.district = snapshot.district;
                this.ageGroup = snapshot.ageGroup;
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
    public Accumulator add(ScoreSnapshotWithRelationCount value, Accumulator accumulator) {
        accumulator.setDimensions(value);
        
        if (Objects.equals(accumulator.ageGroup, value.ageGroup) &&
            Objects.equals(accumulator.gender, value.gender))
        {
            // Store/Update the latest relation count for this citizen in the window
            accumulator.lastRelationCounts.put(value.citizenId, value.relationCount);
            accumulator.distinctCitizens.add(value.citizenId);
        }
        // Else: Ignore snapshots that don't match the group for this accumulator instance

        return accumulator;
    }

    @Override
    public SocialGraphAggregate getResult(Accumulator accumulator) {
        SocialGraphAggregate result = new SocialGraphAggregate();

        result.region = accumulator.region;
        result.city = accumulator.city;
        result.district = accumulator.district;
        result.ageGroup = accumulator.ageGroup;
        result.gender = accumulator.gender;

        // Sum the *last known* relation count for each distinct citizen in the window
        result.totalRelations = accumulator.lastRelationCounts.values().stream()
            .mapToInt(Integer::intValue)
            .sum();

        result.distinctCitizenCount = accumulator.distinctCitizens.size();

        // Calculate average, handle division by zero
        result.avgRelationsPerCitizen = (result.distinctCitizenCount > 0)
                ? (double) result.totalRelations / result.distinctCitizenCount
                : 0.0;

        // windowStart/End set by ProcessWindowFunction
        result.windowStart = null;
        result.windowEnd = null;

        return result;
    }

    @Override
    public Accumulator merge(Accumulator a, Accumulator b) {
        // Merge relation counts, keeping the latest from 'b' if citizen exists in both
        b.lastRelationCounts.forEach(a.lastRelationCounts::put);
        a.distinctCitizens.addAll(b.distinctCitizens);

        if (!a.dimensionsSet && b.dimensionsSet) {
            a.region = b.region;
            a.city = b.city;
            a.district = b.district;
            a.ageGroup = b.ageGroup;
            a.gender = b.gender;
            a.dimensionsSet = true;
        }
        return a;
    }
} 