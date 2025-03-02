package api

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kxrxh/ddss/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

// CITIZENS HANDLERS

// getAllCitizens handles GET /api/citizens
func (s *Server) getAllCitizens(c *fiber.Ctx) error {
	query := make(bson.M)
	limit := 100 // Default limit

	// Parse query parameters for filtering
	if c.Query("limit") != "" {
		if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
			limit = l
		}
	}

	citizens, err := s.client.SearchCitizens(query, limit)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to retrieve citizens: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(citizens)
}

// getCitizen handles GET /api/citizens/:id
func (s *Server) getCitizen(c *fiber.Ctx) error {
	citizenID := c.Params("id")

	citizen, err := s.client.GetCitizenByID(citizenID)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusNotFound, "Citizen not found: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(citizen)
}

// createCitizen handles POST /api/citizens
func (s *Server) createCitizen(c *fiber.Ctx) error {
	var citizen models.Citizen

	if err := c.BodyParser(&citizen); err != nil {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Invalid citizen data: "+err.Error())
	}

	// Generate a SimulatedID if not provided
	if citizen.SimulatedID == "" {
		// Generate a simple ID using the current timestamp and a random component
		// This is a simple approach - in production, use a proper UUID library
		timeComponent := strconv.FormatInt(time.Now().UnixNano(), 10)
		randomComponent := strconv.Itoa(int(time.Now().Unix() % 10000))
		citizen.SimulatedID = "CIT-" + timeComponent[len(timeComponent)-6:] + "-" + randomComponent
	}

	// Add the citizen to MongoDB
	err := s.client.AddCitizenProfile(citizen)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to add citizen: "+err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": citizen.SimulatedID,
	})
}

// updateCitizen handles PUT /api/citizens/:id
func (s *Server) updateCitizen(c *fiber.Ctx) error {
	citizenID := c.Params("id")
	var citizen models.Citizen

	if err := c.BodyParser(&citizen); err != nil {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Invalid citizen data: "+err.Error())
	}

	// Ensure the ID matches
	citizen.SimulatedID = citizenID

	// Update the citizen
	err := s.client.UpdateCitizenProfile(citizen)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to update citizen: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "updated",
	})
}

// deleteCitizen handles DELETE /api/citizens/:id
func (s *Server) deleteCitizen(c *fiber.Ctx) error {
	citizenID := c.Params("id")

	// Delete the citizen from the database
	err := s.client.DeleteCitizen(citizenID)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to delete citizen: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "deleted",
	})
}

// BLACKLIST HANDLERS

// checkBlacklist handles GET /api/blacklist/:id
func (s *Server) checkBlacklist(c *fiber.Ctx) error {
	citizenID := c.Params("id")

	isBlacklisted, reason, err := s.client.IsBlacklisted(citizenID)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to check blacklist: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"citizen_id":  citizenID,
		"blacklisted": isBlacklisted,
		"reason":      reason,
	})
}

// addToBlacklist handles POST /api/blacklist
func (s *Server) addToBlacklist(c *fiber.Ctx) error {
	var data struct {
		CitizenID string `json:"citizen_id"`
		Reason    string `json:"reason"`
	}

	if err := c.BodyParser(&data); err != nil {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Invalid blacklist data: "+err.Error())
	}

	if data.CitizenID == "" {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Citizen ID is required")
	}

	err := s.client.AddToBlacklist(data.CitizenID, data.Reason)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to add to blacklist: "+err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": "added to blacklist",
	})
}

// removeFromBlacklist handles DELETE /api/blacklist/:id
func (s *Server) removeFromBlacklist(c *fiber.Ctx) error {
	citizenID := c.Params("id")

	// Use AddToBlacklist with empty reason string to effectively delete
	err := s.client.AddToBlacklist(citizenID, "")
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to remove from blacklist: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "removed from blacklist",
	})
}

// WHITELIST HANDLERS

// checkWhitelist handles GET /api/whitelist/:id
func (s *Server) checkWhitelist(c *fiber.Ctx) error {
	citizenID := c.Params("id")

	isWhitelisted, reason, err := s.client.IsWhitelisted(citizenID)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to check whitelist: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"citizen_id":  citizenID,
		"whitelisted": isWhitelisted,
		"reason":      reason,
	})
}

