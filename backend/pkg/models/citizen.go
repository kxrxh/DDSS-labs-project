package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MongoModels contains all MongoDB model definitions

// Citizen represents a citizen in the MongoDB database
type Citizen struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	SimulatedID       string             `bson:"simulated_id"`
	PersonalInfo      PersonalInfo       `bson:"personal_info"`
	CreditProfile     CreditProfile      `bson:"credit_profile"`
	ActivityLog       []ActivityRecord   `bson:"activity_log"`
	ComplianceMetrics ComplianceMetrics  `bson:"compliance_metrics"`
	AccessPermissions AccessPermissions  `bson:"access_permissions"`
}

// PersonalInfo contains citizen personal information
type PersonalInfo struct {
	Name              string             `bson:"name"`
	Birthdate         time.Time          `bson:"birthdate"`
	Gender            string             `bson:"gender"`
	NationalID        string             `bson:"national_id"`
	BiometricData     BiometricData      `bson:"biometric_data"`
	Contact           Contact            `bson:"contact"`
	FamilyConnections []FamilyConnection `bson:"family_connections"`
}

// BiometricData contains biometric identification information
type BiometricData struct {
	FingerprintHash     string    `bson:"fingerprint_hash"`
	FacialRecognitionID string    `bson:"facial_recognition_id"`
	LastVerified        time.Time `bson:"last_verified"`
}

// Contact contains contact information
type Contact struct {
	Email             string             `bson:"email"`
	Phone             string             `bson:"phone"`
	Address           Address            `bson:"address"`
	EmergencyContacts []EmergencyContact `bson:"emergency_contacts"`
}

// Address contains detailed address information
type Address struct {
	Street         string         `bson:"street"`
	City           string         `bson:"city"`
	Province       string         `bson:"province"`
	PostalCode     string         `bson:"postal_code"`
	Country        string         `bson:"country"`
	GeoCoordinates GeoCoordinates `bson:"geo_coordinates"`
}

// GeoCoordinates contains geographic coordinates
type GeoCoordinates struct {
	Latitude    float64   `bson:"latitude"`
	Longitude   float64   `bson:"longitude"`
	Precision   float64   `bson:"precision"`
	LastUpdated time.Time `bson:"last_updated"`
}

// EmergencyContact contains emergency contact information
type EmergencyContact struct {
	Name         string `bson:"name"`
	Relationship string `bson:"relationship"`
	Phone        string `bson:"phone"`
	Verified     bool   `bson:"verified"`
}

// FamilyConnection represents a family relationship
type FamilyConnection struct {
	RelativeID         string  `bson:"relative_id"`
	RelationshipType   string  `bson:"relationship_type"`
	InfluenceFactor    float64 `bson:"influence_factor"`
	CohabitationStatus bool    `bson:"cohabitation_status"`
}

// CreditProfile contains credit score information
type CreditProfile struct {
	CurrentScore      int               `bson:"current_score"`
	ScoreHistory      []ScoreHistory    `bson:"score_history"`
	Tier              string            `bson:"tier"`
	TierHistory       []TierHistory     `bson:"tier_history"`
	Flags             []string          `bson:"flags"`
	LastUpdated       time.Time         `bson:"last_updated"`
	PredictiveMetrics PredictiveMetrics `bson:"predictive_metrics"`
}

// ScoreHistory represents a historical score record
type ScoreHistory struct {
	Date              time.Time         `bson:"date"`
	Score             int               `bson:"score"`
	Change            int               `bson:"change"`
	CategoryBreakdown CategoryBreakdown `bson:"category_breakdown"`
}

// CategoryBreakdown breaks down scores by category
type CategoryBreakdown struct {
	Financial     int `bson:"financial"`
	Social        int `bson:"social"`
	Civic         int `bson:"civic"`
	Environmental int `bson:"environmental"`
}

// TierHistory tracks changes in credit tier
type TierHistory struct {
	OldTier    string    `bson:"old_tier"`
	NewTier    string    `bson:"new_tier"`
	ChangeDate time.Time `bson:"change_date"`
	Reason     string    `bson:"reason"`
}

// PredictiveMetrics provides score predictions
type PredictiveMetrics struct {
	TrendDirection string         `bson:"trend_direction"`
	ProjectedScore ProjectedScore `bson:"projected_score"`
	RiskFactors    []RiskFactor   `bson:"risk_factors"`
}

// ProjectedScore contains score projections
type ProjectedScore struct {
	SixMonths  int     `bson:"six_months"`
	OneYear    int     `bson:"one_year"`
	Confidence float64 `bson:"confidence"`
}

