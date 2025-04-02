package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kxrxh/ddss/internal/api/handlers"
	"github.com/kxrxh/ddss/internal/repositories"
)

// RegisterAuthRoutes sets up the authentication routes (login, register).
func RegisterAuthRoutes(router fiber.Router, repos map[string]repositories.Repository) {
	auth := router.Group("/auth")

	auth.Post("/register", handlers.RegisterHandler(repos))
	auth.Post("/login", handlers.LoginHandler(repos))
}
