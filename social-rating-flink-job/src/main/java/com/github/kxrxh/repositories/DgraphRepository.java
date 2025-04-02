package com.github.kxrxh.repositories;

import io.dgraph.DgraphClient;
import io.dgraph.DgraphGrpc;
import io.dgraph.DgraphGrpc.DgraphStub;
import io.dgraph.Transaction;
import com.github.kxrxh.model.ScoreSnapshot; // Import POJO
import com.google.protobuf.ByteString; // Needed for mutations
import io.dgraph.DgraphProto.Mutation; // Needed for Mutation
import io.dgraph.DgraphProto.Request; // Needed for Request
import io.dgraph.DgraphProto.Response; // Needed for Response
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import com.google.gson.Gson;
import com.google.gson.JsonArray;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import com.google.gson.reflect.TypeToken;
import java.lang.reflect.Type;
import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.stream.Collectors;

import java.time.Instant;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.util.concurrent.TimeUnit;
import java.util.HashMap; // Import HashMap

/**
 * Repository for interacting with Dgraph.
 * Handles saving score change history.
 */
public class DgraphRepository implements AutoCloseable {

    private static final Logger LOG = LoggerFactory.getLogger(DgraphRepository.class);

    // Configuration (Consider externalizing this)
    private static final String DGRAPH_URI = "localhost:9080"; // Example Dgraph Alpha address

    private final ManagedChannel channel;
    private final DgraphClient dgraphClient;

    public DgraphRepository() {
        this(DGRAPH_URI);
    }

    public DgraphRepository(String dgraphUri) {
        ManagedChannel tempChannel = null; // Initialize temporary variable
        DgraphClient tempClient = null;
        try {
            // Note: For production, consider secure connections (TLS) instead of
            // usePlaintext()
            tempChannel = ManagedChannelBuilder.forTarget(dgraphUri).usePlaintext().build();
            DgraphStub stub = DgraphGrpc.newStub(tempChannel);
            tempClient = new DgraphClient(stub);
            LOG.info("Dgraph client initialized for target: {}", dgraphUri);
        } catch (Exception e) {
            LOG.error("Failed to initialize Dgraph client", e);
            // Clean up channel if initialization fails partially
            if (tempChannel != null && !tempChannel.isShutdown()) {
                tempChannel.shutdownNow();
            }
            throw new RuntimeException("Failed to initialize Dgraph client", e);
        } finally {
             // Assign final fields outside the try-catch
             this.channel = tempChannel;
             this.dgraphClient = tempClient;
        }
    }

    /**
     * Saves a score change event derived from ScoreSnapshot to Dgraph.
     * Uses Upsert block to find citizen by citizenId and add to scoreHistory.
     *
     * @param snapshot The snapshot containing details of the score change.
     */
    public void saveScoreChange(ScoreSnapshot snapshot) {
        if (snapshot == null || snapshot.citizenId == null) {
            LOG.warn("Cannot save score change to Dgraph: snapshot or citizenId is null.");
            return;
        }
        LOG.debug("Saving score change to Dgraph for citizen ID: {}", snapshot.citizenId);

        // Use transaction for Upsert
        Transaction txn = dgraphClient.newTransaction();
        try {
            // 1. Query to find Citizen UID by citizenId using a variable
            String query = String.format("query citizenUid($citizenId: string) {\n" +
                                         "  var(func: eq(citizenId, $citizenId)) {\n" +
                                         "    citizen as uid\n" +
                                         "  }\n" +
                                         "}");

            // 2. Mutation to create ScoreChange and link it to Citizen
            String timestampISO = Instant.ofEpochMilli(snapshot.eventTimestamp)
                                         .atOffset(ZoneOffset.UTC)
                                         .format(DateTimeFormatter.ISO_OFFSET_DATE_TIME);

            // Use variable 'citizen' from the query part
            String nquadsMutation = String.format(
               "uid(citizen) <dgraph.type> \"Citizen\" . \n" +
               "uid(citizen) <citizenId> \"%s\" . \n" +
               "_:change <dgraph.type> \"ScoreChange\" . \n" +
               "_:change <eventId> \"%s\" . \n" +
               "_:change <ruleId> \"%s\" . \n" +
               "_:change <changePoints> \"%f\" . \n" +
               "_:change <changeTimestamp> \"%s\"^^<xs:dateTime> . \n" +
               "_:change <finalScore> \"%f\" . \n" +
               "uid(citizen) <scoreHistory> _:change .",
               snapshot.citizenId, // For Upsert predicate
               snapshot.sourceEventId != null ? snapshot.sourceEventId : "unknown", // Handle potential nulls
               snapshot.triggeringRuleId != null ? snapshot.triggeringRuleId : "unknown", // Handle potential nulls
               snapshot.scoreChange,  // This is the change, not final score
               timestampISO,
               snapshot.finalScore // Save the final score at this point
            );

            Mutation mu = Mutation.newBuilder()
                    .setSetNquads(ByteString.copyFromUtf8(nquadsMutation))
                    .build();

            // 3. Create and execute the Upsert request
            Request request = Request.newBuilder()
                    .setQuery(query)
                    .putVars("$citizenId", snapshot.citizenId) // Pass variable to query
                    .addMutations(mu)
                    .setCommitNow(true) // Execute and commit
                    .build();

            Response response = txn.doRequest(request);
            // Use processing nanos and convert to ms
            long latencyNanos = response.getLatency().getProcessingNs();
            LOG.debug("Dgraph upsert response latency: {} ms", TimeUnit.NANOSECONDS.toMillis(latencyNanos));
            // Optional: Check response context for errors or UIDs
            // LOG.debug("Dgraph response uids: {}", response.getUidsMap());

        } catch (StatusRuntimeException e) {
            LOG.error("Dgraph transaction failed for citizen {}: Status {}", snapshot.citizenId, e.getStatus(), e);
            txn.discard(); // Discard transaction on gRPC error
            // Consider retry logic or DLQ
        } catch (Exception e) {
             LOG.error("Unexpected error during Dgraph transaction for citizen {}: {}", snapshot.citizenId, e.getMessage(), e);
             txn.discard(); // Discard on other errors
        } 
        // Transaction is either committed (CommitNow=true) or discarded
    }

