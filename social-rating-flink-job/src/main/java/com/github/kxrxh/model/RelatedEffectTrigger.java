package com.github.kxrxh.model;

import java.util.Objects;

/**
 * POJO to trigger the asynchronous processing of related citizen effects.
 */
public class RelatedEffectTrigger {

    public String originalCitizenId;
    public double originalScoreChange; // The change applied to the original citizen
    public String originalSourceEventId;
    public RelatedEffect effectDetails;
    public String triggeringRuleId;

    public RelatedEffectTrigger() {}

    public RelatedEffectTrigger(String originalCitizenId, double originalScoreChange, String originalSourceEventId, RelatedEffect effectDetails, String triggeringRuleId) {
        this.originalCitizenId = originalCitizenId;
        this.originalScoreChange = originalScoreChange;
        this.originalSourceEventId = originalSourceEventId;
        this.effectDetails = effectDetails;
        this.triggeringRuleId = triggeringRuleId; // <-- Assign parameter
    }

    // Getters or public fields

    @Override
    public String toString() {
        return "RelatedEffectTrigger{" +
               "originalCitizenId='" + originalCitizenId + "'" +
               ", originalScoreChange=" + originalScoreChange +
               ", originalSourceEventId='" + originalSourceEventId + "'" +
               ", triggeringRuleId='" + triggeringRuleId + "'" + // <-- Added to toString
               ", effectDetails=" + effectDetails +
               "}";
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        RelatedEffectTrigger that = (RelatedEffectTrigger) o;
        return Double.compare(that.originalScoreChange, originalScoreChange) == 0 &&
               Objects.equals(originalCitizenId, that.originalCitizenId) &&
               Objects.equals(originalSourceEventId, that.originalSourceEventId) &&
               Objects.equals(triggeringRuleId, that.triggeringRuleId) && // <-- Added to equals
               Objects.equals(effectDetails, that.effectDetails);
    }

    @Override
    public int hashCode() {
        return Objects.hash(originalCitizenId, originalScoreChange, originalSourceEventId, effectDetails, triggeringRuleId); // <-- Added to hashCode
    }
} 