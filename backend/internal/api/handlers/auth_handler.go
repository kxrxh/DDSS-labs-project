package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kxrxh/ddss/internal/models"
	"github.com/kxrxh/ddss/internal/repositories"
	"github.com/kxrxh/ddss/internal/repositories/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// TODO: Move JWT secret to configuration
var jwtSecret = []byte("your_very_secret_key_change_me") // CHANGE THIS!

// RegisterPayload defines the expected input for registration.
type RegisterPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginPayload defines the expected input for login.
type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterHandler handles user registration.
func RegisterHandler(repos map[string]repositories.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		payload := new(RegisterPayload)
		if err := c.BodyParser(payload); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		if payload.Username == "" || payload.Password == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Username and password are required")
		}

		// Get MongoDB repository
		mongoRepo, ok := repos["mongo"].(*mongo.Repository)
		if !ok {
			return fiber.NewError(fiber.StatusInternalServerError, "Database service not available")
		}
		usersCollection := mongoRepo.DB().Collection("users")

		// Check if username already exists
		count, err := usersCollection.CountDocuments(c.Context(), bson.M{"username": payload.Username})
		if err != nil {
			fmt.Printf("Error checking username existence: %v\n", err) // Log error
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to check username")
		}
		if count > 0 {
			return fiber.NewError(fiber.StatusConflict, "Username already taken")
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("Error hashing password: %v\n", err) // Log error
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to process registration")
		}

		// Create new user
		newUser := models.User{
			Username: payload.Username,
			Password: string(hashedPassword),
		}

		// Insert user into database
		_, err = usersCollection.InsertOne(c.Context(), newUser)
		if err != nil {
			fmt.Printf("Error inserting user: %v\n", err) // Log error
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to register user")
		}

		// Don't return password hash in response
		newUser.Password = ""
		return c.Status(fiber.StatusCreated).JSON(newUser)
	}
}

// LoginHandler handles user login and issues a JWT.
func LoginHandler(repos map[string]repositories.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		payload := new(LoginPayload)
		if err := c.BodyParser(payload); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		if payload.Username == "" || payload.Password == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Username and password are required")
		}

		// Get MongoDB repository
		mongoRepo, ok := repos["mongo"].(*mongo.Repository)
		if !ok {
			return fiber.NewError(fiber.StatusInternalServerError, "Database service not available")
		}
		usersCollection := mongoRepo.DB().Collection("users")

		// Find user by username
		var foundUser models.User
		err := usersCollection.FindOne(c.Context(), bson.M{"username": payload.Username}).Decode(&foundUser)
		if err != nil {
			// Could be mongo.ErrNoDocuments, treat as invalid credentials
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid username or password")
		}

		// Compare password hash
		err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(payload.Password))
		if err != nil {
			// Password doesn't match
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid username or password")
		}

		// --- Generate JWT ---
		claims := jwt.MapClaims{
			"sub": foundUser.ID.Hex(), // Subject (user ID)
			"usr": foundUser.Username,
			"exp": time.Now().Add(time.Hour * 72).Unix(), // Expiration time (e.g., 72 hours)
			"iat": time.Now().Unix(),                     // Issued at
		}

		// Create token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Generate encoded token and send it as response.
		t, err := token.SignedString(jwtSecret)
		if err != nil {
			fmt.Printf("Error signing JWT: %v\n", err) // Log error
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to create login token")
		}

		return c.JSON(fiber.Map{"token": t})
	}
}
