package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/kxrxh/ddss/pkg/models"
)

// AddToBlacklist adds a citizen to the blacklist
func (c *Client) AddToBlacklist(citizenID, reason string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixBlacklist + citizenID)
	err = txn.Set(key, []byte(reason))
	if err != nil {
		return fmt.Errorf("failed to set blacklist entry: %w", err)
	}

	return txn.Commit(context.Background())
}

// IsBlacklisted checks if a citizen is on the blacklist
func (c *Client) IsBlacklisted(citizenID string) (bool, string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return false, "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixBlacklist + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return false, "", nil // Not found = not blacklisted
	}
	return true, string(val), nil
}

// AddToWhitelist adds a citizen to the whitelist
func (c *Client) AddToWhitelist(citizenID, reason string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixWhitelist + citizenID)
	err = txn.Set(key, []byte(reason))
	if err != nil {
		return fmt.Errorf("failed to set whitelist entry: %w", err)
	}

	return txn.Commit(context.Background())
}

// IsWhitelisted checks if a citizen is on the whitelist
func (c *Client) IsWhitelisted(citizenID string) (bool, string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return false, "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixWhitelist + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return false, "", nil // Not found = not whitelisted
	}
	return true, string(val), nil
}

// SetCitizenStatus sets a citizen's status in TiKV
func (c *Client) SetCitizenStatus(citizenID, status string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixCitizenStatus + citizenID)
	err = txn.Set(key, []byte(status))
	if err != nil {
		return fmt.Errorf("failed to set citizen status: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetCitizenStatus retrieves a citizen's status from TiKV
func (c *Client) GetCitizenStatus(citizenID string) (string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixCitizenStatus + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return models.StatusActive, nil // Default to active if not found
	}
	return string(val), nil
}

// CacheScore caches a citizen's current score in TiKV for faster access
func (c *Client) CacheScore(citizenID string, score int) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixCurrentScore + citizenID)
	err = txn.Set(key, []byte(strconv.Itoa(score)))
	if err != nil {
		return fmt.Errorf("failed to cache score: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetCachedScore gets a citizen's cached score from TiKV
func (c *Client) GetCachedScore(citizenID string) (int, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixCurrentScore + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return 0, fmt.Errorf("score not cached: %w", err)
	}

	score, err := strconv.Atoi(string(val))
	if err != nil {
		return 0, fmt.Errorf("invalid cached score: %w", err)
	}

	return score, nil
}

// StoreScoreBreakdown stores a citizen's score breakdown by category
func (c *Client) StoreScoreBreakdown(citizenID string, breakdown models.CategoryBreakdown) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	data, err := json.Marshal(breakdown)
	if err != nil {
		return fmt.Errorf("failed to marshal breakdown: %w", err)
	}

	key := []byte(models.PrefixScoreBreakdown + citizenID)
	err = txn.Set(key, data)
	if err != nil {
		return fmt.Errorf("failed to store score breakdown: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetScoreBreakdown retrieves a citizen's score breakdown
func (c *Client) GetScoreBreakdown(citizenID string) (models.CategoryBreakdown, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return models.CategoryBreakdown{}, fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixScoreBreakdown + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return models.CategoryBreakdown{}, fmt.Errorf("breakdown not found: %w", err)
	}

	var breakdown models.CategoryBreakdown
	err = json.Unmarshal(val, &breakdown)
	if err != nil {
		return models.CategoryBreakdown{}, fmt.Errorf("failed to unmarshal breakdown: %w", err)
	}

	return breakdown, nil
}

// SetScoreTrajectory sets a citizen's score trajectory
func (c *Client) SetScoreTrajectory(citizenID, trajectory string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixScoreTrajectory + citizenID)
	err = txn.Set(key, []byte(trajectory))
	if err != nil {
		return fmt.Errorf("failed to set trajectory: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetScoreTrajectory gets a citizen's score trajectory
func (c *Client) GetScoreTrajectory(citizenID string) (string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixScoreTrajectory + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return models.TrajectoryStable, nil // Default to stable if not found
	}

	return string(val), nil
}

// SetCitizenTier sets a citizen's tier classification
func (c *Client) SetCitizenTier(citizenID, tier string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixTier + citizenID)
	err = txn.Set(key, []byte(tier))
	if err != nil {
		return fmt.Errorf("failed to set tier: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetCitizenTier gets a citizen's tier classification
func (c *Client) GetCitizenTier(citizenID string) (string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixTier + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return models.TierC, nil // Default to C tier if not found
	}

	return string(val), nil
}

// SetSecurityClearance sets a citizen's security clearance level
func (c *Client) SetSecurityClearance(citizenID, clearance string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixSecurityClearance + citizenID)
	err = txn.Set(key, []byte(clearance))
	if err != nil {
		return fmt.Errorf("failed to set security clearance: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetSecurityClearance gets a citizen's security clearance level
func (c *Client) GetSecurityClearance(citizenID string) (string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixSecurityClearance + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return models.ClearanceLow, nil // Default to low clearance if not found
	}

	return string(val), nil
}

// UpdateLastActivity updates a citizen's last activity timestamp
func (c *Client) UpdateLastActivity(citizenID string, activity string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	data := struct {
		Timestamp time.Time `json:"timestamp"`
		Activity  string    `json:"activity"`
	}{
		Timestamp: time.Now(),
		Activity:  activity,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal activity data: %w", err)
	}

	key := []byte(models.PrefixLastActivity + citizenID)
	err = txn.Set(key, jsonData)
	if err != nil {
		return fmt.Errorf("failed to update last activity: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetLastActivity gets a citizen's last recorded activity
func (c *Client) GetLastActivity(citizenID string) (time.Time, string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return time.Time{}, "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixLastActivity + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("no activity record found: %w", err)
	}

	var data struct {
		Timestamp time.Time `json:"timestamp"`
		Activity  string    `json:"activity"`
	}

	err = json.Unmarshal(val, &data)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("failed to unmarshal activity data: %w", err)
	}

	return data.Timestamp, data.Activity, nil
}

// UpdateLastLocation updates a citizen's last known location
func (c *Client) UpdateLastLocation(citizenID string, latitude, longitude float64, locationName string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	data := struct {
		Timestamp  time.Time `json:"timestamp"`
		Latitude   float64   `json:"latitude"`
		Longitude  float64   `json:"longitude"`
		LocationID string    `json:"location_id"`
	}{
		Timestamp:  time.Now(),
		Latitude:   latitude,
		Longitude:  longitude,
		LocationID: locationName,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal location data: %w", err)
	}

	key := []byte(models.PrefixLastLocation + citizenID)
	err = txn.Set(key, jsonData)
	if err != nil {
		return fmt.Errorf("failed to update last location: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetLastLocation gets a citizen's last known location
func (c *Client) GetLastLocation(citizenID string) (time.Time, float64, float64, string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return time.Time{}, 0, 0, "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixLastLocation + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return time.Time{}, 0, 0, "", fmt.Errorf("no location record found: %w", err)
	}

	var data struct {
		Timestamp  time.Time `json:"timestamp"`
		Latitude   float64   `json:"latitude"`
		Longitude  float64   `json:"longitude"`
		LocationID string    `json:"location_id"`
	}

	err = json.Unmarshal(val, &data)
	if err != nil {
		return time.Time{}, 0, 0, "", fmt.Errorf("failed to unmarshal location data: %w", err)
	}

	return data.Timestamp, data.Latitude, data.Longitude, data.LocationID, nil
}

// SetSurveillancePriority sets a citizen's surveillance priority level
func (c *Client) SetSurveillancePriority(citizenID, priority string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixSurveillancePriority + citizenID)
	err = txn.Set(key, []byte(priority))
	if err != nil {
		return fmt.Errorf("failed to set surveillance priority: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetSurveillancePriority gets a citizen's surveillance priority level
func (c *Client) GetSurveillancePriority(citizenID string) (string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixSurveillancePriority + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return models.SurvPriorityLow, nil // Default to low priority if not found
	}

	return string(val), nil
}

// SetIdeologyAlignment sets a citizen's ideological alignment trajectory
func (c *Client) SetIdeologyAlignment(citizenID, trajectory string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixIdeologyAlignment + citizenID)
	err = txn.Set(key, []byte(trajectory))
	if err != nil {
		return fmt.Errorf("failed to set ideology alignment: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetIdeologyAlignment gets a citizen's ideological alignment trajectory
func (c *Client) GetIdeologyAlignment(citizenID string) (string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixIdeologyAlignment + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return models.IdeologyTrajectoryStable, nil // Default to stable if not found
	}

	return string(val), nil
}

// SetComplianceStreak sets a citizen's compliance streak count
func (c *Client) SetComplianceStreak(citizenID string, streak int) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixComplianceStreak + citizenID)
	err = txn.Set(key, []byte(strconv.Itoa(streak)))
	if err != nil {
		return fmt.Errorf("failed to set compliance streak: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetComplianceStreak gets a citizen's compliance streak count
func (c *Client) GetComplianceStreak(citizenID string) (int, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixComplianceStreak + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return 0, nil // Default to 0 if not found
	}

	streak, err := strconv.Atoi(string(val))
	if err != nil {
		return 0, fmt.Errorf("invalid streak value: %w", err)
	}

	return streak, nil
}

// IncrementDailyTransactionCount increments the daily transaction count for a citizen
func (c *Client) IncrementDailyTransactionCount(citizenID string, date time.Time) (int, error) {
	dateStr := date.Format("2006-01-02")
	txn, err := c.tikv.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixDailyTxCount + citizenID + ":" + dateStr)
	val, err := txn.Get(context.Background(), key)

	var count int
	if err == nil {
		// Key exists, increment
		count, err = strconv.Atoi(string(val))
		if err != nil {
			return 0, fmt.Errorf("invalid count value: %w", err)
		}
		count++
	} else {
		// Key doesn't exist, start at 1
		count = 1
	}

	err = txn.Set(key, []byte(strconv.Itoa(count)))
	if err != nil {
		return 0, fmt.Errorf("failed to set transaction count: %w", err)
	}

	err = txn.Commit(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

// GetDeviceAssociation gets the citizen associated with a device
func (c *Client) GetDeviceAssociation(deviceType, deviceID string) (string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	var key []byte
	if deviceType == "MAC" {
		key = []byte(models.PrefixDeviceMAC + deviceID)
	} else if deviceType == "IMEI" {
		key = []byte(models.PrefixDeviceIMEI + deviceID)
	} else {
		return "", fmt.Errorf("unsupported device type: %s", deviceType)
	}

	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return "", fmt.Errorf("device not associated: %w", err)
	}

	return string(val), nil
}

// AssociateDevice associates a device with a citizen
func (c *Client) AssociateDevice(deviceType, deviceID, citizenID string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	var key []byte
	if deviceType == "MAC" {
		key = []byte(models.PrefixDeviceMAC + deviceID)
	} else if deviceType == "IMEI" {
		key = []byte(models.PrefixDeviceIMEI + deviceID)
	} else {
		return fmt.Errorf("unsupported device type: %s", deviceType)
	}

	err = txn.Set(key, []byte(citizenID))
	if err != nil {
		return fmt.Errorf("failed to associate device: %w", err)
	}

	return txn.Commit(context.Background())
}

// SetReeducationStatus sets a citizen's reeducation status
func (c *Client) SetReeducationStatus(citizenID, status string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixReeducation + citizenID)
	err = txn.Set(key, []byte(status))
	if err != nil {
		return fmt.Errorf("failed to set reeducation status: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetReeducationStatus gets a citizen's reeducation status
func (c *Client) GetReeducationStatus(citizenID string) (string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixReeducation + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return "", fmt.Errorf("no reeducation status found: %w", err)
	}

	return string(val), nil
}

// SetNextAssessmentDate sets a citizen's next assessment date
func (c *Client) SetNextAssessmentDate(citizenID string, date time.Time) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixNextAssessment + citizenID)
	dateStr := date.Format(time.RFC3339)
	err = txn.Set(key, []byte(dateStr))
	if err != nil {
		return fmt.Errorf("failed to set next assessment date: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetNextAssessmentDate gets a citizen's next assessment date
func (c *Client) GetNextAssessmentDate(citizenID string) (time.Time, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixNextAssessment + citizenID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return time.Time{}, fmt.Errorf("no next assessment date found: %w", err)
	}

	date, err := time.Parse(time.RFC3339, string(val))
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %w", err)
	}

	return date, nil
}

// SetZoneAccess sets a citizen's access permission for a zone
func (c *Client) SetZoneAccess(citizenID, zoneID string, permitted bool) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixZoneAccess + citizenID + ":" + zoneID)
	var value string
	if permitted {
		value = "allowed"
	} else {
		value = "denied"
	}

	err = txn.Set(key, []byte(value))
	if err != nil {
		return fmt.Errorf("failed to set zone access: %w", err)
	}

	return txn.Commit(context.Background())
}

// CheckZoneAccess checks if a citizen has access to a zone
func (c *Client) CheckZoneAccess(citizenID, zoneID string) (bool, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return false, fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixZoneAccess + citizenID + ":" + zoneID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return false, nil // Default to no access if not found
	}

	return string(val) == "allowed", nil
}

// ResetCache clears cached data for a citizen
func (c *Client) ResetCache(citizenID string) error {
	txn, err := c.tikv.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	// Generate a new cache token (timestamp) to invalidate cache
	key := []byte(models.PrefixCacheToken + citizenID)
	token := fmt.Sprintf("%d", time.Now().UnixNano())
	err = txn.Set(key, []byte(token))
	if err != nil {
		return fmt.Errorf("failed to reset cache: %w", err)
	}

	return txn.Commit(context.Background())
}

// GenerateAlert creates a new alert in the system
func (c *Client) GenerateAlert(alertType, citizenID, message string, severity int) (string, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	alertID := fmt.Sprintf("%s_%d", citizenID, time.Now().UnixNano())

	alert := struct {
		AlertID   string    `json:"alert_id"`
		Type      string    `json:"type"`
		CitizenID string    `json:"citizen_id"`
		Message   string    `json:"message"`
		Severity  int       `json:"severity"`
		Timestamp time.Time `json:"timestamp"`
		Processed bool      `json:"processed"`
	}{
		AlertID:   alertID,
		Type:      alertType,
		CitizenID: citizenID,
		Message:   message,
		Severity:  severity,
		Timestamp: time.Now(),
		Processed: false,
	}

	data, err := json.Marshal(alert)
	if err != nil {
		return "", fmt.Errorf("failed to marshal alert: %w", err)
	}

	key := []byte(models.PrefixAlert + alertID)
	err = txn.Set(key, data)
	if err != nil {
		return "", fmt.Errorf("failed to store alert: %w", err)
	}

	err = txn.Commit(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return alertID, nil
}

// GetAlert retrieves an alert by ID
func (c *Client) GetAlert(alertID string) (map[string]interface{}, error) {
	txn, err := c.tikv.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin txn: %w", err)
	}
	defer txn.Rollback()

	key := []byte(models.PrefixAlert + alertID)
	val, err := txn.Get(context.Background(), key)
	if err != nil {
		return nil, fmt.Errorf("alert not found: %w", err)
	}

	var alert map[string]interface{}
	err = json.Unmarshal(val, &alert)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal alert: %w", err)
	}

	return alert, nil
}
