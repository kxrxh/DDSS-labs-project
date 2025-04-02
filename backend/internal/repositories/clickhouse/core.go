package clickhouse

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/kxrxh/social-rating-system/internal/repositories"
)

var _ repositories.Repository = (*Repository)(nil)

type Repository struct {
	db *sql.DB
}

// New creates and returns a new Repository instance connected to ClickHouse.
// It takes the Data Source Name (DSN) as input.
func New(dsn string) (*Repository, error) {
	// Open database connection
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	// Set connection pool parameters (optional, adjust as needed)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)

	// Verify the connection
	if err := db.Ping(); err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("failed to ping clickhouse database: %w", err)
	}

	return &Repository{db: db}, nil
}

// Close closes the database connection.
// It's important to call this when the repository is no longer needed.
func (r *Repository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// DB returns the underlying sql.DB object.
// Use this carefully, primarily for operations not covered by repository methods.
func (r *Repository) DB() *sql.DB {
	return r.db
}
