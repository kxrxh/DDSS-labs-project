package com.github.kxrxh.model;

import java.sql.Timestamp;
import java.util.Objects;

/**
 * Represents aggregated credit score metrics within a specific time window
 * for a given geographic location and score level.
 * Matches the schema of the 'credit_score_aggregate' ClickHouse table.
 */
public class CreditScoreAggregate {

    // Time window boundaries
    public Long windowStart; // Populated by ProcessWindowFunction
    public Long windowEnd;   // Populated by ProcessWindowFunction

    // Dimensions
    public String region;
    public String city;
    public String district;
    public String level; // The score level (A, B, C, etc.)
    public String ageGroup; // New demographic dimension
    public String gender;   // New demographic dimension

    // Metrics
    public double avgScore;      // Average final score for citizens in this group/window
    public long citizenCount; // Count of distinct citizens in this group/window

    public CreditScoreAggregate() {
    }

    // Getters (optional, fields are public)

    /**
     * Helper method to get the window end time as a java.sql.Timestamp,
     * suitable for ClickHouse DateTime columns.
     * Returns null if windowEnd is not set.
     */
    public Timestamp getWindowEndDateTime() {
        return (this.windowEnd != null) ? new Timestamp(this.windowEnd) : null;
    }

    @Override
    public String toString() {
        return "CreditScoreAggregate{" +
               "windowStart=" + windowStart +
               ", windowEnd=" + windowEnd +
               ", region='" + region + "'" +
               ", city='" + city + "'" +
               ", district='" + district + "'" +
               ", level='" + level + "'" +
               ", ageGroup='" + ageGroup + "'" +
               ", gender='" + gender + "'" +
               ", avgScore=" + avgScore +
               ", citizenCount=" + citizenCount +
               "}";
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        CreditScoreAggregate that = (CreditScoreAggregate) o;
        return Double.compare(that.avgScore, avgScore) == 0 &&
               citizenCount == that.citizenCount &&
               Objects.equals(windowStart, that.windowStart) &&
               Objects.equals(windowEnd, that.windowEnd) &&
               Objects.equals(region, that.region) &&
               Objects.equals(city, that.city) &&
               Objects.equals(district, that.district) &&
               Objects.equals(level, that.level) &&
               Objects.equals(ageGroup, that.ageGroup) &&
               Objects.equals(gender, that.gender);
    }

    @Override
    public int hashCode() {
        return Objects.hash(windowStart, windowEnd, region, city, district, level, 
                          ageGroup, gender, avgScore, citizenCount);
    }
} 