// RiskFactor represents a risk that could impact credit score
type RiskFactor struct {
	Factor          string  `bson:"factor"`
	Probability     float64 `bson:"probability"`
	PotentialImpact int     `bson:"potential_impact"`
}

// ActivityRecord represents a logged activity
type ActivityRecord struct {
	Timestamp          time.Time           `bson:"timestamp"`
	ActivityType       string              `bson:"activity_type"`
	Description        string              `bson:"description"`
	Location           ActivityLocation    `bson:"location"`
	PointsImpact       int                 `bson:"points_impact"`
	RuleID             string              `bson:"rule_id"`
	EvidenceData       EvidenceData        `bson:"evidence_data"`
	AssociatedCitizens []AssociatedCitizen `bson:"associated_citizens"`
}

// ActivityLocation represents where an activity occurred
type ActivityLocation struct {
	Coordinates        []float64 `bson:"coordinates"` // [longitude, latitude]
	Place              string    `bson:"place"`
	Verified           bool      `bson:"verified"`
	VerificationMethod string    `bson:"verification_method"`
}

// EvidenceData contains information about activity evidence
type EvidenceData struct {
	EvidenceType    string `bson:"evidence_type"`
	SourceDevice    string `bson:"source_device"`
	IntegrityHash   string `bson:"integrity_hash"`
	StorageLocation string `bson:"storage_location"`
}

// AssociatedCitizen represents another citizen involved in an activity
type AssociatedCitizen struct {
	CitizenID string `bson:"citizen_id"`
	Role      string `bson:"role"`
}

// ComplianceMetrics tracks citizen compliance across categories
type ComplianceMetrics struct {
	Financial     FinancialCompliance     `bson:"financial"`
	Social        SocialCompliance        `bson:"social"`
	Legal         LegalCompliance         `bson:"legal"`
	Health        HealthCompliance        `bson:"health"`
	Environmental EnvironmentalCompliance `bson:"environmental"`
}

// FinancialCompliance tracks financial behavior
type FinancialCompliance struct {
	TaxCompliance     float64          `bson:"tax_compliance"`
	DebtRatio         float64          `bson:"debt_ratio"`
	PaymentTimeliness float64          `bson:"payment_timeliness"`
	FinancialHistory  []FinancialEvent `bson:"financial_history"`
}

// FinancialEvent represents a financial transaction or event
type FinancialEvent struct {
	EventType string                `bson:"event_type"`
	Date      time.Time             `bson:"date"`
	Impact    float64               `bson:"impact"`
	Details   FinancialEventDetails `bson:"details"`
}

// FinancialEventDetails contains specifics about financial events
type FinancialEventDetails struct {
	Amount      float64 `bson:"amount"`
	Institution string  `bson:"institution"`
	AccountType string  `bson:"account_type"`
}

// SocialCompliance tracks social behavior
type SocialCompliance struct {
	CommunityParticipation float64         `bson:"community_participation"`
	OnlineBehavior         float64         `bson:"online_behavior"`
	RelationshipStatus     float64         `bson:"relationship_status"`
	SocialInfluence        SocialInfluence `bson:"social_influence"`
}

// SocialInfluence tracks social media presence and influence
type SocialInfluence struct {
	FollowerCount    int                `bson:"follower_count"`
	SentimentScore   float64            `bson:"sentiment_score"`
	ReachMultiplier  float64            `bson:"reach_multiplier"`
	PlatformPresence []PlatformPresence `bson:"platform_presence"`
}

// PlatformPresence represents presence on a social platform
type PlatformPresence struct {
	Platform       string  `bson:"platform"`
	Handle         string  `bson:"handle"`
	Verified       bool    `bson:"verified"`
	InfluenceScore float64 `bson:"influence_score"`
}

// LegalCompliance tracks legal behavior
type LegalCompliance struct {
	ViolationsCount  int               `bson:"violations_count"`
	SeverityIndex    float64           `bson:"severity_index"`
	ViolationHistory []ViolationRecord `bson:"violation_history"`
}

// ViolationRecord represents a legal violation
type ViolationRecord struct {
	ViolationType    string    `bson:"violation_type"`
	Date             time.Time `bson:"date"`
	Severity         float64   `bson:"severity"`
	ResolutionStatus string    `bson:"resolution_status"`
	EnforcingAgency  string    `bson:"enforcing_agency"`
}

// HealthCompliance tracks health-related behavior
type HealthCompliance struct {
	WellnessScore          float64             `bson:"wellness_score"`
	ComplianceWithCheckups float64             `bson:"compliance_with_checkups"`
	SubstanceUseMetrics    SubstanceUseMetrics `bson:"substance_use_metrics"`
	ExerciseCompliance     float64             `bson:"exercise_compliance"`
}

