package sinks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOSink handles backup operations to MinIO
type MinIOSink struct {
	client     *minio.Client
	bucketName string
	logger     *log.Logger
}

// BackupData represents the structure for backup data
type BackupData struct {
	Timestamp    time.Time   `json:"timestamp"`
	DataType     string      `json:"data_type"`
	Source       string      `json:"source"`
	Data         interface{} `json:"data"`
	EventCount   int         `json:"event_count"`
	BackupReason string      `json:"backup_reason"`
}

// NewMinIOSink creates a new MinIO sink instance
func NewMinIOSink(endpoint, accessKey, secretKey, bucketName string, useSSL bool, logger *log.Logger) (*MinIOSink, error) {
	// Initialize MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	sink := &MinIOSink{
		client:     client,
		bucketName: bucketName,
		logger:     logger,
	}

	// Test connection with retries
	if err := sink.testConnectionWithRetries(context.Background(), 5, 2*time.Second); err != nil {
		return nil, fmt.Errorf("failed to connect to MinIO: %w", err)
	}

	// Ensure all required buckets exist
	if err := sink.ensureAllBuckets(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure buckets exist: %w", err)
	}

	return sink, nil
}

// testConnectionWithRetries tests the connection to MinIO with retry logic
func (s *MinIOSink) testConnectionWithRetries(ctx context.Context, maxRetries int, retryDelay time.Duration) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		// Test connection by checking if we can list buckets
		_, err := s.client.ListBuckets(ctx)
		if err == nil {
			s.logger.Printf("✅ Successfully connected to MinIO (attempt %d/%d)", i+1, maxRetries)
			return nil
		}

		lastErr = err
		s.logger.Printf("❌ MinIO connection attempt %d/%d failed: %v", i+1, maxRetries, err)

		if i < maxRetries-1 {
			s.logger.Printf("⏳ Retrying in %v...", retryDelay)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
				// Continue to next retry
			}
		}
	}

	return fmt.Errorf("failed to connect to MinIO after %d retries: %w", maxRetries, lastErr)
}

// ensureAllBuckets ensures all required backup buckets exist
func (s *MinIOSink) ensureAllBuckets(ctx context.Context) error {
	// List of all required backup buckets
	requiredBuckets := []string{
		"mongodb-backups",
		"influxdb-backups",
		"scylladb-backups",
		"dgraph-backups",
		"redpanda-backups",
		"stream-backups",
		s.bucketName, // Also ensure the primary bucket exists
	}

	// Remove duplicates
	bucketSet := make(map[string]bool)
	for _, bucket := range requiredBuckets {
		bucketSet[bucket] = true
	}

	s.logger.Printf("Ensuring %d backup buckets exist...", len(bucketSet))

	for bucket := range bucketSet {
		if err := s.ensureBucket(ctx, bucket); err != nil {
			return fmt.Errorf("failed to ensure bucket %s exists: %w", bucket, err)
		}
	}

	s.logger.Printf("✅ All backup buckets are ready!")
	return nil
}

// ensureBucket ensures a specific bucket exists
func (s *MinIOSink) ensureBucket(ctx context.Context, bucketName string) error {
	// First, check if bucket exists with retry logic
	var exists bool
	var err error

	for retries := 0; retries < 3; retries++ {
		exists, err = s.client.BucketExists(ctx, bucketName)
		if err == nil {
			break
		}

		if retries < 2 {
			s.logger.Printf("⚠️  Retrying bucket existence check for %s (attempt %d/3): %v", bucketName, retries+1, err)
			time.Sleep(time.Duration(retries+1) * time.Second)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to check if bucket %s exists after retries: %w", bucketName, err)
	}

	if !exists {
		if err := s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
		}
		s.logger.Printf("✅ Created backup bucket: %s", bucketName)
	} else {
		s.logger.Printf("✓ Bucket already exists: %s", bucketName)
	}

	return nil
}

// getBucketForDataType returns the appropriate bucket name for a data type
func (s *MinIOSink) getBucketForDataType(dataType string) string {
	bucketMap := map[string]string{
		"processed_events": "stream-backups",
		"metrics":          "stream-backups",
		"user_profiles":    "stream-backups",
		"system_state":     "stream-backups",
		"mongodb":          "mongodb-backups",
		"influxdb":         "influxdb-backups",
		"scylladb":         "scylladb-backups",
		"dgraph":           "dgraph-backups",
		"redpanda":         "redpanda-backups",
	}

	if bucket, exists := bucketMap[dataType]; exists {
		return bucket
	}

	// Default to the configured bucket
	return s.bucketName
}

// BackupProcessedEvents backs up processed events to MinIO
func (s *MinIOSink) BackupProcessedEvents(ctx context.Context, events []interface{}, source string) error {
	if len(events) == 0 {
		return nil
	}

	backupData := BackupData{
		Timestamp:    time.Now(),
		DataType:     "processed_events",
		Source:       source,
		Data:         events,
		EventCount:   len(events),
		BackupReason: "periodic_backup",
	}

	return s.uploadBackup(ctx, backupData, "processed_events")
}

// BackupMetrics backs up metrics data to MinIO
func (s *MinIOSink) BackupMetrics(ctx context.Context, metrics interface{}, source string) error {
	backupData := BackupData{
		Timestamp:    time.Now(),
		DataType:     "metrics",
		Source:       source,
		Data:         metrics,
		EventCount:   1,
		BackupReason: "metrics_backup",
	}

	return s.uploadBackup(ctx, backupData, "metrics")
}

