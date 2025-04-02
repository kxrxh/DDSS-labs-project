package com.github.kxrxh.model;

import java.sql.Timestamp;
import java.util.Objects;

/**
 * Represents calculated score velocity metrics over different time windows
 * for a given geographic location, intended for the 'score_velocity_metrics' ClickHouse table.
 */
public class ScoreVelocityMetrics {

    // Timestamp representing the end of the calculation period (set during calculation)
    public Long calculationTimestamp;

    // Dimensions
    public String region;
    public String city;
    public String district;

    // Metrics
    public double scoreChange1h;    // Average score change over the past 1 hour
    public double scoreChange24h;   // Average score change over the past 24 hours
    public long positiveEvents1h;   // Count of positive score changes in the past 1 hour
    public long negativeEvents1h;   // Count of negative score changes in the past 1 hour

    public ScoreVelocityMetrics() {
    }

    /**
     * Helper method to get the calculation timestamp as a java.sql.Timestamp,
     * suitable for ClickHouse DateTime columns.
     * Returns null if calculationTimestamp is not set.
     */
    public Timestamp getCalculationEndDateTime() {
        return (this.calculationTimestamp != null) ? new Timestamp(this.calculationTimestamp) : null;
    }

    // Getters (optional, fields are public)

    @Override
    public String toString() {
        return "ScoreVelocityMetrics{" +
               "calculationTimestamp=" + calculationTimestamp +
               ", region='" + region + "'" +
               ", city='" + city + "'" +
               ", district='" + district + "'" +
               ", scoreChange1h=" + scoreChange1h +
               ", scoreChange24h=" + scoreChange24h +
               ", positiveEvents1h=" + positiveEvents1h +
               ", negativeEvents1h=" + negativeEvents1h +
               "}";
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        ScoreVelocityMetrics that = (ScoreVelocityMetrics) o;
        return Double.compare(that.scoreChange1h, scoreChange1h) == 0 &&
               Double.compare(that.scoreChange24h, scoreChange24h) == 0 &&
               positiveEvents1h == that.positiveEvents1h &&
               negativeEvents1h == that.negativeEvents1h &&
               Objects.equals(calculationTimestamp, that.calculationTimestamp) &&
               Objects.equals(region, that.region) &&
               Objects.equals(city, that.city) &&
               Objects.equals(district, that.district);
    }

    @Override
    public int hashCode() {
        return Objects.hash(calculationTimestamp, region, city, district, scoreChange1h, scoreChange24h, positiveEvents1h, negativeEvents1h);
    }
} 