    /**
     * Finds citizens related to the given citizen within a max distance (currently supports distance 1).
     *
     * @param citizenId ID of the starting citizen.
     * @param relationTypes List of relationship predicates (e.g., "family", "colleagues").
     * @param maxDistance Max depth for the search (currently only 1 is effectively supported).
     * @return A Map where keys are the UIDs of related citizens and values are their corresponding external citizenIds.
     *         Returns an empty map on error or if none found.
     */
    public Map<String, String> findRelatedCitizens(String citizenId, List<String> relationTypes, int maxDistance) { // Changed return type
         if (citizenId == null || relationTypes == null || relationTypes.isEmpty() || maxDistance < 1) {
             LOG.warn("Invalid input for findRelatedCitizens: citizenId={}, relationTypes={}, maxDistance={}",
                      citizenId, relationTypes, maxDistance);
             return Collections.emptyMap(); // Changed return type
         }
          if (maxDistance > 1) {
             // TODO: Implement recursive query or var blocks for maxDistance > 1
             LOG.warn("findRelatedCitizens currently only supports maxDistance=1. Proceeding with distance 1.");
         }

         LOG.debug("Finding related citizens (distance=1) for {} via relations {}", citizenId, relationTypes);

        // Build the query projection part dynamically
        StringBuilder projection = new StringBuilder();
        projection.append("startUid: uid\n"); // Alias for clarity
        for (String relation : relationTypes) {
            // Add filter to exclude self and ensure type. Fetch uid and citizenId.
            projection.append(String.format("  %s @filter(type(Citizen) AND NOT uid(startUid)) { relatedUid: uid, citizenId }\n", relation)); // Added citizenId
        }

        // Build the full query
        String query = String.format(
            "query findRelated($citizenId: string) {\n" +
            "  var(func: eq(citizenId, $citizenId)) { startUid as uid } \n" +
            "  startNode(func: uid(startUid)) {\n" +
            "    %s" +
            "  }\n" +
            "}", projection.toString()
        );

        Transaction txn = dgraphClient.newReadOnlyTransaction();
        Map<String, String> relatedCitizenInfo = new HashMap<>(); // Changed to Map
        try {
            Request request = Request.newBuilder()
                .setQuery(query)
                .putVars("$citizenId", citizenId)
                .build();

            Response response = txn.queryWithVars(request.getQuery(), request.getVarsMap()); // Use queryWithVars for read-only
            String jsonResponse = response.getJson().toStringUtf8();
            LOG.trace("Dgraph findRelatedCitizens response JSON: {}", jsonResponse);

            // Parse the JSON response using Gson
            JsonObject result = JsonParser.parseString(jsonResponse).getAsJsonObject();
            JsonArray startNodes = result.getAsJsonArray("startNode");

            if (startNodes == null || startNodes.size() == 0) {
                LOG.warn("Could not find start node in Dgraph for citizenId: {}", citizenId);
                return Collections.emptyMap(); // Changed return type
            }

            JsonObject startNodeData = startNodes.get(0).getAsJsonObject();
            // String startUid = startNodeData.has("startUid") ? startNodeData.get("startUid").getAsString() : null;

            // Iterate through the requested relation types found in the response
            for (String relation : relationTypes) {
                if (startNodeData.has(relation)) {
                    JsonArray relatedArray = startNodeData.getAsJsonArray(relation);
                    if (relatedArray != null) {
                        for (JsonElement relatedElement : relatedArray) {
                            JsonObject relatedObj = relatedElement.getAsJsonObject();
                            String relatedUid = null;
                            String relatedCitizenId = null;
                            
                            if (relatedObj.has("relatedUid")) {
                                relatedUid = relatedObj.get("relatedUid").getAsString();
                            }
                            if (relatedObj.has("citizenId")) { // Check for citizenId
                                relatedCitizenId = relatedObj.get("citizenId").getAsString();
                            }

                            // Add to map only if both UID and citizenId are present
                            if (relatedUid != null && !relatedUid.isEmpty() && relatedCitizenId != null && !relatedCitizenId.isEmpty()) {
                                relatedCitizenInfo.put(relatedUid, relatedCitizenId);
                            } else {
                                LOG.warn("Found related node via '{}' but missing UID ({}) or citizenId ({}). Skipping.", 
                                         relation, relatedUid, relatedCitizenId);
                            }
                        }
                    }
                }
            }

            // Exclude the start node UID itself, although the query filter should handle this
            // if (startUid != null) { relatedUids.remove(startUid); } 

            LOG.debug("Found {} unique related citizens (UID -> citizenId) for {}: {}", relatedCitizenInfo.size(), citizenId, relatedCitizenInfo);
            return relatedCitizenInfo; // Return the map

        } catch (Exception e) {
            LOG.error("Error querying/parsing related citizens for {}: {}", citizenId, e.getMessage(), e);
            return Collections.emptyMap(); // Return empty map on error
        } finally {
            txn.discard(); 
        }
    }

