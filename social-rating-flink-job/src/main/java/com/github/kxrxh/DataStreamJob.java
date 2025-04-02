package com.github.kxrxh;

import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.api.common.state.MapStateDescriptor;
import org.apache.flink.api.common.typeinfo.BasicTypeInfo;
import org.apache.flink.api.common.typeinfo.TypeHint;
import org.apache.flink.api.common.typeinfo.TypeInformation;
import org.apache.flink.connector.kafka.source.KafkaSource;
import org.apache.flink.streaming.api.datastream.BroadcastStream;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

// Import model classes
import com.github.kxrxh.model.CitizenEvent;
import com.github.kxrxh.model.ScoringRule;
import com.github.kxrxh.model.ScoreUpdate;
import com.github.kxrxh.model.ScoreSnapshot;
import com.github.kxrxh.model.RelatedEffectTrigger;
import com.github.kxrxh.model.EventAggregate;
import com.github.kxrxh.model.CreditScoreAggregate;
import com.github.kxrxh.model.ScoreSnapshotWithRelationCount;
import com.github.kxrxh.model.SocialGraphAggregate;
import com.github.kxrxh.model.ScoreVelocityMetrics;
// Import repository
import com.github.kxrxh.repositories.MongoRepository;
// Import sinks
import com.github.kxrxh.sinks.InfluxDBSinkFunction;
import com.github.kxrxh.sinks.MongoDbSnapshotSinkFunction;
import com.github.kxrxh.sinks.DgraphSinkFunction;
import com.github.kxrxh.sinks.ClickHouseEventAggregateSink;
import com.github.kxrxh.sinks.ClickHouseCreditScoreSink;
import com.github.kxrxh.sinks.MongoDbFirstSeenSinkFunction;
import com.github.kxrxh.sinks.ClickHouseSocialGraphSink;
import com.github.kxrxh.sinks.ClickHouseVelocitySink;
// Import the stateful function
import com.github.kxrxh.functions.StatefulApplyRulesFunction;
import com.github.kxrxh.functions.AsyncRelatedEffectFunction;
import com.github.kxrxh.functions.AsyncEnrichWithRelationCountFunction;
import com.github.kxrxh.functions.CalculateVelocityFunction;
import com.github.kxrxh.functions.AsyncEnrichWithDemographicsFunction;
import org.apache.flink.streaming.api.datastream.AsyncDataStream;
import org.apache.flink.streaming.api.datastream.SingleOutputStreamOperator;
import org.apache.flink.connector.kafka.sink.KafkaRecordSerializationSchema;
import org.apache.flink.connector.kafka.sink.KafkaSink;
import org.apache.flink.api.common.serialization.SerializationSchema;
import org.apache.flink.api.common.serialization.SimpleStringSchema;
import com.google.gson.Gson;

// Import the new source function
import com.github.kxrxh.sources.MongoRulePollingSource;

// Imports for Aggregation
import org.apache.flink.streaming.api.windowing.assigners.TumblingProcessingTimeWindows; // Or TumblingEventTimeWindows
import org.apache.flink.streaming.api.windowing.time.Time;
import org.apache.flink.api.java.functions.KeySelector;
import com.github.kxrxh.functions.EventAggregator; // Import aggregator function
import com.github.kxrxh.functions.CreditScoreAggregator;
import com.github.kxrxh.functions.SocialGraphAggregator;
import org.apache.flink.streaming.api.windowing.windows.TimeWindow;
import org.apache.flink.streaming.api.functions.windowing.ProcessWindowFunction;
import org.apache.flink.streaming.api.windowing.windows.Window;
import org.apache.flink.util.Collector;

import org.apache.flink.api.java.tuple.Tuple2;

import java.util.ArrayList; // Needed for empty list fallback
import java.util.List;
import java.util.concurrent.TimeUnit; // For Async I/O timeout
import org.apache.flink.api.java.utils.ParameterTool;
import com.github.kxrxh.factories.KafkaSourceFactory;
import com.github.kxrxh.config.ParameterNames;

