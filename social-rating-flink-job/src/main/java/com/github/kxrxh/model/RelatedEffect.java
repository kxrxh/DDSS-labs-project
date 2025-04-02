package com.github.kxrxh.model;

import java.util.List;
import java.util.Objects;

public class RelatedEffect {
    public List<String> relationType;
    public Double percentage;
    public Integer maxDistance;

    public RelatedEffect() {}

    // Getters/Setters or public fields

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        RelatedEffect that = (RelatedEffect) o;
        return Objects.equals(relationType, that.relationType) && Objects.equals(percentage, that.percentage) && Objects.equals(maxDistance, that.maxDistance);
    }

    @Override
    public int hashCode() {
        return Objects.hash(relationType, percentage, maxDistance);
    }

     @Override
    public String toString() {
        return "RelatedEffect{" +
               "relationType=" + relationType +
               ", percentage=" + percentage +
               ", maxDistance=" + maxDistance +
               "}";
    }
} 