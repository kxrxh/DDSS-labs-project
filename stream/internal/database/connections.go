package database

import (
	"context"
	"log"
	"time"

	"github.com/gocql/gocql"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/kxrxh/social-rating-system/stream/internal/config"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InitMongoDB initializes MongoDB connection
func InitMongoDB(ctx context.Context, cfg *config.Config) (*mongo.Client, error) {
	// Optimize MongoDB client for high throughput
	clientOptions := options.Client().ApplyURI(cfg.MongoURI).
		SetMaxPoolSize(100).                       // Increased connection pool
		SetMinPoolSize(10).                        // Minimum pool size
		SetMaxConnIdleTime(30 * time.Second).      // Connection idle timeout
		SetConnectTimeout(10 * time.Second).       // Connection timeout
		SetSocketTimeout(30 * time.Second).        // Socket timeout
		SetServerSelectionTimeout(5 * time.Second) // Server selection timeout

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB with optimized settings")
	return client, nil
}

// InitInfluxDB initializes InfluxDB connection
func InitInfluxDB(cfg *config.Config) influxdb2.Client {
	client := influxdb2.NewClient(cfg.InfluxURL, cfg.InfluxToken)
	log.Printf("Connected to InfluxDB at %s", cfg.InfluxURL)
	log.Printf("InfluxDB Organization: %s", cfg.InfluxOrg)
	log.Printf("InfluxDB Raw Events Bucket: %s", cfg.InfluxRawEventsBucket)
	log.Printf("InfluxDB Derived Metrics Bucket: %s", cfg.InfluxDerivedMetricsBucket)

	// Test connection by trying to ping the server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	health, err := client.Health(ctx)
	if err != nil {
		log.Printf("WARNING: Could not verify InfluxDB health: %v", err)
	} else {
		log.Printf("InfluxDB health status: %s (version: %s)", health.Status, *health.Version)
	}

	return client
}

// InitScyllaDB initializes ScyllaDB connection
func InitScyllaDB(cfg *config.Config) (*gocql.Session, error) {
	cluster := gocql.NewCluster(cfg.ScyllaHosts...)
	cluster.Keyspace = cfg.ScyllaKeyspace
	cluster.Consistency = gocql.LocalQuorum // Better performance than Quorum
	cluster.ConnectTimeout = 10 * time.Second
	cluster.Timeout = 5 * time.Second // Reduced timeout

	// Optimize for high throughput
	cluster.NumConns = 4             // More connections per host
	cluster.MaxPreparedStmts = 1000  // More prepared statements
	cluster.MaxRoutingKeyInfo = 1000 // Routing optimization
	cluster.PageSize = 5000          // Larger page size
	cluster.Consistency = gocql.LocalQuorum

	// Connection pooling
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	log.Println("Connected to ScyllaDB with optimized settings")
	return session, nil
}

// CreateKafkaClient creates a Kafka client for consuming events
func CreateKafkaClient(cfg *config.Config) (*kgo.Client, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.KafkaBrokers...),
		kgo.ConsumeTopics(cfg.KafkaTopic),
		kgo.ConsumerGroup(cfg.KafkaGroupID),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),

		// Optimize for high throughput and stability
		kgo.FetchMinBytes(1),                         // Allow small fetches to reduce latency
		kgo.FetchMaxBytes(50 * 1024 * 1024),          // Increased to 50MB for better batch processing
		kgo.FetchMaxWait(500 * time.Millisecond),     // Balanced wait time
		kgo.FetchMaxPartitionBytes(10 * 1024 * 1024), // 10MB per partition

		// Session and heartbeat optimization for stability
		kgo.SessionTimeout(60 * time.Second),    // Increased for stability
		kgo.HeartbeatInterval(10 * time.Second), // More conservative heartbeat
		kgo.RebalanceTimeout(60 * time.Second),  // Longer rebalance timeout

		// Auto commit settings for better performance
		kgo.AutoCommitInterval(5 * time.Second), // Commit every 5 seconds
		kgo.DisableAutoCommit(),                 // We'll commit manually for better control

		// Connection optimization
		kgo.RequestRetries(3),
		kgo.RequestTimeoutOverhead(10 * time.Second),

		kgo.WithLogger(kgo.BasicLogger(log.Writer(), kgo.LogLevelInfo, func() string {
			return "KAFKA: "
		})),
	}

	return kgo.NewClient(opts...)
}
