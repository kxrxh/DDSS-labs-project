package middleware

import (
	"time"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/kxrxh/social-rating-system/internal/logger" // Adjust import path if needed
	"go.uber.org/zap"
)

// NewJWTMiddleware creates a new JWT middleware handler with custom error handling.
func NewJWTMiddleware(secret []byte) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: secret, JWTAlg: jwtware.HS256},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusUnauthorized // Default to Unauthorized for JWT errors
			message := "Invalid or expired JWT"

			if err.Error() == "Missing or malformed JWT" {
				code = fiber.StatusBadRequest
				message = "Missing or malformed JWT"
			}

			// Log the JWT error
			logger.Log.Warn("JWT Error",
				zap.Error(err),
				zap.String("path", c.Path()),
				zap.Int("status", code),
			)

			// Create standardized error response
			errorResponse := ResponseData{
				Status:    code,
				Message:   message,
				Result:    nil,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Path:      c.Path(),
			}

			// Set content type and send response
			c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			return c.Status(code).JSON(errorResponse)
		},
	})
}
