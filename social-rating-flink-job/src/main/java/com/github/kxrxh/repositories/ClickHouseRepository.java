package com.github.kxrxh.repositories;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.SQLException;

// Add necessary model imports
import com.github.kxrxh.model.EventAggregate;
import com.github.kxrxh.model.CreditScoreAggregate;
import com.github.kxrxh.model.SocialGraphAggregate;
import com.github.kxrxh.model.ScoreVelocityMetrics;

/**
 * Repository for interacting with ClickHouse using JDBC.
 * Handles saving aggregated data or archived events.
 */
public class ClickHouseRepository implements AutoCloseable {

    private static final Logger LOG = LoggerFactory.getLogger(ClickHouseRepository.class);

    // Driver class name
    private static final String CLICKHOUSE_DRIVER = "com.clickhouse.jdbc.ClickHouseDriver";

    private Connection connection;

    // Default constructor is not supported - requires a JDBC URL.
    @Deprecated // Mark as deprecated, prefer explicit URL constructor
    public ClickHouseRepository() {
         throw new UnsupportedOperationException("ClickHouseRepository requires a JDBC URL. Use the constructor ClickHouseRepository(String jdbcUrl).");
    }

    /**
     * Constructs a ClickHouseRepository and establishes a connection.
     * Reads credentials (user/password) from environment variables CLICKHOUSE_USER and CLICKHOUSE_PASSWORD.
     *
     * @param jdbcUrl The base ClickHouse JDBC URL (e.g., "jdbc:clickhouse://host:port/database")
     * @throws RuntimeException if the JDBC driver is not found, credentials are missing, or connection fails.
     */
    public ClickHouseRepository(String jdbcUrl) {
        try {
            // Ensure the JDBC driver is loaded
            Class.forName(CLICKHOUSE_DRIVER);

            // Get credentials from environment variables
            String user = System.getenv("CLICKHOUSE_USER");
            String password = System.getenv("CLICKHOUSE_PASSWORD");

            // Validate required environment variables
            if (jdbcUrl == null || jdbcUrl.trim().isEmpty()) {
                 LOG.error("ClickHouse JDBC URL provided is null or empty.");
                 throw new IllegalArgumentException("ClickHouse JDBC URL cannot be null or empty.");
            }
            if (user == null || user.trim().isEmpty()) {
                LOG.error("ClickHouse user environment variable CLICKHOUSE_USER is not set or empty.");
                throw new RuntimeException("ClickHouse user environment variable CLICKHOUSE_USER is not set or empty.");
            }
             if (password == null) {
                 // Decide how to handle missing password. ClickHouse can allow empty passwords.
                 // For a specific user like 'kxrxh', it's safer to assume a password is required.
                 // If an empty password *is* valid for this user, set password = "" here instead of throwing.
                 LOG.error("ClickHouse password environment variable CLICKHOUSE_PASSWORD is not set.");
                 throw new RuntimeException("ClickHouse password environment variable CLICKHOUSE_PASSWORD is not set.");
                 // Alternatively, allow empty password:
                 // LOG.warn("ClickHouse password environment variable CLICKHOUSE_PASSWORD is not set. Using empty password.");
                 // password = "";
             }

            LOG.info("Attempting to connect to ClickHouse at [{}] with user [{}]", jdbcUrl, user);

            // Establish connection using URL and credentials from Env Vars
            // Note: DriverManager requires URL, user, password arguments for authentication.
            this.connection = DriverManager.getConnection(jdbcUrl, user, password);

            // Disable auto-commit for potential batching later
            this.connection.setAutoCommit(false);

            LOG.info("ClickHouse connection established successfully.");

        } catch (ClassNotFoundException e) {
            LOG.error("ClickHouse JDBC Driver not found: {}", CLICKHOUSE_DRIVER, e);
            throw new RuntimeException("ClickHouse JDBC Driver not found", e);
        } catch (SQLException e) {
            // Log connection details carefully (avoid logging password if possible)
            LOG.error("Failed to establish ClickHouse connection to [{}] with user [{}]. Error: {}", jdbcUrl, user, e.getMessage(), e);
            throw new RuntimeException("Failed to establish ClickHouse connection", e);
        } catch (IllegalArgumentException | NullPointerException e) {
            // Catch validation errors for URL/User/Password
             LOG.error("Invalid configuration for ClickHouse connection: {}", e.getMessage(), e);
             throw new RuntimeException("Invalid configuration for ClickHouse connection", e);
        } catch (Exception e) { // Catch broader exceptions during init
            LOG.error("Unexpected error initializing ClickHouse repository", e);
            throw new RuntimeException("Unexpected error initializing ClickHouse repository", e);
        }
    }

