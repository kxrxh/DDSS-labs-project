package com.github.kxrxh.model;

import java.util.Date; // Or java.time.Instant
import java.util.Map;
import java.util.Objects;

/**
 * Represents a scoring rule fetched from MongoDB.
 * Updated based on schemas/mongodb.txt.
 */
public class ScoringRule {

    // Note: MongoDB _id needs mapping if it's an ObjectId. Use a String representation.
    // Consider adding a @BsonId annotation or similar if using a MongoDB mapping framework.
    public String id; // Mapped from _id or rule_id

    public String rule_id; // Unique human-readable ID
    public String description;
    public String eventType; // Condition directly mapped
    public String eventSubtype; // Condition directly mapped
    public Map<String, Object> conditions; // Flexible conditions (e.g., payload.severity, sourceSystem)
    public ScoreChangeDetails scoreChange; // Nested object
    public RelatedEffect effectOnRelated; // Nested object, nullable
    public Boolean isActive;
    public Integer priority;
    public Date createdAt; // Or Instant
    public Date updatedAt; // Or Instant

    public ScoringRule() {}

    // Getters/Setters or public fields

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        ScoringRule that = (ScoringRule) o;
        return Objects.equals(id, that.id) && Objects.equals(rule_id, that.rule_id) && Objects.equals(description, that.description) && Objects.equals(eventType, that.eventType) && Objects.equals(eventSubtype, that.eventSubtype) && Objects.equals(conditions, that.conditions) && Objects.equals(scoreChange, that.scoreChange) && Objects.equals(effectOnRelated, that.effectOnRelated) && Objects.equals(isActive, that.isActive) && Objects.equals(priority, that.priority) && Objects.equals(createdAt, that.createdAt) && Objects.equals(updatedAt, that.updatedAt);
    }

    @Override
    public int hashCode() {
        return Objects.hash(id, rule_id, description, eventType, eventSubtype, conditions, scoreChange, effectOnRelated, isActive, priority, createdAt, updatedAt);
    }

    @Override
    public String toString() {
        return "ScoringRule{" +
                "id='" + id + "'" +
                ", rule_id='" + rule_id + "'" +
                ", description='" + description + "'" +
                ", eventType='" + eventType + "'" +
                ", eventSubtype='" + eventSubtype + "'" +
                ", conditions=" + conditions +
                ", scoreChange=" + scoreChange +
                ", effectOnRelated=" + effectOnRelated +
                ", isActive=" + isActive +
                ", priority=" + priority +
                ", createdAt=" + createdAt +
                ", updatedAt=" + updatedAt +
                "}";
    }
} 