package com.github.kxrxh.model;

import java.sql.Timestamp;
import java.util.Objects;

/**
 * Represents aggregated social graph metrics within a specific time window
 * for a given geographic location.
 * Matches the schema of the 'social_graph_aggregate' ClickHouse table.
 */
public class SocialGraphAggregate {

    // Time window boundaries
    public Long windowStart; // Populated by ProcessWindowFunction
    public Long windowEnd;   // Populated by ProcessWindowFunction

    // Dimensions
    public String region;
    public String city;
    public String district;
    public String ageGroup; // New demographic dimension
    public String gender;   // New demographic dimension

    // Metrics
    public long totalRelations;         // Sum of relationship counts for all citizens in this group/window
    public long distinctCitizenCount;   // Count of distinct citizens in this group/window
    public double avgRelationsPerCitizen; // Calculated metric: totalRelations / distinctCitizenCount

    public SocialGraphAggregate() {
    }

    /**
     * Helper method to get the window end time as a java.sql.Timestamp,
     * suitable for ClickHouse DateTime columns.
     * Returns null if windowEnd is not set.
     */
    public Timestamp getWindowEndDateTime() {
        return (this.windowEnd != null) ? new Timestamp(this.windowEnd) : null;
    }

    // Getters (optional, fields are public)

    @Override
    public String toString() {
        return "SocialGraphAggregate{" +
               "windowStart=" + windowStart +
               ", windowEnd=" + windowEnd +
               ", region='" + region + "'" +
               ", city='" + city + "'" +
               ", district='" + district + "'" +
               ", ageGroup='" + ageGroup + "'" +
               ", gender='" + gender + "'" +
               ", totalRelations=" + totalRelations +
               ", distinctCitizenCount=" + distinctCitizenCount +
               ", avgRelationsPerCitizen=" + avgRelationsPerCitizen +
               "}";
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        SocialGraphAggregate that = (SocialGraphAggregate) o;
        return totalRelations == that.totalRelations &&
               distinctCitizenCount == that.distinctCitizenCount &&
               Double.compare(that.avgRelationsPerCitizen, avgRelationsPerCitizen) == 0 &&
               Objects.equals(windowStart, that.windowStart) &&
               Objects.equals(windowEnd, that.windowEnd) &&
               Objects.equals(region, that.region) &&
               Objects.equals(city, that.city) &&
               Objects.equals(district, that.district) &&
               Objects.equals(ageGroup, that.ageGroup) &&
               Objects.equals(gender, that.gender);
    }

    @Override
    public int hashCode() {
        return Objects.hash(windowStart, windowEnd, region, city, district, 
                          ageGroup, gender,
                          totalRelations, distinctCitizenCount, avgRelationsPerCitizen);
    }
} 