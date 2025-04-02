package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// ResponseData represents the standard API response structure
type ResponseData struct {
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	Result    interface{} `json:"result"`
	Timestamp string      `json:"timestamp"`
	Path      string      `json:"path"`
}

// ResponseFormatter is a middleware that formats all API responses
func ResponseFormatter() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Process the request first
		err := c.Next()
		if err != nil {
			return err
		}

		// Get the status code and body from the response
		statusCode := c.Response().StatusCode()
		body := c.Response().Body()

		// Create the formatted response
		response := ResponseData{
			Status:    statusCode,
			Message:   getStatusMessage(statusCode),
			Result:    body,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Path:      c.Path(),
		}

		// Send the formatted response
		return c.JSON(response)
	}
}

// getStatusMessage returns a human-readable message for HTTP status codes
func getStatusMessage(status int) string {
	switch status {
	case fiber.StatusOK:
		return "Success"
	case fiber.StatusCreated:
		return "Created"
	case fiber.StatusBadRequest:
		return "Bad Request"
	case fiber.StatusUnauthorized:
		return "Unauthorized"
	case fiber.StatusForbidden:
		return "Forbidden"
	case fiber.StatusNotFound:
		return "Not Found"
	case fiber.StatusInternalServerError:
		return "Internal Server Error"
	default:
		return "Unknown Status"
	}
}