    /**
     * Gets the count of relationships (outgoing edges of type 'relatedEffectOn') 
     * for a citizen identified by their external ID.
     *
     * @param externalId The external ID (e.g., citizenId) of the citizen.
     * @return The number of relationships, or 0 if the citizen is not found or has no relationships.
     */
    public int getCitizenRelationCountByExternalId(String externalId) {
        if (externalId == null || externalId.isEmpty()) {
            LOG.warn("Cannot get relation count: externalId is null or empty.");
            return 0;
        }
        LOG.debug("Getting relationship count for citizen externalId: {}", externalId);

        // Query to find the citizen and count the 'relatedEffectOn' edges
        String query = String.format(
            "query relationCount($id: string) {\n" +
            "  citizen(func: eq(citizenId, $id)) {\n" +
            "    count(relatedEffectOn) \n" +
            "  }\n" +
            "}");

        Transaction txn = dgraphClient.newReadOnlyTransaction();
        int count = 0;
        try {
            Request request = Request.newBuilder()
                .setQuery(query)
                .putVars("$id", externalId)
                .build();

            Response response = txn.queryWithVars(request.getQuery(), request.getVarsMap());
            String jsonResponse = response.getJson().toStringUtf8();
            LOG.trace("Dgraph getCitizenRelationCount response JSON: {}", jsonResponse);

            // Parse the JSON response
            JsonObject result = JsonParser.parseString(jsonResponse).getAsJsonObject();
            JsonArray citizenArray = result.getAsJsonArray("citizen");

            if (citizenArray != null && citizenArray.size() > 0) {
                JsonObject citizenData = citizenArray.get(0).getAsJsonObject();
                String countPredicate = "count(relatedEffectOn)";
                if (citizenData.has(countPredicate)) {
                    count = citizenData.get(countPredicate).getAsInt();
                }
            }
             LOG.debug("Found {} relationships for citizen externalId {}", count, externalId);

        } catch (Exception e) {
            LOG.error("Error querying/parsing relationship count for {}: {}", externalId, e.getMessage(), e);
            // Return 0 on error
        } finally {
            txn.discard();
        }
        return count;
    }

    @Override
    public void close() {
        if (channel != null && !channel.isShutdown()) {
            try {
                LOG.info("Shutting down Dgraph channel...");
                channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
                LOG.info("Dgraph channel shut down.");
            } catch (InterruptedException e) {
                LOG.warn("Dgraph channel shutdown interrupted.", e);
                channel.shutdownNow(); // Force shutdown
                Thread.currentThread().interrupt();
            } catch (Exception e) {
                LOG.error("Error closing Dgraph channel", e);
                channel.shutdownNow(); // Force shutdown
            }
        }
    }
}