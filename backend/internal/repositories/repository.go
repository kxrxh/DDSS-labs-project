package repositories

// Repository defines the common methods that all database repository implementations should provide.
type Repository interface {
	// Close terminates the connection to the database and cleans up resources.
	Close() error
}