     /**
      * Deprecated constructor. Use {@link #ClickHouseRepository(String)} instead.
      * @param jdbcUrl The JDBC URL.
      * @param properties Not used for credentials anymore.
      */
     @Deprecated
     public ClickHouseRepository(String jdbcUrl, java.util.Properties properties) { // Explicitly import Properties
         this(jdbcUrl); // Delegate to the main constructor
         LOG.warn("Constructor ClickHouseRepository(String, Properties) is deprecated. Properties argument is ignored for credentials.");
     }

    /**
     * Saves EventAggregate data to the events_aggregate table.
     *
     * @param aggregate The EventAggregate object to save.
     * @throws SQLException if a database access error occurs.
     */
    public void saveEventAggregate(EventAggregate aggregate) throws SQLException {
        String sql = "INSERT INTO events_aggregate " +
                     "(window_end, region, city, district, event_type, event_subtype, age_group, gender, " +
                     "event_count, total_score_impact, avg_score_impact, citizen_count) " +
                     "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"; // Matched columns

        LOG.debug("Saving EventAggregate to ClickHouse: {}", aggregate);
        try (PreparedStatement pstmt = connection.prepareStatement(sql)) {
            // Use helper method for timestamp conversion
            pstmt.setTimestamp(1, aggregate.getWindowEndDateTime()); 
            pstmt.setString(2, aggregate.region);
            pstmt.setString(3, aggregate.city);
            pstmt.setString(4, aggregate.district);
            pstmt.setString(5, aggregate.eventType);
            pstmt.setString(6, aggregate.eventSubtype);
            pstmt.setString(7, aggregate.ageGroup);
            pstmt.setString(8, aggregate.gender);
            pstmt.setLong(9, aggregate.eventCount);
            pstmt.setDouble(10, aggregate.totalScoreImpact);
            pstmt.setDouble(11, aggregate.avgScoreImpact);
            // Assuming citizen_count corresponds to distinctCitizenCount
            pstmt.setLong(12, aggregate.distinctCitizenCount); 

            pstmt.executeUpdate();
            LOG.info("Successfully queued EventAggregate for type {}/{} in region {} to ClickHouse.",
                     aggregate.eventType, aggregate.eventSubtype, aggregate.region);
        } catch (SQLException e) {
            LOG.error("Failed to queue EventAggregate for ClickHouse: {}. Error: {}", aggregate, e.getMessage(), e);
            throw e; // Re-throw exception
        }
    }

    /**
     * Saves CreditScoreAggregate data to the credit_score_aggregate table.
     *
     * @param aggregate The CreditScoreAggregate object to save.
     * @throws SQLException if a database access error occurs.
     */
    public void saveCreditScoreAggregate(CreditScoreAggregate aggregate) throws SQLException {
        String sql = "INSERT INTO credit_score_aggregate " +
                     "(window_end, region, city, district, level, age_group, gender, avg_score, citizen_count) " +
                     "VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)";

        LOG.debug("Saving CreditScoreAggregate to ClickHouse: {}", aggregate);
        try (PreparedStatement pstmt = connection.prepareStatement(sql)) {
            pstmt.setTimestamp(1, aggregate.getWindowEndDateTime());
            pstmt.setString(2, aggregate.region);
            pstmt.setString(3, aggregate.city);
            pstmt.setString(4, aggregate.district);
            pstmt.setString(5, aggregate.level);
            pstmt.setString(6, aggregate.ageGroup);
            pstmt.setString(7, aggregate.gender);
            pstmt.setDouble(8, aggregate.avgScore);
            pstmt.setLong(9, aggregate.citizenCount);

            pstmt.executeUpdate();
            LOG.info("Successfully queued CreditScoreAggregate for region {} to ClickHouse.", aggregate.region);
        } catch (SQLException e) {
            LOG.error("Failed to queue CreditScoreAggregate for ClickHouse: {}. Error: {}", aggregate, e.getMessage(), e);
            throw e;
        }
    }

