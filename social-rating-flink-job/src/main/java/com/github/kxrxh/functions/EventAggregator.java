package com.github.kxrxh.functions;

import com.github.kxrxh.model.EventAggregate;
import com.github.kxrxh.model.ScoreSnapshot;
import org.apache.flink.api.common.functions.AggregateFunction;

import java.util.HashSet;
import java.util.Set;

/**
 * Aggregates ScoreSnapshots into EventAggregate metrics within a window.
 */
public class EventAggregator implements AggregateFunction<ScoreSnapshot, EventAggregator.Accumulator, EventAggregate> {

    // Accumulator stores intermediate aggregation results
    public static class Accumulator {
        long count = 0L;
        double totalScoreImpact = 0.0;
        Set<String> distinctCitizens = new HashSet<>(); // To count distinct citizens

        // Dimensions (initialize with first element's values)
        String region = null;
        String city = null;
        String district = null;
        String eventType = null;
        String eventSubtype = null;
        String ageGroup = null; // Add demographic dimensions
        String gender = null;
        boolean dimensionsSet = false;

        void setDimensions(ScoreSnapshot snapshot) {
             if (!dimensionsSet) {
                 this.region = snapshot.region;
                 this.city = snapshot.city;
                 this.district = snapshot.district;
                 this.eventType = snapshot.eventType;
                 this.eventSubtype = snapshot.eventSubtype;
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
        // Initialize dimensions from the first element
        accumulator.setDimensions(value);

        // Update metrics
        accumulator.count++;
        accumulator.totalScoreImpact += value.scoreChange;
        accumulator.distinctCitizens.add(value.citizenId);

        return accumulator;
    }

    @Override
    public EventAggregate getResult(Accumulator accumulator) {
        EventAggregate result = new EventAggregate();

        // Set dimensions
        result.region = accumulator.region;
        result.city = accumulator.city;
        result.district = accumulator.district;
        result.eventType = accumulator.eventType;
        result.eventSubtype = accumulator.eventSubtype;
        result.ageGroup = accumulator.ageGroup; // Set demographic dimensions
        result.gender = accumulator.gender;

        // Set calculated metrics
        result.eventCount = accumulator.count;
        result.totalScoreImpact = accumulator.totalScoreImpact;
        result.distinctCitizenCount = accumulator.distinctCitizens.size();
        // Calculate average, handle division by zero
        result.avgScoreImpact = (accumulator.count > 0)
                ? accumulator.totalScoreImpact / accumulator.count
                : 0.0;

        // Window start/end will be set later by the ProcessWindowFunction if needed
        result.windowStart = null;
        result.windowEnd = null;


        return result;
    }

    @Override
    public Accumulator merge(Accumulator a, Accumulator b) {
        // Merge metrics
        a.count += b.count;
        a.totalScoreImpact += b.totalScoreImpact;
        a.distinctCitizens.addAll(b.distinctCitizens); // Merge sets for distinct count

        // Dimensions should be the same within a merged group, take from 'a'
        if (!a.dimensionsSet && b.dimensionsSet) {
           a.region = b.region;
           a.city = b.city;
           a.district = b.district;
           a.eventType = b.eventType;
           a.eventSubtype = b.eventSubtype;
           a.ageGroup = b.ageGroup; // Merge demographic dimensions
           a.gender = b.gender;
           a.dimensionsSet = true;
        }

        return a;
    }
} 