/**
 * Flink Streaming Job for calculating Social Rating.
 *
 * Reads events from Redpanda, applies scoring rules statefully (fetched from MongoDB broadcast state),
 * and writes results to various databases.
 * Configuration is passed via command-line arguments.
 */
public class DataStreamJob {

	private static final Logger LOG = LoggerFactory.getLogger(DataStreamJob.class);

	// Descriptor for broadcast state (Rule ID -> ScoringRule)
	public static final MapStateDescriptor<String, ScoringRule> rulesStateDescriptor = 
		new MapStateDescriptor<>(
			"ScoringRulesBroadcastState",
			BasicTypeInfo.STRING_TYPE_INFO, // Key type: Rule ID
			TypeInformation.of(new TypeHint<ScoringRule>() {}) // Value type: ScoringRule POJO
		);


	public static void main(String[] args) throws Exception {
		final StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment();

		// --- Configuration via ParameterTool ---
		final ParameterTool params = ParameterTool.fromArgs(args);

        // Register parameters globally so they can be accessed in RichFunctions
        env.getConfig().setGlobalJobParameters(params);

        // Fetch required parameters (job fails if missing)
        final String kafkaBrokers = params.getRequired(ParameterNames.KAFKA_BROKERS);
        final String inputTopic = params.getRequired(ParameterNames.KAFKA_INPUT_TOPIC);
        final String consumerGroupId = params.getRequired(ParameterNames.KAFKA_CONSUMER_GROUP);
        final String influxDbUrl = params.getRequired(ParameterNames.INFLUXDB_URL);
        final String influxDbToken = params.getRequired(ParameterNames.INFLUXDB_TOKEN);
        final String influxDbOrg = params.getRequired(ParameterNames.INFLUXDB_ORG);
        final String influxDbBucket = params.getRequired(ParameterNames.INFLUXDB_BUCKET);
        final String mongoUri = params.getRequired(ParameterNames.MONGO_URI);
        final String mongoDbName = params.getRequired(ParameterNames.MONGO_DB_NAME);
        final String dgraphUri = params.getRequired(ParameterNames.DGRAPH_URI);
        final String secondaryUpdateTopic = params.getRequired(ParameterNames.KAFKA_SECONDARY_UPDATE_TOPIC);
        final long mongoPollingIntervalSeconds = params.getLong(ParameterNames.MONGO_POLLING_INTERVAL_SECONDS, 60L); // Default 60s
        final String clickhouseJdbcUrl = params.getRequired(ParameterNames.CLICKHOUSE_JDBC_URL);
        // Async timeouts/capacities (optional parameters with defaults)
        final long enrichDemoTimeoutMs = params.getLong(ParameterNames.ASYNC_DEMO_TIMEOUT_MS, 5000L);
        final int enrichDemoCapacity = params.getInt(ParameterNames.ASYNC_DEMO_CAPACITY, 100);
        final long enrichRelTimeoutMs = params.getLong(ParameterNames.ASYNC_REL_TIMEOUT_MS, 10000L);
        final int enrichRelCapacity = params.getInt(ParameterNames.ASYNC_REL_CAPACITY, 100);
        final long relatedEffectTimeoutMs = params.getLong(ParameterNames.ASYNC_RELATED_EFFECT_TIMEOUT_MS, 10000L);
        final int relatedEffectCapacity = params.getInt(ParameterNames.ASYNC_RELATED_EFFECT_CAPACITY, 100);


		// --- 1. Configure Redpanda (Kafka) Source ---
        // Assuming KafkaSourceFactory uses ParameterTool or takes parameters
        KafkaSource<CitizenEvent> source = KafkaSourceFactory.createCitizenEventSource(
            kafkaBrokers, inputTopic, consumerGroupId
            // Alternatively, pass params to KafkaSourceFactory if it's designed for it
            // params 
        );
		DataStream<CitizenEvent> parsedEvents = env.fromSource(source, WatermarkStrategy.noWatermarks(), "Redpanda Source (JSON)")
			.name("ParsedCitizenEvents");

		// --- 1a. Enrichment: Add Demographics (Async) --- 
		DataStream<CitizenEvent> enrichedParsedEvents = AsyncDataStream.unorderedWait(
			parsedEvents, 
			new AsyncEnrichWithDemographicsFunction(), // Use parameterless constructor
			enrichDemoTimeoutMs, 
			TimeUnit.MILLISECONDS,
			enrichDemoCapacity 
		).name("EnrichWithDemographics");

		// --- 1b. Source: Rules from MongoDB (Polling) & Broadcast --- 
        DataStream<ScoringRule> rulesStream = env
            .addSource(new MongoRulePollingSource(mongoUri, mongoDbName, mongoPollingIntervalSeconds))
            .name("MongoRulePollingSource")
            .setParallelism(1); 
		
		BroadcastStream<ScoringRule> broadcastRulesStream = rulesStream
			.broadcast(rulesStateDescriptor);

		// --- 2. Process Events using Keyed Broadcast State ---
		SingleOutputStreamOperator<ScoreSnapshot> scoreSnapshots = enrichedParsedEvents
			.keyBy(event -> event.citizenId) 
			.connect(broadcastRulesStream) 
			.process(new StatefulApplyRulesFunction(rulesStateDescriptor)) 
			.name("ApplyScoringRulesStatefully"); 

        // --- Enrichment: Add Relation Count (Async) --- 
        DataStream<ScoreSnapshotWithRelationCount> enrichedSnapshots = AsyncDataStream.unorderedWait(
            scoreSnapshots, 
            new AsyncEnrichWithRelationCountFunction(), // Use parameterless constructor
            enrichRelTimeoutMs, 
            TimeUnit.MILLISECONDS,
            enrichRelCapacity
        ).name("EnrichWithRelationCount");

        // --- 2x. Processing: Get First Seen Timestamps --- 
        DataStream<Tuple2<String, Long>> firstSeenEvents = scoreSnapshots
            .getSideOutput(StatefulApplyRulesFunction.firstSeenOutputTag);

        // --- 2b. Processing: Handle Related Effects --- 
        DataStream<RelatedEffectTrigger> relatedEffectTriggers = scoreSnapshots
            .getSideOutput(StatefulApplyRulesFunction.relatedEffectOutputTag);

        DataStream<ScoreUpdate> secondaryUpdates = AsyncDataStream.unorderedWait(
            relatedEffectTriggers, 
            new AsyncRelatedEffectFunction(params.getRequired(ParameterNames.DGRAPH_URI)), // Pass Dgraph URI
            relatedEffectTimeoutMs, 
            TimeUnit.MILLISECONDS,
            relatedEffectCapacity
        ).name("QueryRelatedCitizensAndGenerateUpdates");

        // --- 2c. Processing: Aggregate Events for ClickHouse --- 
        DataStream<EventAggregate> eventAggregates = scoreSnapshots
            .keyBy(new KeySelector<ScoreSnapshot, String>() {
                @Override
                public String getKey(ScoreSnapshot s) throws Exception {
                    // Add ageGroup and gender to the key
                    return String.join("|",
                        s.region != null ? s.region : "NULL",
                        s.city != null ? s.city : "NULL",
                        s.district != null ? s.district : "NULL",
                        s.eventType != null ? s.eventType : "NULL",
                        s.eventSubtype != null ? s.eventSubtype : "NULL",
                        s.ageGroup != null ? s.ageGroup : "NULL", // Add ageGroup
                        s.gender != null ? s.gender : "NULL"   // Add gender
                    );
                }
            })
            .window(TumblingProcessingTimeWindows.of(Time.minutes(1)))
            .aggregate(new EventAggregator(), new EventAggregateWindowFunction())
            .name("AggregateEventsPerWindow");

        // --- 2d. Processing: Aggregate Credit Scores for ClickHouse --- 
        DataStream<CreditScoreAggregate> creditScoreAggregates = scoreSnapshots
            .keyBy(new KeySelector<ScoreSnapshot, String>() {
                @Override
                public String getKey(ScoreSnapshot s) throws Exception {
                    // Key by Region, City, District, Level, ageGroup, gender
                    return String.join("|",
                        s.region != null ? s.region : "NULL",
                        s.city != null ? s.city : "NULL",
                        s.district != null ? s.district : "NULL",
                        s.level != null ? s.level : "NULL",
                        s.ageGroup != null ? s.ageGroup : "NULL", // Add ageGroup
                        s.gender != null ? s.gender : "NULL"   // Add gender
                    );
                }
            })
            .window(TumblingProcessingTimeWindows.of(Time.minutes(1))) // Same window as event aggregates
            .aggregate(new CreditScoreAggregator(), new CreditScoreAggregateWindowFunction()) // Use the new aggregator and window function
            .name("AggregateCreditScoresPerWindow");

        // --- 2e. Processing: Aggregate Social Graph Metrics for ClickHouse --- 
        DataStream<SocialGraphAggregate> socialGraphAggregates = enrichedSnapshots 
            .keyBy(new KeySelector<ScoreSnapshotWithRelationCount, String>() {
                @Override
                public String getKey(ScoreSnapshotWithRelationCount s) throws Exception {
                    // Key by Region, City, District, ageGroup, gender
                    return String.join("|",
                        s.region != null ? s.region : "NULL",
                        s.city != null ? s.city : "NULL",
                        s.district != null ? s.district : "NULL",
                        s.ageGroup != null ? s.ageGroup : "NULL", // Add ageGroup
                        s.gender != null ? s.gender : "NULL"   // Add gender
                    );
                }
            })
            .window(TumblingProcessingTimeWindows.of(Time.minutes(1))) // Same window
            .aggregate(new SocialGraphAggregator(), new SocialGraphAggregateWindowFunction()) // Use SocialGraph aggregator/window func
            .name("AggregateSocialGraphMetricsPerWindow");

        // --- 2f. Processing: Calculate Score Velocity Metrics --- 
        DataStream<ScoreVelocityMetrics> velocityMetrics = scoreSnapshots
            .keyBy(new KeySelector<ScoreSnapshot, String>() {
                @Override
                public String getKey(ScoreSnapshot s) throws Exception {
                    // Key by Region, City, District
                    return String.join("|",
                        s.region != null ? s.region : "NULL",
                        s.city != null ? s.city : "NULL",
                        s.district != null ? s.district : "NULL"
                    );
                }
            })
            .process(new CalculateVelocityFunction()) // Apply the KeyedProcessFunction
            .name("CalculateScoreVelocity");

		// --- 3. Configure Sinks --- 
		// (Sinks will often access config via getRuntimeContext().getExecutionConfig().getGlobalJobParameters() in open())
		
		// 3.1 Sink: MongoDB Snapshots
		scoreSnapshots.addSink(new MongoDbSnapshotSinkFunction()) // Use parameterless constructor
			  .name("MongoDB Sink (Snapshots)")
			  .setParallelism(1); 

		// 3.2 Sink: Dgraph Score History
		scoreSnapshots.addSink(new DgraphSinkFunction()) // Use parameterless constructor
			  .name("Dgraph Sink (Score History)")
			  .setParallelism(1); 

		// 3.3 Sink: InfluxDB Score Changes
		scoreSnapshots.addSink(new InfluxDBSinkFunction()) // Use parameterless constructor
			.setParallelism(1) 
			.name("InfluxDB Sink (Score History)");
		
		// 3.4 Sink: Secondary Updates to Kafka
        KafkaSink<ScoreUpdate> secondaryUpdatesKafkaSink = KafkaSink.<ScoreUpdate>builder()
            .setBootstrapServers(kafkaBrokers)
            .setRecordSerializer(KafkaRecordSerializationSchema.builder()
                .setTopic(secondaryUpdateTopic)
                .setValueSerializationSchema(new JsonSerializationSchema<ScoreUpdate>())
                .build()
            )
            .build();

        secondaryUpdates.sinkTo(secondaryUpdatesKafkaSink)
            .name("Kafka Sink (Secondary Updates)");

        // 3.5 Sink: ClickHouse Event Aggregates
         eventAggregates.addSink(new ClickHouseEventAggregateSink()) // Use parameterless constructor
             .name("ClickHouse Sink (Event Aggregates)")
             .setParallelism(1); 

        // 3.6 Sink: ClickHouse Credit Score Aggregates
        creditScoreAggregates.addSink(new ClickHouseCreditScoreSink()) // Use parameterless constructor
            .name("ClickHouse Sink (Credit Score Aggregates)")
            .setParallelism(1); 

        // 3.7 Sink: MongoDB First Seen Timestamps
        firstSeenEvents.addSink(new MongoDbFirstSeenSinkFunction()) // Use parameterless constructor
            .name("MongoDB Sink (First Seen Timestamps)")
            .setParallelism(1); 

        // 3.8 Sink: ClickHouse Social Graph Aggregates
        socialGraphAggregates.addSink(new ClickHouseSocialGraphSink()) // Use parameterless constructor
            .name("ClickHouse Sink (Social Graph Aggregates)")
            .setParallelism(1); 

        // 3.9 Sink: ClickHouse Velocity Metrics
        velocityMetrics.addSink(new ClickHouseVelocitySink()) // Use parameterless constructor
            .name("ClickHouse Sink (Velocity Metrics)")
            .setParallelism(1); 

		// --- 4. Execute Job ---
		env.execute("Social Rating Flink Job");
	}