// BackupUserProfiles backs up user profile data to MinIO
func (s *MinIOSink) BackupUserProfiles(ctx context.Context, profiles interface{}, source string) error {
	backupData := BackupData{
		Timestamp:    time.Now(),
		DataType:     "user_profiles",
		Source:       source,
		Data:         profiles,
		EventCount:   1,
		BackupReason: "profile_backup",
	}

	return s.uploadBackup(ctx, backupData, "user_profiles")
}

// BackupSystemState backs up system configuration and state to MinIO
func (s *MinIOSink) BackupSystemState(ctx context.Context, state interface{}, source string) error {
	backupData := BackupData{
		Timestamp:    time.Now(),
		DataType:     "system_state",
		Source:       source,
		Data:         state,
		EventCount:   1,
		BackupReason: "system_state_backup",
	}

	return s.uploadBackup(ctx, backupData, "system_state")
}

// uploadBackup handles the actual upload to MinIO
func (s *MinIOSink) uploadBackup(ctx context.Context, data BackupData, dataType string) error {
	// Convert data to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup data: %w", err)
	}

	// Get the appropriate bucket for this data type
	bucketName := s.getBucketForDataType(dataType)

	// Create object name with timestamp and data type
	timestamp := data.Timestamp.Format("2006/01/02/15")
	objectName := fmt.Sprintf("%s/%s/%s_%d.json",
		dataType,
		timestamp,
		data.Source,
		data.Timestamp.Unix())

	// Upload to MinIO
	reader := bytes.NewReader(jsonData)
	_, err = s.client.PutObject(ctx, bucketName, objectName, reader, int64(len(jsonData)), minio.PutObjectOptions{
		ContentType: "application/json",
		UserMetadata: map[string]string{
			"data_type":     data.DataType,
			"source":        data.Source,
			"event_count":   fmt.Sprintf("%d", data.EventCount),
			"backup_reason": data.BackupReason,
			"bucket":        bucketName,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to upload backup to MinIO: %w", err)
	}

	s.logger.Printf("✅ Backed up %s data to bucket %s: %s", dataType, bucketName, objectName)
	return nil
}

// ScheduledBackup performs scheduled backup of various data types
func (s *MinIOSink) ScheduledBackup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Println("Stopping scheduled backup routine")
			return
		case <-ticker.C:
			s.logger.Println("Starting scheduled backup routine")

			// Here you would implement logic to gather data from various sources
			// and backup to MinIO. For example:
			// - Backup recent events from processing pipeline
			// - Backup current metrics state
			// - Backup user profile summaries
			// - Backup system configuration

			// This is a placeholder - actual implementation would depend on
			// your specific data access patterns
			s.logger.Println("Scheduled backup completed")
		}
	}
}

// ListBackups lists available backups for a given data type
func (s *MinIOSink) ListBackups(ctx context.Context, dataType string, limit int) ([]string, error) {
	var backups []string

	// Get the appropriate bucket for this data type
	bucketName := s.getBucketForDataType(dataType)

	// List objects with the specified prefix
	objectCh := s.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    dataType + "/",
		Recursive: true,
	})

	count := 0
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects in bucket %s: %w", bucketName, object.Err)
		}

		backups = append(backups, fmt.Sprintf("%s/%s", bucketName, object.Key))
		count++

		if limit > 0 && count >= limit {
			break
		}
	}

	return backups, nil
}

// RestoreBackup retrieves backup data from MinIO
func (s *MinIOSink) RestoreBackup(ctx context.Context, objectPath string) (*BackupData, error) {
	// Parse bucket and object name from path (format: bucket/object)
	parts := strings.SplitN(objectPath, "/", 2)
	var bucketName, objectName string

	if len(parts) == 2 {
		bucketName = parts[0]
		objectName = parts[1]
	} else {
		// Fallback to default bucket if path doesn't include bucket
		bucketName = s.bucketName
		objectName = objectPath
	}

	// Get object from MinIO
	object, err := s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get backup object %s from bucket %s: %w", objectName, bucketName, err)
	}
	defer object.Close()

	// Read the object content
	var data BackupData
	decoder := json.NewDecoder(object)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode backup data: %w", err)
	}

	s.logger.Printf("✅ Restored backup from %s/%s", bucketName, objectName)
	return &data, nil
}

// CleanupOldBackups removes backups older than the specified duration
func (s *MinIOSink) CleanupOldBackups(ctx context.Context, olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan)

	// List all objects in the bucket
	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Recursive: true,
	})

	var toDelete []string
	for object := range objectCh {
		if object.Err != nil {
			return fmt.Errorf("error listing objects for cleanup: %w", object.Err)
		}

		// Check if object is older than cutoff time
		if object.LastModified.Before(cutoffTime) {
			toDelete = append(toDelete, object.Key)
		}
	}

	// Delete old backups
	for _, objectName := range toDelete {
		if err := s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{}); err != nil {
			s.logger.Printf("Failed to delete old backup %s: %v", objectName, err)
		} else {
			s.logger.Printf("Deleted old backup: %s", objectName)
		}
	}

	return nil
}
