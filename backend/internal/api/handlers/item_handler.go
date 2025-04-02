package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/kxrxh/ddss/internal/repositories"
	"github.com/kxrxh/ddss/internal/repositories/mongo" // Example for type assertion
)

// GetItemsHandler handles requests to retrieve items.
func GetItemsHandler(repos map[string]repositories.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Access a specific repository, e.g., MongoDB
		mongoRepo, ok := repos["mongo"].(*mongo.Repository)
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
