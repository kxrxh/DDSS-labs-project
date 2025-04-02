package com.github.kxrxh.model;

import java.sql.Timestamp;
import java.util.Objects;

/**
 * Represents aggregated event metrics within a specific time window
 * for a given geographic location, event type, and demographic group.
 * Matches the schema of the 'events_aggregate' ClickHouse table.
 */
public class EventAggregate {

    // Time window boundaries
    public Long windowStart; // Populated by ProcessWindowFunction
    public Long windowEnd;   // Populated by ProcessWindowFunction

    // Dimensions
    public String region;
    public String city;
    public String district;
    public String eventType;
    public String eventSubtype;
    public String ageGroup; // New demographic dimension
    public String gender;   // New demographic dimension

    // Metrics
    public long eventCount;          // Count of events in this group/window
    public double totalScoreImpact;    // Sum of score changes for events in this group/window
    public double avgScoreImpact;      // Calculated: totalScoreImpact / eventCount
    public long distinctCitizenCount; // Count of distinct citizens involved in events in this group/window

    public EventAggregate() {
    }

    /**
     * Helper method to get the window end time as a java.sql.Timestamp.
     */
    public Timestamp getWindowEndDateTime() {
        return (this.windowEnd != null) ? new Timestamp(this.windowEnd) : null;
    }

    @Override
    public String toString() {
        return "EventAggregate{" +
               "windowStart=" + windowStart +
               ", windowEnd=" + windowEnd +
               ", region='" + region + "'" +
               ", city='" + city + "'" +
               ", district='" + district + "'" +
               ", eventType='" + eventType + "'" +
               ", eventSubtype='" + eventSubtype + "'" +
               ", ageGroup='" + ageGroup + "'" +
               ", gender='" + gender + "'" +
               ", eventCount=" + eventCount +
               ", totalScoreImpact=" + totalScoreImpact +
               ", avgScoreImpact=" + avgScoreImpact +
               ", distinctCitizenCount=" + distinctCitizenCount +
               "}";
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        EventAggregate that = (EventAggregate) o;
        return eventCount == that.eventCount &&
               Double.compare(that.totalScoreImpact, totalScoreImpact) == 0 &&
               Double.compare(that.avgScoreImpact, avgScoreImpact) == 0 &&
               distinctCitizenCount == that.distinctCitizenCount &&
               Objects.equals(windowStart, that.windowStart) &&
               Objects.equals(windowEnd, that.windowEnd) &&
               Objects.equals(region, that.region) &&
               Objects.equals(city, that.city) &&
               Objects.equals(district, that.district) &&
               Objects.equals(eventType, that.eventType) &&
               Objects.equals(eventSubtype, that.eventSubtype) &&
               Objects.equals(ageGroup, that.ageGroup) &&
               Objects.equals(gender, that.gender);
    }

    @Override
    public int hashCode() {
        return Objects.hash(windowStart, windowEnd, region, city, district, eventType, eventSubtype, 
                          ageGroup, gender, eventCount, totalScoreImpact, avgScoreImpact, distinctCitizenCount);
    }
} 