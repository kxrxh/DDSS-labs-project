package middleware

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kxrxh/social-rating-system/internal/logger"
	"go.uber.org/zap"
)

// NewCustomErrorHandler creates a new Fiber ErrorHandler function
// that logs errors and returns standardized JSON responses using the ResponseData format.
func NewCustomErrorHandler() func(ctx *fiber.Ctx, err error) error {
	return func(ctx *fiber.Ctx, err error) error {
		// Default error code
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		// Retrieve the custom error type if possible
		var e *fiber.Error
		if errors.As(err, &e) {
			code = e.Code
			message = e.Message
		}

		// Log the error using zap logger, adding method and status
		logger.Log.Error("Unhandled error",
			zap.Error(err),
			zap.String("method", ctx.Method()),
			zap.String("path", ctx.Path()),
			zap.Int("status", code),
		)

		// Create standardized error response using ResponseData structure
		// Assumes ResponseData is defined in this package or imported correctly
		errorResponse := ResponseData{
			Status:    code,
			Message:   message,
			Result:    nil, // No result data for errors
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Path:      ctx.Path(),
		}
		
		ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return ctx.Status(code).JSON(errorResponse)
	}
}
