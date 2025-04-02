package com.github.kxrxh.model;

import java.util.Objects;

public class ScoreChangeDetails {
    public Double points;
    public String multiplierExpression; // Keep as String for now, evaluation separate
    public Integer durationMonths;

    public ScoreChangeDetails() {}

    // Getters/Setters or public fields

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        ScoreChangeDetails that = (ScoreChangeDetails) o;
        return Objects.equals(points, that.points) && Objects.equals(multiplierExpression, that.multiplierExpression) && Objects.equals(durationMonths, that.durationMonths);
    }

    @Override
    public int hashCode() {
        return Objects.hash(points, multiplierExpression, durationMonths);
    }

     @Override
    public String toString() {
        return "ScoreChangeDetails{" +
               "points=" + points +
               ", multiplierExpression='" + multiplierExpression + "'" +
               ", durationMonths=" + durationMonths +
               "}";
    }
} 