package com.github.kxrxh.model;

import java.util.Objects;

/**
 * Represents the calculated score change for a specific event and citizen.
 * This is the output of the core processing logic.
 */
public class ScoreUpdate {

    public String citizenId;
    public double scoreChange;
    public long eventTimestamp; // Timestamp from the original event
    public String triggeringRuleId; // ID of the rule that caused this change
    public String sourceEventId; // ID of the event that caused this change
    // You might want to include the original event for context in sinks
    // public CitizenEvent originalEvent;

    public ScoreUpdate() {}

    public ScoreUpdate(String citizenId, double scoreChange, long eventTimestamp, String triggeringRuleId, String sourceEventId) {
        this.citizenId = citizenId;
        this.scoreChange = scoreChange;
        this.eventTimestamp = eventTimestamp;
        this.triggeringRuleId = triggeringRuleId;
        this.sourceEventId = sourceEventId;
    }

    // Getters/Setters or public fields

    @Override
    public String toString() {
        return "ScoreUpdate{" +
                "citizenId='" + citizenId + "'" +
                ", scoreChange=" + scoreChange +
                ", eventTimestamp=" + eventTimestamp +
                ", triggeringRuleId='" + triggeringRuleId + "'" +
                ", sourceEventId='" + sourceEventId + "'" +
                "}";
    }

     @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        ScoreUpdate that = (ScoreUpdate) o;
        return Double.compare(that.scoreChange, scoreChange) == 0 &&
               eventTimestamp == that.eventTimestamp &&
               Objects.equals(citizenId, that.citizenId) &&
               Objects.equals(triggeringRuleId, that.triggeringRuleId) &&
               Objects.equals(sourceEventId, that.sourceEventId);
    }

    @Override
    public int hashCode() {
        return Objects.hash(citizenId, scoreChange, eventTimestamp, triggeringRuleId, sourceEventId);
    }
} 