package utils

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// getUserClaims retrieves the JWT claims from the Fiber context.
// Assumes the JWT middleware stores the token in c.Locals("user").
func getUserClaims(c *fiber.Ctx) (jwt.MapClaims, error) {
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok || token == nil {
		return nil, fmt.Errorf("JWT token not found in context")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims format")
	}

	return claims, nil
}

// GetUserIDFromContext extracts the user ID (from 'sub' claim) from the JWT in the context.
func GetUserIDFromContext(c *fiber.Ctx) (string, error) {
	claims, err := getUserClaims(c)
	if err != nil {
		return "", err
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("user ID ('sub') claim not found or not a string")
	}

	return sub, nil
}

// GetUsernameFromContext extracts the username (from 'usr' claim) from the JWT in the context.
func GetUsernameFromContext(c *fiber.Ctx) (string, error) {
	claims, err := getUserClaims(c)
	if err != nil {
		return "", err
	}

	usr, ok := claims["usr"].(string)
	if !ok {
		return "", fmt.Errorf("username ('usr') claim not found or not a string")
	}

	return usr, nil
}
