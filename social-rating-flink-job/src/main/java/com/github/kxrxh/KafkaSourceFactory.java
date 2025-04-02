package com.github.kxrxh;

// import org.apache.flink.api.common.serialization.SimpleStringSchema; // No longer needed
import org.apache.flink.connector.kafka.source.KafkaSource;
import org.apache.flink.connector.kafka.source.enumerator.initializer.OffsetsInitializer;
import org.apache.flink.formats.json.JsonDeserializationSchema;
import com.github.kxrxh.model.CitizenEvent;

/**
 * Factory class to create KafkaSource instances for the Social Rating job.
 */
public class KafkaSourceFactory {

    /**
     * Creates a KafkaSource for reading and deserializing CitizenEvent JSON objects.
     *
     * @param brokers          The Kafka broker list (e.g., "localhost:9092").
     * @param topic            The topic to read from (e.g., "social_events").
     * @param consumerGroupId The consumer group ID.
     * @return A configured KafkaSource<CitizenEvent>.
     */
    public static KafkaSource<CitizenEvent> createCitizenEventSource(String brokers, String topic, String consumerGroupId) {
        return KafkaSource.<CitizenEvent>builder()
                .setBootstrapServers(brokers)
                .setTopics(topic)
                .setGroupId(consumerGroupId)
                .setStartingOffsets(OffsetsInitializer.earliest()) // Or latest()
                // Use JsonDeserializationSchema to map JSON to the CitizenEvent POJO
                .setValueOnlyDeserializer(new JsonDeserializationSchema<>(CitizenEvent.class))
                .build();
    }

    // Keep the string source method if needed elsewhere, or remove it.
    // public static KafkaSource<String> createStringSource(String brokers, String topic, String consumerGroupId) { ... }
}
