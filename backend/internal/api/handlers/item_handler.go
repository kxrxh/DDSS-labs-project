package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	// "github.com/kxrxh/social-rating-system/internal/repositories" // No longer needed directly
	"github.com/kxrxh/social-rating-system/internal/repositories/mongo" // Example for type assertion
)

// GetItemsHandler handles requests to retrieve items.
func GetItemsHandler(deps *Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Access a specific repository from dependencies
		mongoRepo, ok := deps.Repos["mongo"].(*mongo.Repository)
		if !ok {
			return fiber.NewError(fiber.StatusInternalServerError, "Required database service (mongo) not available")
		}

		fmt.Printf("Accessed mongo repository in GetItemsHandler: %T\n", mongoRepo)

		// --- Perform database operation ---
		// Replace with actual logic to fetch items from mongoRepo.DB()
		// ... database query logic ...
		// --- End database operation ---

		// Placeholder response
		items := []string{"item1-from-handler", "item2-from-handler"}
		return c.JSON(items)
	}
}

// Add more item-related handlers here (e.g., CreateItemHandler, GetItemByIDHandler, etc.)