// addToWhitelist handles POST /api/whitelist
func (s *Server) addToWhitelist(c *fiber.Ctx) error {
	var data struct {
		CitizenID string `json:"citizen_id"`
		Reason    string `json:"reason"`
	}

	if err := c.BodyParser(&data); err != nil {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Invalid whitelist data: "+err.Error())
	}

	if data.CitizenID == "" {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Citizen ID is required")
	}

	err := s.client.AddToWhitelist(data.CitizenID, data.Reason)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to add to whitelist: "+err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": "added to whitelist",
	})
}

// removeFromWhitelist handles DELETE /api/whitelist/:id
func (s *Server) removeFromWhitelist(c *fiber.Ctx) error {
	citizenID := c.Params("id")

	// Use AddToWhitelist with empty reason string to effectively delete
	err := s.client.AddToWhitelist(citizenID, "")
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to remove from whitelist: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "removed from whitelist",
	})
}

// RELATIONSHIPS HANDLERS

// getRelationships handles GET /api/relationships/:id
func (s *Server) getRelationships(c *fiber.Ctx) error {
	citizenID := c.Params("id")

	relationships, err := s.client.GetCitizenRelationships(citizenID)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to get relationships: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(relationships)
}

// createRelationship handles POST /api/relationships
func (s *Server) createRelationship(c *fiber.Ctx) error {
	var relationship models.GraphRelationship

	if err := c.BodyParser(&relationship); err != nil {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Invalid relationship data: "+err.Error())
	}

	// Set default values if not provided
	if relationship.StartDate.IsZero() {
		relationship.StartDate = time.Now()
	}
	if relationship.LastInteraction.IsZero() {
		relationship.LastInteraction = time.Now()
	}

	err := s.client.CreateSocialRelationship(relationship)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to create relationship: "+err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status": "relationship created",
	})
}

// RULES HANDLERS

// getAllRules handles GET /api/rules
func (s *Server) getAllRules(c *fiber.Ctx) error {
	query := make(bson.M)
	limit := 100 // Default limit

	// Parse query parameters for filtering
	if c.Query("limit") != "" {
		if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
			limit = l
		}
	}

	rules, err := s.client.GetAllRules(query, limit)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to retrieve rules: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(rules)
}

// getRule handles GET /api/rules/:id
func (s *Server) getRule(c *fiber.Ctx) error {
	ruleID := c.Params("id")

	rule, err := s.client.GetRuleByID(ruleID)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusNotFound, "Rule not found: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(rule)
}

// createRule handles POST /api/rules
func (s *Server) createRule(c *fiber.Ctx) error {
	var rule models.Rule

	if err := c.BodyParser(&rule); err != nil {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Invalid rule data: "+err.Error())
	}

	// Generate a RuleID if not provided
	if rule.RuleID == "" {
		// Generate a simple ID using the current timestamp and a random component
		timeComponent := strconv.FormatInt(time.Now().UnixNano(), 10)
		randomComponent := strconv.Itoa(int(time.Now().Unix() % 10000))
		rule.RuleID = "RULE-" + timeComponent[len(timeComponent)-6:] + "-" + randomComponent
	}

	err := s.client.AddRule(rule)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to add rule: "+err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": rule.RuleID,
	})
}

// TRANSACTIONS HANDLERS

// getTransactions handles GET /api/transactions/:id
func (s *Server) getTransactions(c *fiber.Ctx) error {
	citizenID := c.Params("id")
	limit := 100 // Default limit

	// Parse query parameters
	if c.Query("limit") != "" {
		if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
			limit = l
		}
	}

	transactions, err := s.client.GetCitizenTransactions(citizenID, limit)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to get transactions: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(transactions)
}

// createTransaction handles POST /api/transactions
func (s *Server) createTransaction(c *fiber.Ctx) error {
	var transaction models.Transaction

	if err := c.BodyParser(&transaction); err != nil {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Invalid transaction data: "+err.Error())
	}

	if transaction.Timestamp.IsZero() {
		transaction.Timestamp = time.Now()
	}

	err := s.client.RecordTransaction(transaction)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to record transaction: "+err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": transaction.TransactionID,
	})
}

