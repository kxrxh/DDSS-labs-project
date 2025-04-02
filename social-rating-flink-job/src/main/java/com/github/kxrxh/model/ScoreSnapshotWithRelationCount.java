package com.github.kxrxh.model;

import java.util.Objects;

/**
 * Extends ScoreSnapshot to include the count of a citizen's relationships.
 */
public class ScoreSnapshotWithRelationCount extends ScoreSnapshot {

    public int relationCount; // Count of relations (e.g., from relatedEffectOn)

    // Default constructor
    public ScoreSnapshotWithRelationCount() {
        super();
    }

    // Constructor to initialize from ScoreSnapshot and add relationCount
    public ScoreSnapshotWithRelationCount(ScoreSnapshot snapshot, int relationCount) {
        // Copy fields from the original snapshot
        super(
            snapshot.citizenId,
            snapshot.finalScore,
            snapshot.level,
            snapshot.scoreChange,
            snapshot.eventTimestamp,
            snapshot.sourceEventId,
            snapshot.triggeringRuleId,
            snapshot.eventType,
            snapshot.eventSubtype,
            snapshot.region,
            snapshot.city,
            snapshot.district,
            snapshot.ageGroup,
            snapshot.gender
        );
        this.relationCount = relationCount;
    }

    // Getters (optional, fields are public)
    public int getRelationCount() {
        return relationCount;
    }

    @Override
    public String toString() {
        return "ScoreSnapshotWithRelationCount{" +
               "relationCount=" + relationCount +
               ", " + super.toString().substring(super.toString().indexOf('{') + 1); // Append parent toString without the class name
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        if (!super.equals(o)) return false; // Check parent equality
        ScoreSnapshotWithRelationCount that = (ScoreSnapshotWithRelationCount) o;
        return relationCount == that.relationCount;
    }

    @Override
    public int hashCode() {
        return Objects.hash(super.hashCode(), relationCount);
    }
} 