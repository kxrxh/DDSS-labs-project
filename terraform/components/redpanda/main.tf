resource "helm_release" "redpanda" {
  name       = var.redpanda_name
  repository = var.redpanda_repository
  chart      = var.redpanda_chart_name
  version    = var.redpanda_chart_version
  namespace  = var.redpanda_namespace
  create_namespace = true

  values = [
    <<-EOT
    # Optimized Redpanda configuration for high-throughput stream processing
    
    # Single replica for memory-constrained environments
    statefulset:
      replicas: 1 # Single replica to save memory
      podAntiAffinity:
        type: soft # Keep soft for single-node dev environments

    # Memory-optimized resources for development
    resources:
      requests:
        cpu: 1     # Balanced CPU for good performance
        memory: "2Gi"  # Reduced memory for constrained systems
      limits:
        cpu: 2     # Allow burst to 2 CPU when available
        memory: "2Gi"  # Match requests for predictable memory usage

    # Storage optimization for high throughput
    storage:
      storageClassName: "" # Use default storage class
      size: "10Gi" # Reduced storage for memory-constrained systems

    # Redpanda configuration optimized for stream processing
    config:
      redpanda:
        # Message size limits - accommodate larger batches
        max_message_size: 10485760  # 10MB (increased for batch processing)
        
        # Consumer optimization
        fetch_max_bytes: 52428800   # 50MB total fetch size
        fetch_max_partition_bytes: 10485760  # 10MB per partition
        
        # Producer optimization  
        compression_type: snappy    # Good balance of speed vs compression
        batch_max_size: 1048576     # 1MB batch size for producers
        
        # Topic defaults optimized for processing
        default_topic_partitions: 3  # Fewer partitions for single broker
        default_topic_replications: 1  # Single replica to match broker count
        
        # Memory and performance tuning
        seastar_memory: "1.5Gi"     # Reduced for memory-constrained systems
        memory_per_core: 512        # MB per core, reduced for lower memory
        
        # Retention and cleanup
        log_segment_size: 134217728 # 128MB segments
        log_retention_ms: 604800000 # 7 days retention
        log_compaction_interval_ms: 300000  # 5 minutes
        
        # Timeouts optimized for processing workload
        group_max_session_timeout_ms: 300000   # 5 minutes
        group_min_session_timeout_ms: 6000     # 6 seconds
        
        # Replication optimization
        raft_heartbeat_interval_ms: 150
        raft_heartbeat_timeout_ms: 3000
        
    # Disable TLS for development environment
    tls:
      enabled: false
    
    listeners:
      kafka:
        tls:
          enabled: false
      rpc:
        tls:
          enabled: false
      admin:
        tls:
          enabled: false
      http:
        tls:
          enabled: false
      schemaRegistry:
        tls:
          enabled: false
        
    # Console configuration for monitoring
    console:
      enabled: true
      resources:
        requests:
          cpu: 100m
          memory: 128Mi
        limits:
          cpu: 200m
          memory: 256Mi

    # Enable JMX for monitoring (optional)
    monitoring:
      enabled: false
      port: 9644

    # Network policies for security
    networkPolicy:
      enabled: false  # Set to true if using network policies
    
    EOT
  ]

  set {
    name  = "console.enabled"
    value = var.redpanda_console_enabled
  }

  # Set topic configuration optimized for stream processing
  set {
    name  = "config.redpanda.topic_partitions_per_shard"
    value = "1000"  # Allow more partitions per shard
  }

  set {
    name  = "config.redpanda.disable_batch_cache"
    value = "false"  # Keep batch cache for better performance
  }

  set {
    name  = "config.redpanda.enable_transactions"
    value = "false"  # Disable if not needed for better performance
  }
} 