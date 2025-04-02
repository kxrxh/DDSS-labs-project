package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kxrxh/social-rating-system/internal/api/handlers"
	// "github.com/kxrxh/social-rating-system/internal/repositories" // No longer needed directly
)

// RegisterAuthRoutes sets up the authentication routes (login, register).
func RegisterAuthRoutes(router fiber.Router, deps *handlers.Dependencies) {
	auth := router.Group("/auth")

	auth.Post("/register", handlers.RegisterHandler(deps))
	auth.Post("/login", handlers.LoginHandler(deps))
}