    // Simple JSON Serialization Schema using Gson (Add Gson dependency to pom.xml)
    private static class JsonSerializationSchema<T> implements SerializationSchema<T> {
        private static final Gson gson = new Gson();

        @Override
        public byte[] serialize(T element) {
            return gson.toJson(element).getBytes(java.nio.charset.StandardCharsets.UTF_8);
        }
    }

    // ProcessWindowFunction to add window start/end times to the aggregate result
    public static class EventAggregateWindowFunction extends ProcessWindowFunction<EventAggregate, EventAggregate, String, TimeWindow> {
        @Override
        public void process(String key, Context context, Iterable<EventAggregate> aggregates, Collector<EventAggregate> out) throws Exception {
            EventAggregate aggregate = aggregates.iterator().next();
            aggregate.windowStart = context.window().getStart();
            aggregate.windowEnd = context.window().getEnd();
            out.collect(aggregate);
        }
    }

    // ProcessWindowFunction to add window start/end times to the CreditScoreAggregate result
    public static class CreditScoreAggregateWindowFunction extends ProcessWindowFunction<CreditScoreAggregate, CreditScoreAggregate, String, TimeWindow> {
        @Override
        public void process(String key, Context context, Iterable<CreditScoreAggregate> aggregates, Collector<CreditScoreAggregate> out) throws Exception {
            CreditScoreAggregate aggregate = aggregates.iterator().next(); // Expecting one result per key/window from AggregateFunction
            aggregate.windowStart = context.window().getStart();
            aggregate.windowEnd = context.window().getEnd();
            out.collect(aggregate);
        }
    }

    // ProcessWindowFunction to add window start/end times to the SocialGraphAggregate result
    public static class SocialGraphAggregateWindowFunction extends ProcessWindowFunction<SocialGraphAggregate, SocialGraphAggregate, String, TimeWindow> {
        @Override
        public void process(String key, Context context, Iterable<SocialGraphAggregate> aggregates, Collector<SocialGraphAggregate> out) throws Exception {
            SocialGraphAggregate aggregate = aggregates.iterator().next(); // Expecting one result per key/window
            aggregate.windowStart = context.window().getStart();
            aggregate.windowEnd = context.window().getEnd();
            out.collect(aggregate);
        }
    }

}
