package com.github.kxrxh.repositories;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.SQLException;
import java.util.Properties;

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

    // Configuration (Consider externalizing this)
    private static final String CLICKHOUSE_JDBC_URL = "jdbc:clickhouse://localhost:8123/default"; // Example
    // Add user/password if required by your ClickHouse setup
    // private static final String CLICKHOUSE_USER = "user";
    // private static final String CLICKHOUSE_PASSWORD = "password";
    private static final String CLICKHOUSE_DRIVER = "com.clickhouse.jdbc.ClickHouseDriver";

    private Connection connection;

    public ClickHouseRepository() {
        this(CLICKHOUSE_JDBC_URL, null); // Pass null for properties if no user/pass needed
    }

    public ClickHouseRepository(String jdbcUrl, Properties properties) {
        try {
            // Ensure the JDBC driver is loaded
            Class.forName(CLICKHOUSE_DRIVER);

            LOG.info("Connecting to ClickHouse at: {}", jdbcUrl);
            if (properties == null) {
                this.connection = DriverManager.getConnection(jdbcUrl);
            } else {
                // Example if using user/password:
                // properties.setProperty("user", CLICKHOUSE_USER);
                // properties.setProperty("password", CLICKHOUSE_PASSWORD);
                this.connection = DriverManager.getConnection(jdbcUrl, properties);
            }
            // Disable auto-commit for potential batching later
            this.connection.setAutoCommit(false);

            LOG.info("ClickHouse connection established successfully.");

        } catch (ClassNotFoundException e) {
            LOG.error("ClickHouse JDBC Driver not found: {}", CLICKHOUSE_DRIVER, e);
            throw new RuntimeException("ClickHouse JDBC Driver not found", e);
        } catch (SQLException e) {
            LOG.error("Failed to establish ClickHouse connection to {}", jdbcUrl, e);
            throw new RuntimeException("Failed to establish ClickHouse connection", e);
        } catch (Exception e) {
            LOG.error("Failed to initialize ClickHouse repository", e);
            throw new RuntimeException("Failed to initialize ClickHouse repository", e);
        }
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
     * Should be called after a batch of operations.
     * @throws SQLException if a database access error occurs or auto-commit is enabled.
     */
    public void commit() throws SQLException {
        if (connection != null && !connection.getAutoCommit()) {
            connection.commit();
            LOG.debug("Committed ClickHouse transaction.");
        }
    }

    /**
     * Rolls back the current transaction.
     * Should be called if an error occurs during a batch of operations.
     */
    public void rollback() {
         try {
             if (connection != null && !connection.getAutoCommit()) {
                 connection.rollback();
                 LOG.warn("Rolled back ClickHouse transaction.");
             }
         } catch (SQLException ex) {
             LOG.error("ClickHouse rollback failed", ex);
         }
     }

    @Override
    public void close() {
        if (connection != null) {
            try {
                if (!connection.isClosed()) {
                    // Commit any pending changes before closing? Depends on strategy.
                    // connection.commit();
                    connection.close();
                    LOG.info("ClickHouse connection closed.");
                }
            } catch (SQLException e) {
                LOG.error("Error closing ClickHouse connection", e);
            }
        }
    }
}