package routes

import (
	"fmt"
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/kxrxh/ddss/internal/api/handlers"
	"github.com/kxrxh/ddss/internal/api/middleware"
	"github.com/kxrxh/ddss/internal/flink"
	"github.com/kxrxh/ddss/internal/repositories"
)

// getJWTSecret retrieves the JWT secret from environment variables
func getJWTSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is not set")
	}
	return []byte(secret), nil
}

// RegisterRoutes sets up all the application routes.
func RegisterRoutes(app *fiber.App, repos map[string]repositories.Repository, flinkClient *flink.Client) error {
	// Get JWT secret from environment
	jwtSecret, err := getJWTSecret()
	if err != nil {
		return fmt.Errorf("failed to initialize routes: %v", err)
	}

	// Apply global middleware
	app.Use(middleware.ResponseFormatter())

	// Public routes
	app.Get("/health", handlers.HealthHandler(repos))

	// Auth routes (Login, Register)
	RegisterAuthRoutes(app, repos)

	// TODO: Add routes that use flinkClient here, e.g.:
	// flinkGroup := app.Group("/flink")
	// flinkGroup.Get("/jobs", handlers.GetFlinkJobsHandler(flinkClient))

	// --- JWT Middleware Setup ---
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: jwtSecret, JWTAlg: jwtware.HS256},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if err.Error() == "Missing or malformed JWT" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Missing or malformed JWT",
					"data":    nil,
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid or expired JWT",
				"data":    nil,
			})
		},
	}))

	// --- Protected Routes ---
	app.Get("/items", handlers.GetItemsHandler(repos))

	return nil
}
