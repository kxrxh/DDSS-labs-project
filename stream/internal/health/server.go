package health

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Server represents a health check server
type Server struct {
	server        *http.Server
	pipeline      PipelineHealthChecker
	backupManager BackupManager
}

// PipelineHealthChecker interface for checking pipeline health
type PipelineHealthChecker interface {
	IsRunning() bool
	GetPipelineStatistics() map[string]interface{}
}

// BackupManager interface for backup operations
type BackupManager interface {
	BackupSystemSnapshot(ctx context.Context) error
	GetBackupStatistics(ctx context.Context) (map[string]interface{}, error)
	ListRecentBackups(ctx context.Context, dataType string, limit int) ([]string, error)
}

// NewServer creates a new health server
func NewServer(addr string, pipeline PipelineHealthChecker) *Server {
	s := &Server{
		pipeline:      pipeline,
		backupManager: nil, // Will be set later
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ready", s.readyHandler)
	mux.HandleFunc("/stats", s.statsHandler)
	mux.HandleFunc("/backup/trigger", s.triggerBackupHandler)
	mux.HandleFunc("/backup/stats", s.backupStatsHandler)
	mux.HandleFunc("/backup/list", s.listBackupsHandler)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.server = server
	return s
}

// SetBackupManager sets the backup manager for the health server
func (s *Server) SetBackupManager(manager BackupManager) {
	s.backupManager = manager
}

// healthHandler handles basic health checks
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// readyHandler checks if the service is ready
func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	if !s.pipeline.IsRunning() {
		http.Error(w, "Not Ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

// statsHandler returns pipeline statistics
func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := s.pipeline.GetPipelineStatistics()
	stats["timestamp"] = time.Now()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Failed to encode stats", http.StatusInternalServerError)
		return
	}
}

// triggerBackupHandler triggers a manual backup
func (s *Server) triggerBackupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.backupManager == nil {
		http.Error(w, "Backup manager not available", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.backupManager.BackupSystemSnapshot(ctx); err != nil {
		http.Error(w, "Backup failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":    "success",
		"message":   "Backup triggered successfully",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// backupStatsHandler returns backup statistics
func (s *Server) backupStatsHandler(w http.ResponseWriter, r *http.Request) {
	if s.backupManager == nil {
		http.Error(w, "Backup manager not available", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats, err := s.backupManager.GetBackupStatistics(ctx)
	if err != nil {
		http.Error(w, "Failed to get backup stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// listBackupsHandler lists recent backups
func (s *Server) listBackupsHandler(w http.ResponseWriter, r *http.Request) {
	if s.backupManager == nil {
		http.Error(w, "Backup manager not available", http.StatusServiceUnavailable)
		return
	}

	dataType := r.URL.Query().Get("type")
	if dataType == "" {
		dataType = "processed_events"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	backups, err := s.backupManager.ListRecentBackups(ctx, dataType, 10)
	if err != nil {
		http.Error(w, "Failed to list backups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data_type": dataType,
		"backups":   backups,
		"count":     len(backups),
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Start starts the health check server
func (s *Server) Start() {
	log.Printf("Health check server starting on %s", s.server.Addr)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /health - Basic health check")
	log.Printf("  GET  /ready - Readiness check")
	log.Printf("  GET  /stats - Pipeline statistics")
	log.Printf("  POST /backup/trigger - Trigger manual backup")
	log.Printf("  GET  /backup/stats - Backup statistics")
	log.Printf("  GET  /backup/list?type=<datatype> - List recent backups")

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Health server error: %v", err)
	}
}

// Stop stops the health check server
func (s *Server) Stop() error {
	return s.server.Close()
}
