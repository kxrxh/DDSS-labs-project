package com.github.kxrxh.model;

import org.bson.Document;

import java.util.Objects;

/**
 * POJO representing demographic profile data fetched from MongoDB.
 * Used for enriching CitizenEvent.
 */
public class CitizenProfile {

    public String citizenId;
    public String ageGroup; // e.g., "18-24", "25-34", etc. (fetched from DB)
    public String gender;   // e.g., "M", "F" (fetched from DB)

    // Default constructor
    public CitizenProfile() {
    }

    // Constructor from MongoDB Document
    public CitizenProfile(Document doc) {
        if (doc != null) {
            this.citizenId = doc.getString("_id"); // Assuming citizenId is the MongoDB _id
            this.ageGroup = doc.getString("ageGroup"); // Assuming field name is ageGroup
            this.gender = doc.getString("gender");   // Assuming field name is gender
        }
    }

    // Getters (optional, fields are public)

    @Override
    public String toString() {
        return "CitizenProfile{" +
               "citizenId='" + citizenId + "'" +
               ", ageGroup='" + ageGroup + "'" +
               ", gender='" + gender + "'" +
               "}";
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        CitizenProfile that = (CitizenProfile) o;
        return Objects.equals(citizenId, that.citizenId) &&
               Objects.equals(ageGroup, that.ageGroup) &&
               Objects.equals(gender, that.gender);
    }

    @Override
    public int hashCode() {
        return Objects.hash(citizenId, ageGroup, gender);
    }
} 