package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/kxrxh/social-rating-system/internal/api/handlers"   // Import handlers for Dependencies struct
	"github.com/kxrxh/social-rating-system/internal/api/middleware" // Import middleware package
	"github.com/kxrxh/social-rating-system/internal/api/routes"
	"github.com/kxrxh/social-rating-system/internal/flink" // <-- Import Flink
	"github.com/kxrxh/social-rating-system/internal/repositories"
)

// NewServer creates and configures a new Fiber app.
func NewServer(repos map[string]repositories.Repository, flinkClient *flink.Client) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:   "Social Rating System App v1.0",
		BodyLimit: 4 * 1024 * 1024, // 4MB request body limit
		// Prefork: true, // Enable prefork for potential performance gains on multi-core systems.
		ErrorHandler: middleware.NewCustomErrorHandler(),
	})

	// Create Dependencies struct
	deps := &handlers.Dependencies{
		Repos:       repos,
		FlinkClient: flinkClient,
	}

	// Register all routes, passing the Dependencies struct
	if err := routes.RegisterRoutes(app, deps); err != nil {
		// We panic here because the application cannot function without proper JWT setup
		panic(fmt.Sprintf("Failed to initialize routes: %v", err))
	}

	return app
}