// SubstanceUseMetrics tracks substance usage
type SubstanceUseMetrics struct {
	Alcohol             float64 `bson:"alcohol"`
	Tobacco             float64 `bson:"tobacco"`
	MonitoredSubstances float64 `bson:"monitored_substances"`
}

// EnvironmentalCompliance tracks environmental behavior
type EnvironmentalCompliance struct {
	CarbonFootprint     float64             `bson:"carbon_footprint"`
	RecyclingCompliance float64             `bson:"recycling_compliance"`
	ResourceConsumption ResourceConsumption `bson:"resource_consumption"`
}

// ResourceConsumption tracks resource usage
type ResourceConsumption struct {
	Water       float64 `bson:"water"`
	Electricity float64 `bson:"electricity"`
	FossilFuels float64 `bson:"fossil_fuels"`
}

// AccessPermissions tracks access control settings
type AccessPermissions struct {
	RestrictedAreas   []RestrictedArea  `bson:"restricted_areas"`
	TravelPermissions TravelPermissions `bson:"travel_permissions"`
}

// RestrictedArea represents an area with controlled access
type RestrictedArea struct {
	AreaID            string    `bson:"area_id"`
	AccessLevel       string    `bson:"access_level"`
	GrantedDate       time.Time `bson:"granted_date"`
	ExpirationDate    time.Time `bson:"expiration_date"`
	GrantingAuthority string    `bson:"granting_authority"`
}

// TravelPermissions contains travel permission settings
type TravelPermissions struct {
	Domestic          bool     `bson:"domestic"`
	International     bool     `bson:"international"`
	RestrictedRegions []string `bson:"restricted_regions"`
	ApprovedRegions   []string `bson:"approved_regions"`
}

// Rule represents a social credit rule in MongoDB
type Rule struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	RuleID      string             `bson:"rule_id"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	Category    string             `bson:"category"`
	Subcategory string             `bson:"subcategory"`
	PointImpact PointImpact        `bson:"point_impact"`
	Conditions  RuleConditions     `bson:"conditions"`
	Enforcement RuleEnforcement    `bson:"enforcement"`
	Metadata    RuleMetadata       `bson:"metadata"`
}

// PointImpact defines how a rule impacts credit score
type PointImpact struct {
	BaseValue         int                `bson:"base_value"`
	MultiplierFactors []MultiplierFactor `bson:"multiplier_factors"`
	MaxImpact         int                `bson:"max_impact"`
}

// MultiplierFactor represents a factor that can multiply score impact
type MultiplierFactor struct {
	FactorName string  `bson:"factor_name"`
	Threshold  float64 `bson:"threshold"`
	Multiplier float64 `bson:"multiplier"`
}

// RuleConditions defines when a rule applies
type RuleConditions struct {
	Prerequisites []string `bson:"prerequisites"`
	Exclusions    []string `bson:"exclusions"`
}

// RuleEnforcement defines how a rule is enforced
type RuleEnforcement struct {
	Automatic      bool `bson:"automatic"`
	RequiresReview bool `bson:"requires_review"`
	ReviewLevel    int  `bson:"review_level"`
}

// RuleMetadata contains metadata about a rule
type RuleMetadata struct {
	CreatedAt  time.Time `bson:"created_at"`
	ModifiedAt time.Time `bson:"modified_at"`
	Version    string    `bson:"version"`
	Active     bool      `bson:"active"`
}

// Transaction represents a credit score transaction in MongoDB
type Transaction struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	TransactionID string             `bson:"transaction_id"`
	CitizenID     string             `bson:"citizen_id"`
	RuleID        string             `bson:"rule_id"`
	Timestamp     time.Time          `bson:"timestamp"`
	PointsChange  int                `bson:"points_change"`
	PreviousScore int                `bson:"previous_score"`
	NewScore      int                `bson:"new_score"`
	Reason        string             `bson:"reason"`
	Evidence      Evidence           `bson:"evidence"`
	ProcessedBy   ProcessingEntity   `bson:"processed_by"`
	Status        string             `bson:"status"`
}

// Evidence contains evidence supporting a transaction
type Evidence struct {
	Type         string   `bson:"type"`
	Source       string   `bson:"source"`
	ReferenceIDs []string `bson:"reference_ids"`
	Attachments  []string `bson:"attachments"`
}

// ProcessingEntity represents who/what processed a transaction
type ProcessingEntity struct {
	SystemID   string `bson:"system_id"`
	OperatorID string `bson:"operator_id"`
	Method     string `bson:"method"`
}