// ZONES HANDLERS

// checkZoneAccess handles GET /api/zones/:citizenId/:zoneId
func (s *Server) checkZoneAccess(c *fiber.Ctx) error {
	citizenID := c.Params("citizenId")
	zoneID := c.Params("zoneId")

	hasAccess, err := s.client.CheckZoneAccess(citizenID, zoneID)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to check zone access: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"citizen_id": citizenID,
		"zone_id":    zoneID,
		"has_access": hasAccess,
	})
}

// setZoneAccess handles POST /api/zones
func (s *Server) setZoneAccess(c *fiber.Ctx) error {
	var data struct {
		CitizenID string `json:"citizen_id"`
		ZoneID    string `json:"zone_id"`
		Permitted bool   `json:"permitted"`
	}

	if err := c.BodyParser(&data); err != nil {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Invalid zone access data: "+err.Error())
	}

	if data.CitizenID == "" || data.ZoneID == "" {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Both citizen_id and zone_id are required")
	}

	err := s.client.SetZoneAccess(data.CitizenID, data.ZoneID, data.Permitted)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to set zone access: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "zone access updated",
	})
}

// SECURITY HANDLERS

// getSecurityClearance handles GET /api/security/:id
func (s *Server) getSecurityClearance(c *fiber.Ctx) error {
	citizenID := c.Params("id")

	clearance, err := s.client.GetSecurityClearance(citizenID)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to get security clearance: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"citizen_id": citizenID,
		"clearance":  clearance,
	})
}

// setSecurityClearance handles POST /api/security
func (s *Server) setSecurityClearance(c *fiber.Ctx) error {
	var data struct {
		CitizenID string `json:"citizen_id"`
		Clearance string `json:"clearance"`
	}

	if err := c.BodyParser(&data); err != nil {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Invalid security clearance data: "+err.Error())
	}

	if data.CitizenID == "" || data.Clearance == "" {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Both citizen_id and clearance are required")
	}

	err := s.client.SetSecurityClearance(data.CitizenID, data.Clearance)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to set security clearance: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "security clearance updated",
	})
}

// TIERS HANDLERS

// getTier handles GET /api/tiers/:id
func (s *Server) getTier(c *fiber.Ctx) error {
	citizenID := c.Params("id")

	tier, err := s.client.GetCitizenTier(citizenID)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to get tier: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"citizen_id": citizenID,
		"tier":       tier,
	})
}

// setTier handles POST /api/tiers
func (s *Server) setTier(c *fiber.Ctx) error {
	var data struct {
		CitizenID string `json:"citizen_id"`
		Tier      string `json:"tier"`
	}

	if err := c.BodyParser(&data); err != nil {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Invalid tier data: "+err.Error())
	}

	if data.CitizenID == "" || data.Tier == "" {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Both citizen_id and tier are required")
	}

	err := s.client.SetCitizenTier(data.CitizenID, data.Tier)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to set tier: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "tier updated",
	})
}

// REEDUCATION HANDLERS

// getReeducationStatus handles GET /api/reeducation/:id
func (s *Server) getReeducationStatus(c *fiber.Ctx) error {
	citizenID := c.Params("id")

	status, err := s.client.GetReeducationStatus(citizenID)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to get reeducation status: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"citizen_id": citizenID,
		"status":     status,
	})
}

// setReeducationStatus handles POST /api/reeducation
func (s *Server) setReeducationStatus(c *fiber.Ctx) error {
	var data struct {
		CitizenID string `json:"citizen_id"`
		Status    string `json:"status"`
	}

	if err := c.BodyParser(&data); err != nil {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Invalid reeducation data: "+err.Error())
	}

	if data.CitizenID == "" || data.Status == "" {
		return sendErrorJSON(c, fiber.StatusBadRequest, "Both citizen_id and status are required")
	}

	err := s.client.SetReeducationStatus(data.CitizenID, data.Status)
	if err != nil {
		return sendErrorJSON(c, fiber.StatusInternalServerError, "Failed to set reeducation status: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "reeducation status updated",
	})
}
