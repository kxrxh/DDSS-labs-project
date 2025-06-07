package health

import (
	"log"
	"net/http"
)

// Server represents a health check server
type Server struct {
	server *http.Server
}

// PipelineHealthChecker interface for checking pipeline health
type PipelineHealthChecker interface {
	IsRunning() bool
}

// NewServer creates a new health server
func NewServer(addr string, pipeline PipelineHealthChecker) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Check if pipeline is running
		if !pipeline.IsRunning() {
			http.Error(w, "Not Ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &Server{server: server}
}

// Start starts the health check server
func (s *Server) Start() {
	log.Printf("Health check server starting on %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Health server error: %v", err)
	}
}

// Stop stops the health check server
func (s *Server) Stop() error {
	return s.server.Close()
}
