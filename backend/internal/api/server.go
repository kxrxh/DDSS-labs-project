package api

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/kxrxh/ddss/internal/db"
)

// Server represents the REST API server
type Server struct {
	app    *fiber.App
	addr   string
	client *db.Client
}

// NewServer creates a new REST API server
func NewServer(addr string, client *db.Client) *Server {
	app := fiber.New(fiber.Config{
		AppName:      "DDSS-REST-API",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	})

	// Add middleware
	app.Use(logger.New())
	app.Use(recover.New())

	server := &Server{
		app:    app,
		addr:   addr,
		client: client,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// Start starts the REST API server
func (s *Server) Start() error {
	log.Printf("Starting REST API server on %s\n", s.addr)
	return s.app.Listen(s.addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}

// setupRoutes configures all the routes for the API
func (s *Server) setupRoutes() {
	// API group
	api := s.app.Group("/api")

	// Citizens endpoints
	citizens := api.Group("/citizens")
	citizens.Get("/", s.getAllCitizens)
	citizens.Get("/:id", s.getCitizen)
	citizens.Post("/", s.createCitizen)
	citizens.Put("/:id", s.updateCitizen)

	// Blacklist endpoints
	blacklist := api.Group("/blacklist")
	blacklist.Get("/:id", s.checkBlacklist)
	blacklist.Post("/", s.addToBlacklist)
	blacklist.Delete("/:id", s.removeFromBlacklist)

	// Whitelist endpoints
	whitelist := api.Group("/whitelist")
	whitelist.Get("/:id", s.checkWhitelist)
	whitelist.Post("/", s.addToWhitelist)
	whitelist.Delete("/:id", s.removeFromWhitelist)

	// Relationships endpoints
	relationships := api.Group("/relationships")
	relationships.Get("/:id", s.getRelationships)
	relationships.Post("/", s.createRelationship)

	// Rules endpoints
	rules := api.Group("/rules")
	rules.Get("/", s.getAllRules)
	rules.Get("/:id", s.getRule)
	rules.Post("/", s.createRule)

	// Transactions endpoints
	transactions := api.Group("/transactions")
	transactions.Get("/:id", s.getTransactions)
	transactions.Post("/", s.createTransaction)

	// Zones endpoints
	zones := api.Group("/zones")
	zones.Get("/:citizenId/:zoneId", s.checkZoneAccess)
	zones.Post("/", s.setZoneAccess)

	// Security endpoints
	security := api.Group("/security")
	security.Get("/:id", s.getSecurityClearance)
	security.Post("/", s.setSecurityClearance)

	// Tiers endpoints
	tiers := api.Group("/tiers")
	tiers.Get("/:id", s.getTier)
	tiers.Post("/", s.setTier)

	// Reeducation endpoints
	reeducation := api.Group("/reeducation")
	reeducation.Get("/:id", s.getReeducationStatus)
	reeducation.Post("/", s.setReeducationStatus)
}

// Helper to send JSON error responses
func sendErrorJSON(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": message,
	})
}
