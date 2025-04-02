package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/kxrxh/ddss/internal/repositories"
	"github.com/kxrxh/ddss/internal/repositories/mongo" // Example for type assertion
)

// HealthHandler checks the health of the service, potentially including DB connections.
func HealthHandler(repos map[string]repositories.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Example: Check MongoDB connection specifically
		if mongoRepo, ok := repos["mongo"].(*mongo.Repository); ok {
			if err := mongoRepo.Client().Ping(c.Context(), nil); err != nil {
				fmt.Printf("Health check failed (Mongo ping): %v\n", err) // Replace with proper logging
				return fiber.NewError(fiber.StatusInternalServerError, "Database health check failed")
			}
			fmt.Println("Health check: Mongo ping successful") // Example log
		} else {
			fmt.Println("Health check warning: Mongo repository not found or type assertion failed")
		}

		// Add checks for other databases if needed...

		return c.SendStatus(fiber.StatusOK)
	}
}
