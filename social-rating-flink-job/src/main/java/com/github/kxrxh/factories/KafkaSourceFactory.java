package com.github.kxrxh.factories;

import com.github.kxrxh.model.CitizenEvent;
import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.connector.kafka.source.KafkaSource;
import org.apache.flink.connector.kafka.source.enumerator.initializer.OffsetsInitializer;
import org.apache.flink.formats.json.JsonDeserializationSchema; // Import standard Flink JSON deserializer
import org.apache.flink.api.common.typeinfo.TypeInformation; // Import TypeInformation
import org.apache.flink.api.java.utils.ParameterTool; // Consider if needed, passing strings is simpler for now

/**
 * Factory class for creating KafkaSource instances.
 */
public class KafkaSourceFactory {

    /**
     * Creates a KafkaSource for consuming CitizenEvent objects.
     *
     * @param brokers       Comma-separated list of Kafka broker addresses.
     * @param topic         The Kafka topic to consume from.
     * @param consumerGroup The consumer group ID.
     * @return Configured KafkaSource<CitizenEvent>.
     */
    public static KafkaSource<CitizenEvent> createCitizenEventSource(
            String brokers, 
            String topic, 
            String consumerGroup) {

        // Use Flink's built-in JSON deserialization schema
        JsonDeserializationSchema<CitizenEvent> jsonSchema = 
            new JsonDeserializationSchema<>(CitizenEvent.class);

        return KafkaSource.<CitizenEvent>builder()
                .setBootstrapServers(brokers)
                .setTopics(topic)
                .setGroupId(consumerGroup)
                .setStartingOffsets(OffsetsInitializer.earliest()) // Or latest(), or specific offsets
                .setValueOnlyDeserializer(jsonSchema) // Use the JSON schema
                // Add any other necessary configurations, e.g., Kafka properties
                // .setProperty("security.protocol", "SASL_SSL") 
                // .setProperty(...) 
                .build();
    }

    // Private constructor to prevent instantiation if all methods are static
    private KafkaSourceFactory() {}
} 