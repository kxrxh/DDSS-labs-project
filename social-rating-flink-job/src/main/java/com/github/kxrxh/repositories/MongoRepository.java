package com.github.kxrxh.repositories;

import com.github.kxrxh.model.RelatedEffect;
import com.github.kxrxh.model.ScoreChangeDetails;
import com.github.kxrxh.model.ScoringRule;
import com.github.kxrxh.model.CitizenProfile;
import com.mongodb.client.MongoClient;
import com.mongodb.client.MongoClients;
import com.mongodb.client.MongoCollection;
import com.mongodb.client.MongoCursor;
import com.mongodb.client.MongoDatabase;
import org.bson.Document; // Import MongoDB Document
import com.mongodb.client.model.Filters; // Import Filters for querying
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
// Add imports for Bson, UpdateOptions, Updates etc. when implementing methods

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * Repository for interacting with MongoDB.
 * Handles fetching scoring rules, citizen profiles, and updating citizen snapshots.
 */
public class MongoRepository implements AutoCloseable {

    private static final Logger LOG = LoggerFactory.getLogger(MongoRepository.class);

    // Configuration (Consider externalizing this)
    private static final String MONGO_URI = "mongodb://localhost:27017";
    private static final String MONGO_DB_NAME = "social_rating";
    private static final String RULES_COLLECTION = "scoring_rules";
    private static final String CITIZENS_COLLECTION = "citizens";

    private final MongoClient mongoClient;
    private final MongoDatabase database;

    public MongoRepository() {
        // Consider making configuration parameters injectable
        this(MONGO_URI, MONGO_DB_NAME);
    }

    public MongoRepository(String uri, String dbName) {
        try {
            this.mongoClient = MongoClients.create(uri);
            this.database = mongoClient.getDatabase(dbName);
            LOG.info("MongoDB client initialized for database: {}", dbName);
        } catch (Exception e) {
            LOG.error("Failed to initialize MongoDB client", e);
            throw new RuntimeException("Failed to initialize MongoDB client", e);
        }
    }

    /**
     * Fetches active scoring rules from MongoDB.
     *
     * @return A List of active ScoringRule objects.
     */
    public List<ScoringRule> getActiveScoringRules() {
        LOG.debug("Fetching active scoring rules from collection: {}", RULES_COLLECTION);
        List<ScoringRule> rules = new ArrayList<>();
        MongoCollection<Document> collection = database.getCollection(RULES_COLLECTION);

        // Query for documents where isActive is true
        try (MongoCursor<Document> cursor = collection.find(Filters.eq("isActive", true)).iterator()) {
            while (cursor.hasNext()) {
                Document doc = cursor.next();
                try {
                    ScoringRule rule = mapDocumentToScoringRule(doc);
                    rules.add(rule);
                } catch (Exception e) {
                    // Log error mapping specific document but continue with others
                    LOG.error("Error mapping MongoDB document to ScoringRule: {}. Document: {}", e.getMessage(), doc.toJson(), e);
                }
            }
        }
        LOG.info("Fetched {} active scoring rules.", rules.size());
        return rules;
    }

    // Helper method to map BSON Document to ScoringRule POJO
    private ScoringRule mapDocumentToScoringRule(Document doc) {
        ScoringRule rule = new ScoringRule();

        // Handle potential ObjectId for _id, convert to String
        Object idObject = doc.get("_id");
        rule.id = (idObject != null) ? idObject.toString() : null;

        rule.rule_id = doc.getString("rule_id");
        rule.description = doc.getString("description");
        rule.eventType = doc.getString("eventType");
        rule.eventSubtype = doc.getString("eventSubtype");
        rule.isActive = doc.getBoolean("isActive", false); // Default to false if missing
        rule.priority = doc.getInteger("priority", 0); // Default priority
        rule.createdAt = doc.getDate("createdAt");
        rule.updatedAt = doc.getDate("updatedAt");

        // Map nested conditions (assuming it's a Document/Map)
        Object conditionsObj = doc.get("conditions");
        if (conditionsObj instanceof Map) {
             @SuppressWarnings("unchecked") // Be cautious with direct casting
             Map<String, Object> conditionsMap = (Map<String, Object>) conditionsObj;
             rule.conditions = conditionsMap;
        }

        // Map nested scoreChange
        Document scoreChangeDoc = doc.get("scoreChange", Document.class);
        if (scoreChangeDoc != null) {
            ScoreChangeDetails details = new ScoreChangeDetails();
            details.points = scoreChangeDoc.getDouble("points");
            details.multiplierExpression = scoreChangeDoc.getString("multiplierExpression");
            details.durationMonths = scoreChangeDoc.getInteger("durationMonths");
            rule.scoreChange = details;
        }

         // Map nested effectOnRelated
        Document effectDoc = doc.get("effectOnRelated", Document.class);
        if (effectDoc != null) {
            RelatedEffect effect = new RelatedEffect();
            // BSON arrays are often List<Object>, needs casting
            Object relationTypeObj = effectDoc.get("relationType");
             if(relationTypeObj instanceof List){ // Check if it's a list before casting
                 @SuppressWarnings("unchecked")
                 List<String> relationTypes = (List<String>) relationTypeObj; // Might need individual casting if list contains non-strings
                 effect.relationType = relationTypes;
             }
            effect.percentage = effectDoc.getDouble("percentage");
            effect.maxDistance = effectDoc.getInteger("maxDistance");
            rule.effectOnRelated = effect;
        }

        return rule;
    }

