package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/kxrxh/ddss/internal/api/routes"
	"github.com/kxrxh/ddss/internal/flink" // <-- Import Flink
	"github.com/kxrxh/ddss/internal/repositories"
)

// NewServer creates a new Fiber app, registers routes using the routes package,
// and returns the app instance. It now requires repositories and a Flink client.
// It will panic if route registration fails due to missing required environment variables.
func NewServer(repos map[string]repositories.Repository, flinkClient *flink.Client) *fiber.App { // <-- Added flinkClient
	app := fiber.New()

	// Register all routes, passing the Flink client
	if err := routes.RegisterRoutes(app, repos, flinkClient); err != nil {
		// We panic here because the application cannot function without proper JWT setup
		panic(fmt.Sprintf("Failed to initialize routes: %v", err))
	}

	return app
}
