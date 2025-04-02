package com.github.kxrxh.model;

import java.util.Map;
import java.util.Objects;

/**
 * Represents a citizen-related event consumed from Redpanda.
 * Corresponds to the CitizenEvent record in the Avro schema.
 * NOTE: Flink requires public fields or getter/setters for POJO detection.
 */
public class CitizenEvent {

    public String eventId;
    public String citizenId;
    public Long timestamp; // Epoch milliseconds
    public String eventType;
    public String eventSubtype; // Nullable
    public String sourceSystem;
    public EventLocation location; // Nullable, nested type
    public Map<String, String> payload;
    public Integer version;

    // --- Added Demographic Fields (to be populated by enrichment) ---
    public String ageGroup; // e.g., "18-24", "25-34", etc.
    public String gender;   // e.g., "Male", "Female"
    // Note: These are added directly. Alternatively, use a nested 'Demographics' POJO.
    // --- End Added Fields ---

    // Flink requires a public no-argument constructor for POJO types
    public CitizenEvent() {}

    // Optional: Constructor for convenience
    public CitizenEvent(String eventId, String citizenId, Long timestamp, String eventType, String eventSubtype, String sourceSystem, EventLocation location, Map<String, String> payload, Integer version) {
        this.eventId = eventId;
        this.citizenId = citizenId;
        this.timestamp = timestamp;
        this.eventType = eventType;
        this.eventSubtype = eventSubtype;
        this.sourceSystem = sourceSystem;
        this.location = location;
        this.payload = payload;
        this.version = version != null ? version : 1; // Handle default value
    }

    // Optional: Getters and Setters (or public fields)

    // Optional: toString, equals, hashCode
    @Override
    public String toString() {
        return "CitizenEvent{" +
                "eventId='" + eventId + "'" +
                ", citizenId='" + citizenId + "'" +
                ", timestamp=" + timestamp +
                ", eventType='" + eventType + "'" +
                ", eventSubtype='" + eventSubtype + "'" +
                ", sourceSystem='" + sourceSystem + "'" +
                ", location=" + location +
                ", payload=" + payload +
                ", version=" + version +
                "}";
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        CitizenEvent that = (CitizenEvent) o;
        return Objects.equals(eventId, that.eventId) && Objects.equals(citizenId, that.citizenId) && Objects.equals(timestamp, that.timestamp) && Objects.equals(eventType, that.eventType) && Objects.equals(eventSubtype, that.eventSubtype) && Objects.equals(sourceSystem, that.sourceSystem) && Objects.equals(location, that.location) && Objects.equals(payload, that.payload) && Objects.equals(version, that.version);
    }

    @Override
    public int hashCode() {
        return Objects.hash(eventId, citizenId, timestamp, eventType, eventSubtype, sourceSystem, location, payload, version);
    }
} 