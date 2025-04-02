package com.github.kxrxh.model;

import java.util.Objects;

/**
 * Represents the final state of a citizen's score after processing an event.
 * Includes the final score, level, change details, and original event details for aggregation.
 */
public class ScoreSnapshot {

    // Core scoring results
    public String citizenId;
    public double finalScore;
    public String level; // Calculated based on finalScore
    public double scoreChange; // The change caused by the specific rule/event

    // Event details for context and aggregation
    public Long eventTimestamp; // Timestamp from the original event
    public String sourceEventId; // ID of the event that caused this change
    public String triggeringRuleId; // ID of the rule that caused this change
    public String eventType;
    public String eventSubtype;
    public String region; // Nullable
    public String city;   // Nullable
    public String district; // Nullable

    // --- Added Demographic Fields --- 
    public String ageGroup; // Nullable
    public String gender;   // Nullable
    // --- End Added Fields ---

    public ScoreSnapshot() {}

    // Updated constructor
    public ScoreSnapshot(String citizenId, double finalScore, String level, double scoreChange, 
                         Long eventTimestamp, String sourceEventId, String triggeringRuleId,
                         String eventType, String eventSubtype, String region, String city, String district,
                         String ageGroup, String gender) {
        this.citizenId = citizenId;
        this.finalScore = finalScore;
        this.level = level;
        this.scoreChange = scoreChange;
        this.eventTimestamp = eventTimestamp;
        this.sourceEventId = sourceEventId;
        this.triggeringRuleId = triggeringRuleId;
        this.eventType = eventType;
        this.eventSubtype = eventSubtype;
        this.region = region;
        this.city = city;
        this.district = district;
        this.ageGroup = ageGroup;
        this.gender = gender;
    }

    // Getters or public fields

    // Updated toString, equals, hashCode
    @Override
    public String toString() {
        return "ScoreSnapshot{" +
               "citizenId='" + citizenId + "'" +
               ", finalScore=" + finalScore +
               ", level='" + level + "'" +
               ", scoreChange=" + scoreChange +
               ", eventTimestamp=" + eventTimestamp +
               ", sourceEventId='" + sourceEventId + "'" +
               ", triggeringRuleId='" + triggeringRuleId + "'" +
               ", eventType='" + eventType + "'" +
               ", eventSubtype='" + eventSubtype + "'" +
               ", region='" + region + "'" +
               ", city='" + city + "'" +
               ", district='" + district + "'" +
               ", ageGroup='" + ageGroup + "'" +
               ", gender='" + gender + "'" +
               "}";
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        ScoreSnapshot that = (ScoreSnapshot) o;
        return Double.compare(that.finalScore, finalScore) == 0 && 
               Double.compare(that.scoreChange, scoreChange) == 0 && 
               Objects.equals(eventTimestamp, that.eventTimestamp) &&
               Objects.equals(citizenId, that.citizenId) && 
               Objects.equals(level, that.level) && 
               Objects.equals(sourceEventId, that.sourceEventId) && 
               Objects.equals(triggeringRuleId, that.triggeringRuleId) && 
               Objects.equals(eventType, that.eventType) && 
               Objects.equals(eventSubtype, that.eventSubtype) && 
               Objects.equals(region, that.region) && 
               Objects.equals(city, that.city) && 
               Objects.equals(district, that.district) &&
               Objects.equals(ageGroup, that.ageGroup) &&
               Objects.equals(gender, that.gender);
    }

    @Override
    public int hashCode() {
        return Objects.hash(citizenId, finalScore, level, scoreChange, eventTimestamp, sourceEventId, 
                          triggeringRuleId, eventType, eventSubtype, region, city, district,
                          ageGroup, gender);
    }
} 