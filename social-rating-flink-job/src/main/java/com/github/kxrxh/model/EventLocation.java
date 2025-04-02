package com.github.kxrxh.model;

import java.util.Objects;

/**
 * Represents the geographic location associated with an event.
 * Corresponds to the EventLocation record in the Avro schema.
 * NOTE: Flink requires public fields or getter/setters for POJO detection.
 */
public class EventLocation {

    public Double latitude;
    public Double longitude;
    public String region; // Nullable
    public String city;   // Nullable
    public String district; // Nullable

    // Flink requires a public no-argument constructor for POJO types
    public EventLocation() {}

    // Optional: Constructor for convenience
    public EventLocation(Double latitude, Double longitude, String region, String city, String district) {
        this.latitude = latitude;
        this.longitude = longitude;
        this.region = region;
        this.city = city;
        this.district = district;
    }

    // Optional: Getters and Setters (or public fields)
    // Flink can work with public fields directly

    // Optional: toString, equals, hashCode for easier debugging/testing
    @Override
    public String toString() {
        return "EventLocation{" +
                "latitude=" + latitude +
                ", longitude=" + longitude +
                ", region='" + region + '\'' +
                ", city='" + city + '\'' +
                ", district='" + district + '\'' +
                '}';
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        EventLocation that = (EventLocation) o;
        return Objects.equals(latitude, that.latitude) && Objects.equals(longitude, that.longitude) && Objects.equals(region, that.region) && Objects.equals(city, that.city) && Objects.equals(district, that.district);
    }

    @Override
    public int hashCode() {
        return Objects.hash(latitude, longitude, region, city, district);
    }
} 