    /**
     * Saves SocialGraphAggregate data to the social_graph_aggregate table.
     *
     * @param aggregate The SocialGraphAggregate object to save.
     * @throws SQLException if a database access error occurs.
     */
    public void saveSocialGraphAggregate(SocialGraphAggregate aggregate) throws SQLException {
        String sql = "INSERT INTO social_graph_aggregate " +
                     "(window_end, region, city, district, age_group, gender, total_relations, avg_relations_per_citizen) " +
                     "VALUES (?, ?, ?, ?, ?, ?, ?, ?)";

        LOG.debug("Saving SocialGraphAggregate to ClickHouse: {}", aggregate);
        try (PreparedStatement pstmt = connection.prepareStatement(sql)) {
            pstmt.setTimestamp(1, aggregate.getWindowEndDateTime());
            pstmt.setString(2, aggregate.region);
            pstmt.setString(3, aggregate.city);
            pstmt.setString(4, aggregate.district);
            pstmt.setString(5, aggregate.ageGroup);
            pstmt.setString(6, aggregate.gender);
            pstmt.setLong(7, aggregate.totalRelations);
            pstmt.setDouble(8, aggregate.avgRelationsPerCitizen);

            pstmt.executeUpdate();
            LOG.info("Successfully queued SocialGraphAggregate for region {} to ClickHouse.", aggregate.region);
        } catch (SQLException e) {
            LOG.error("Failed to queue SocialGraphAggregate for ClickHouse: {}. Error: {}", aggregate, e.getMessage(), e);
            throw e;
        }
    }

    /**
     * Saves ScoreVelocityMetrics data to the score_velocity_metrics table.
     *
     * @param metrics The ScoreVelocityMetrics object to save.
     * @throws SQLException if a database access error occurs.
     */
    public void saveVelocityMetrics(ScoreVelocityMetrics metrics) throws SQLException {
         String sql = "INSERT INTO score_velocity_metrics " +
                      "(window_end, region, city, district, score_change_1h, score_change_24h, positive_events_1h, negative_events_1h) " +
                      "VALUES (?, ?, ?, ?, ?, ?, ?, ?)";

         LOG.debug("Saving ScoreVelocityMetrics to ClickHouse: {}", metrics);
         try (PreparedStatement pstmt = connection.prepareStatement(sql)) {
             // Assuming 'window_end' corresponds to calculationTimestamp
             pstmt.setTimestamp(1, metrics.getCalculationEndDateTime());
             pstmt.setString(2, metrics.region);
             pstmt.setString(3, metrics.city);
             pstmt.setString(4, metrics.district);
             pstmt.setDouble(5, metrics.scoreChange1h);
             pstmt.setDouble(6, metrics.scoreChange24h);
             pstmt.setLong(7, metrics.positiveEvents1h);
             pstmt.setLong(8, metrics.negativeEvents1h);

             pstmt.executeUpdate();
             LOG.info("Successfully queued ScoreVelocityMetrics for region {} to ClickHouse.", metrics.region);
         } catch (SQLException e) {
             LOG.error("Failed to queue ScoreVelocityMetrics for ClickHouse: {}. Error: {}", metrics, e.getMessage(), e);
             throw e;
         }
    }

    /**
     * Commits the current transaction.
     * Should be called periodically or at the end of a batch in the SinkFunction.
     * @throws SQLException if a database access error occurs or the connection is closed.
     */
    public void commit() throws SQLException {
        if (connection != null && !connection.isClosed()) {
            try {
                connection.commit();
                LOG.debug("ClickHouse transaction committed.");
            } catch (SQLException e) {
                LOG.error("Failed to commit ClickHouse transaction.", e);
                throw e;
            }
        } else {
             LOG.warn("Attempted to commit on a closed or null ClickHouse connection.");
        }
    }

    /**
     * Rolls back the current transaction in case of an error.
     */
    public void rollback() {
        if (connection != null) {
            try {
                if (!connection.isClosed()) {
                    connection.rollback();
                    LOG.warn("ClickHouse transaction rolled back.");
                }
            } catch (SQLException e) {
                LOG.error("Failed to rollback ClickHouse transaction.", e);
            }
        }
    }

    /**
     * Closes the database connection.
     */
    @Override
    public void close() {
        if (connection != null) {
            try {
                if (!connection.isClosed()) {
                    connection.close();
                    LOG.info("ClickHouse connection closed.");
                }
            } catch (SQLException e) {
                LOG.error("Error closing ClickHouse connection.", e);
            } finally {
                connection = null; // Ensure connection is set to null
            }
        }
    }
}