    /**
     * Fetches the demographic profile for a given citizen ID.
     *
     * @param citizenId The ID of the citizen.
     * @return A CitizenProfile object, or null if the citizen is not found or an error occurs.
     */
    public CitizenProfile getCitizenProfile(String citizenId) {
        if (citizenId == null || citizenId.isEmpty()) {
            LOG.warn("Cannot fetch citizen profile: citizenId is null or empty.");
            return null;
        }
        LOG.debug("Fetching citizen profile from MongoDB for ID: {}", citizenId);
        MongoCollection<Document> collection = database.getCollection(CITIZENS_COLLECTION);

        try {
            // Find the document where _id matches the citizenId
            Document doc = collection.find(Filters.eq("_id", citizenId)).first();

            if (doc != null) {
                LOG.debug("Found profile for citizen {}: {}", citizenId, doc.toJson());
                // Map the document to the POJO
                return new CitizenProfile(doc);
            } else {
                LOG.warn("No profile found in collection '{}' for citizen ID: {}", CITIZENS_COLLECTION, citizenId);
                return null;
            }
        } catch (Exception e) {
            LOG.error("Error fetching citizen profile for {}: {}", citizenId, e.getMessage(), e);
            return null; // Return null on error
        }
    }

    /**
     * Updates the score snapshot for a citizen in MongoDB.
     * Uses upsert to create the document if it doesn't exist.
     * TODO: Implement actual update logic using BSON updates.
     *
     * @param citizenId The ID of the citizen.
     * @param score     The new score snapshot.
     * @param level     The new calculated level.
     */
    public void updateCitizenSnapshot(String citizenId, double score, String level) {
        LOG.debug("Updating citizen snapshot in MongoDB for ID: {} with score: {}, level: {}", citizenId, score, level);
        // Example using Filters, Updates, UpdateOptions:
        /*
        MongoCollection<Document> citizenCollection = database.getCollection(CITIZENS_COLLECTION);
        Bson filter = Filters.eq("citizenId", citizenId); // Assuming citizenId field exists
        Bson update = Updates.combine(
            Updates.set("scoreSnapshot.score", score),
            Updates.set("scoreSnapshot.level", level),
            Updates.set("scoreSnapshot.lastUpdated", Instant.now()), // Use Instant
            Updates.setOnInsert("citizenId", citizenId) // Set citizenId only on insert
        );
        UpdateOptions options = new UpdateOptions().upsert(true);
        try {
            UpdateResult result = citizenCollection.updateOne(filter, update, options);
            LOG.debug("MongoDB update result: {}", result);
        } catch (Exception e) {
            LOG.error("Failed to update citizen snapshot for {}: {}", citizenId, e.getMessage(), e);
        }
        */
    }

    @Override
    public void close() {
        if (mongoClient != null) {
            try {
                mongoClient.close();
                LOG.info("MongoDB client closed.");
            } catch (Exception e) {
                LOG.error("Error closing MongoDB client", e);
            }
        }
    }

    // Helper method to access collection (used by MongoDbFirstSeenSinkFunction)
    public MongoCollection<Document> getCollection(String collectionName) {
        return database.getCollection(collectionName);
    }
} 