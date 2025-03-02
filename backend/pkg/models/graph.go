package models

import (
	"time"
)

// DgraphModels contains all Dgraph model definitions

// Citizen represents a citizen node in Dgraph
type GraphCitizen struct {
	UID                string               `json:"uid,omitempty"`
	Type               string               `json:"dgraph.type,omitempty"`
	SimulatedID        string               `json:"simulated_id,omitempty"`
	Name               string               `json:"name,omitempty"`
	Birthdate          time.Time            `json:"birthdate,omitempty"`
	Gender             string               `json:"gender,omitempty"`
	NationalID         string               `json:"national_id,omitempty"`
	BiometricIDs       []*BiometricID       `json:"biometric_identifiers,omitempty"`
	SurveillanceProf   *SurveillanceProfile `json:"surveillance_profile,omitempty"`
	CurrentScore       int                  `json:"current_score,omitempty"`
	Tier               string               `json:"tier,omitempty"`
	Residence          *Location            `json:"residence,omitempty"`
	Employment         []*Employment        `json:"employment,omitempty"`
	Education          []*Education         `json:"education,omitempty"`
	Relationships      []*Relationship      `json:"relationships,omitempty"`
	SocialGroups       []*SocialGroup       `json:"socialGroups,omitempty"`
	Transactions       []*GraphTransaction  `json:"transactions,omitempty"`
	Flags              []*Flag              `json:"flags,omitempty"`
	InfluenceScore     float64              `json:"influenceScore,omitempty"`
	BehavioralPatterns []*BehavioralPattern `json:"behavioralPatterns,omitempty"`
	TravelHistory      []*TravelEvent       `json:"travelHistory,omitempty"`
	DigitalFootprint   *DigitalFootprint    `json:"digitalFootprint,omitempty"`
	ConsumptionProfile *ConsumptionProfile  `json:"consumptionProfile,omitempty"`
	IdeologicalProfile *IdeologicalProfile  `json:"ideologicalAlignment,omitempty"`
	ComplianceHist     *ComplianceHistory   `json:"complianceHistory,omitempty"`
}

// Location represents a physical location
type Location struct {
	UID                  string          `json:"uid,omitempty"`
	Type                 string          `json:"dgraph.type,omitempty"`
	ID                   string          `json:"id,omitempty"`
	Address              string          `json:"address,omitempty"`
	City                 string          `json:"city,omitempty"`
	Province             string          `json:"province,omitempty"`
	Country              string          `json:"country,omitempty"`
	PostalCode           string          `json:"postal_code,omitempty"`
	Coordinates          string          `json:"coordinates,omitempty"` // geo format
	ReputationScore      float64         `json:"reputation_score,omitempty"`
	Citizens             []*GraphCitizen `json:"citizens,omitempty"`
	SurveillanceLevel    int             `json:"surveillance_level,omitempty"`
	RestrictedAccess     bool            `json:"restricted_access,omitempty"`
	AccessRequirements   []string        `json:"access_requirements,omitempty"`
	TypicalVisitors      []*GraphCitizen `json:"typical_visitors,omitempty"`
	LastSecurityIncident time.Time       `json:"last_security_incident,omitempty"`
	ReputationCategory   string          `json:"reputation_category,omitempty"`
	ExcludedCitizens     []*GraphCitizen `json:"excluded_citizens,omitempty"`
	PermittedCitizens    []*GraphCitizen `json:"permitted_citizens,omitempty"`
}

// Employment represents employment information
type Employment struct {
	UID                string        `json:"uid,omitempty"`
	Type               string        `json:"dgraph.type,omitempty"`
	ID                 string        `json:"id,omitempty"`
	Employer           string        `json:"employer,omitempty"`
	Position           string        `json:"position,omitempty"`
	StartDate          time.Time     `json:"start_date,omitempty"`
	EndDate            time.Time     `json:"end_date,omitempty"`
	VerificationStatus string        `json:"verification_status,omitempty"`
	ComplianceScore    float64       `json:"compliance_score,omitempty"`
	Employee           *GraphCitizen `json:"employee,omitempty"`
}

// Education represents educational background
type Education struct {
	UID                string        `json:"uid,omitempty"`
	Type               string        `json:"dgraph.type,omitempty"`
	ID                 string        `json:"id,omitempty"`
	Institution        string        `json:"institution,omitempty"`
	Degree             string        `json:"degree,omitempty"`
	Field              string        `json:"field,omitempty"`
	StartDate          time.Time     `json:"start_date,omitempty"`
	EndDate            time.Time     `json:"end_date,omitempty"`
	VerificationStatus string        `json:"verification_status,omitempty"`
	AcademicScore      float64       `json:"academic_score,omitempty"`
	Student            *GraphCitizen `json:"student,omitempty"`
}

// Relationship represents a relationship between citizens
type Relationship struct {
	UID             string          `json:"uid,omitempty"`
	Type            string          `json:"dgraph.type,omitempty"`
	ID              string          `json:"id,omitempty"`
	RelType         string          `json:"type,omitempty"`
	Strength        float64         `json:"strength,omitempty"`
	StartDate       time.Time       `json:"start_date,omitempty"`
	Status          string          `json:"status,omitempty"`
	TrustScore      float64         `json:"trust_score,omitempty"`
	MutualInfluence float64         `json:"mutual_influence,omitempty"`
	Participants    []*GraphCitizen `json:"participants,omitempty"`
}

// SocialGroup represents a group of citizens
type SocialGroup struct {
	UID             string          `json:"uid,omitempty"`
	Type            string          `json:"dgraph.type,omitempty"`
	ID              string          `json:"id,omitempty"`
	Name            string          `json:"name,omitempty"`
	GroupType       string          `json:"type,omitempty"`
	FormationDate   time.Time       `json:"formation_date,omitempty"`
	Status          string          `json:"status,omitempty"`
	ReputationScore float64         `json:"reputation_score,omitempty"`
	InfluenceRadius float64         `json:"influence_radius,omitempty"`
	Members         []*GraphCitizen `json:"members,omitempty"`
	ParentGroup     *SocialGroup    `json:"parent_group,omitempty"`
	Subgroups       []*SocialGroup  `json:"subgroups,omitempty"`
}

// GraphTransaction represents a transaction in Dgraph
type GraphTransaction struct {
	UID             string        `json:"uid,omitempty"`
	Type            string        `json:"dgraph.type,omitempty"`
	ID              string        `json:"id,omitempty"`
	Timestamp       time.Time     `json:"timestamp,omitempty"`
	TransactionType string        `json:"type,omitempty"`
	PointsChange    int           `json:"points_change,omitempty"`
	Description     string        `json:"description,omitempty"`
	Rule            *Rule         `json:"rule,omitempty"`
	Citizen         *GraphCitizen `json:"citizen,omitempty"`
	Location        string        `json:"location,omitempty"` // geo format
	EvidenceID      string        `json:"evidence_id,omitempty"`
}

// GraphRule represents a rule in Dgraph
type GraphRule struct {
	UID          string              `json:"uid,omitempty"`
	Type         string              `json:"dgraph.type,omitempty"`
	ID           string              `json:"id,omitempty"`
	RuleID       string              `json:"rule_id,omitempty"`
	Name         string              `json:"name,omitempty"`
	Description  string              `json:"description,omitempty"`
	Category     string              `json:"category,omitempty"`
	BaseImpact   int                 `json:"base_impact,omitempty"`
	Active       bool                `json:"active,omitempty"`
	Transactions []*GraphTransaction `json:"transactions,omitempty"`
}

// Flag represents a flag applied to citizens
type Flag struct {
	UID            string          `json:"uid,omitempty"`
	Type           string          `json:"dgraph.type,omitempty"`
	ID             string          `json:"id,omitempty"`
	Name           string          `json:"name,omitempty"`
	FlagType       string          `json:"type,omitempty"`
	Severity       int             `json:"severity,omitempty"`
	CreationDate   time.Time       `json:"creation_date,omitempty"`
	ExpirationDate time.Time       `json:"expiration_date,omitempty"`
	Description    string          `json:"description,omitempty"`
	Citizens       []*GraphCitizen `json:"citizens,omitempty"`
}

// BiometricID represents biometric identification
type BiometricID struct {
	UID               string        `json:"uid,omitempty"`
	Type              string        `json:"dgraph.type,omitempty"`
	ID                string        `json:"id,omitempty"`
	BiometricType     string        `json:"type,omitempty"`
	HashValue         string        `json:"hash_value,omitempty"`
	ConfidenceScore   float64       `json:"confidence_score,omitempty"`
	LastUpdated       time.Time     `json:"last_updated,omitempty"`
	VerificationCount int           `json:"verification_count,omitempty"`
	Citizen           *GraphCitizen `json:"citizen,omitempty"`
}

// SurveillanceProfile represents surveillance settings for a citizen
type SurveillanceProfile struct {
	UID              string              `json:"uid,omitempty"`
	Type             string              `json:"dgraph.type,omitempty"`
	ID               string              `json:"id,omitempty"`
	RiskLevel        string              `json:"risk_level,omitempty"`
	SurveillanceTier int                 `json:"surveillance_tier,omitempty"`
	MonitoringFreq   string              `json:"monitoring_frequency,omitempty"`
	AssignedAlgs     []string            `json:"assigned_algorithms,omitempty"`
	OverrideAuth     string              `json:"override_authority,omitempty"`
	Citizen          *GraphCitizen       `json:"citizen,omitempty"`
	ExclusionZones   []*Location         `json:"exclusion_zones,omitempty"`
	PermittedZones   []*Location         `json:"permitted_zones,omitempty"`
	SurvNotes        []*SurveillanceNote `json:"surveillance_notes,omitempty"`
}

// SurveillanceNote represents a note about surveillance
type SurveillanceNote struct {
	UID            string               `json:"uid,omitempty"`
	Type           string               `json:"dgraph.type,omitempty"`
	ID             string               `json:"id,omitempty"`
	Timestamp      time.Time            `json:"timestamp,omitempty"`
	ObserverID     string               `json:"observer_id,omitempty"`
	Content        string               `json:"content,omitempty"`
	Classification string               `json:"classification,omitempty"`
	Attachments    []string             `json:"attachments,omitempty"`
	Profile        *SurveillanceProfile `json:"profile,omitempty"`
}

// BehavioralPattern represents a citizen's behavior pattern
type BehavioralPattern struct {
	UID             string              `json:"uid,omitempty"`
	Type            string              `json:"dgraph.type,omitempty"`
	ID              string              `json:"id,omitempty"`
	PatternType     string              `json:"pattern_type,omitempty"`
	Confidence      float64             `json:"confidence,omitempty"`
	FirstObserved   time.Time           `json:"first_observed,omitempty"`
	LastObserved    time.Time           `json:"last_observed,omitempty"`
	Frequency       string              `json:"frequency,omitempty"`
	Locations       []*Location         `json:"locations,omitempty"`
	AssociatedCitis []*GraphCitizen     `json:"associated_citizens,omitempty"`
	Description     string              `json:"description,omitempty"`
	PotentialMotivs []string            `json:"potential_motivations,omitempty"`
	Citizen         *GraphCitizen       `json:"citizen,omitempty"`
	PatternComps    []*PatternComponent `json:"pattern_components,omitempty"`
}

// PatternComponent represents a component of a behavioral pattern
type PatternComponent struct {
	UID             string             `json:"uid,omitempty"`
	Type            string             `json:"dgraph.type,omitempty"`
	ID              string             `json:"id,omitempty"`
	ActivityType    string             `json:"activity_type,omitempty"`
	TypicalTime     time.Time          `json:"typical_time,omitempty"`
	TypicalDuration int                `json:"typical_duration,omitempty"`
	TypicalLocation *Location          `json:"typical_location,omitempty"`
	Reliability     float64            `json:"reliability,omitempty"`
	ParentPattern   *BehavioralPattern `json:"parent_pattern,omitempty"`
}

// TravelEvent represents a citizen's travel event
type TravelEvent struct {
	UID             string          `json:"uid,omitempty"`
	Type            string          `json:"dgraph.type,omitempty"`
	ID              string          `json:"id,omitempty"`
	DepartureTime   time.Time       `json:"departure_time,omitempty"`
	ArrivalTime     time.Time       `json:"arrival_time,omitempty"`
	Origin          *Location       `json:"origin,omitempty"`
	Destination     *Location       `json:"destination,omitempty"`
	Purpose         string          `json:"purpose,omitempty"`
	AuthStatus      string          `json:"authorization_status,omitempty"`
	TransportMethod string          `json:"transportation_method,omitempty"`
	Companions      []*GraphCitizen `json:"companions,omitempty"`
	VerifyStatus    string          `json:"verification_status,omitempty"`
	Citizen         *GraphCitizen   `json:"citizen,omitempty"`
}

// DigitalFootprint represents a citizen's online presence
type DigitalFootprint struct {
	UID              string                   `json:"uid,omitempty"`
	Type             string                   `json:"dgraph.type,omitempty"`
	ID               string                   `json:"id,omitempty"`
	OnlineHandles    []string                 `json:"online_handles,omitempty"`
	SentimentAnal    float64                  `json:"sentiment_analysis,omitempty"`
	PlatformPresence []*GraphPlatformPresence `json:"platform_presence,omitempty"`
	ContentCategs    []string                 `json:"content_categories,omitempty"`
	InfluenceSphere  []*GraphCitizen          `json:"influence_sphere,omitempty"`
	SuspConns        []*GraphCitizen          `json:"suspicious_connections,omitempty"`
	EncryptionUsage  float64                  `json:"encryption_usage,omitempty"`
	BrowsingPatterns string                   `json:"browsing_patterns,omitempty"`
	Citizen          *GraphCitizen            `json:"citizen,omitempty"`
}

// PlatformPresence represents presence on a specific platform
type GraphPlatformPresence struct {
	UID              string            `json:"uid,omitempty"`
	Type             string            `json:"dgraph.type,omitempty"`
	ID               string            `json:"id,omitempty"`
	PlatformName     string            `json:"platform_name,omitempty"`
	AccountHandle    string            `json:"account_handle,omitempty"`
	VerifyStatus     string            `json:"verification_status,omitempty"`
	FollowerCount    int               `json:"follower_count,omitempty"`
	InfluenceScore   float64           `json:"influence_score,omitempty"`
	SentimentScore   float64           `json:"sentiment_score,omitempty"`
	ActivityFreq     float64           `json:"activity_frequency,omitempty"`
	ContentAnalysis  string            `json:"content_analysis,omitempty"`
	DigitalFootprint *DigitalFootprint `json:"digital_footprint,omitempty"`
}

// ConsumptionProfile represents a citizen's consumption behavior
type ConsumptionProfile struct {
	UID              string                 `json:"uid,omitempty"`
	Type             string                 `json:"dgraph.type,omitempty"`
	ID               string                 `json:"id,omitempty"`
	LuxuryScore      float64                `json:"luxury_score,omitempty"`
	NecessityScore   float64                `json:"necessity_score,omitempty"`
	ResourceEffic    float64                `json:"resource_efficiency,omitempty"`
	ConsumptCategs   []*ConsumptionCategory `json:"consumption_categories,omitempty"`
	PurchasePatterns []*PurchasingPattern   `json:"purchasing_patterns,omitempty"`
	Citizen          *GraphCitizen          `json:"citizen,omitempty"`
}

// ConsumptionCategory represents a category of consumption
type ConsumptionCategory struct {
	UID            string              `json:"uid,omitempty"`
	Type           string              `json:"dgraph.type,omitempty"`
	ID             string              `json:"id,omitempty"`
	CategoryName   string              `json:"category_name,omitempty"`
	SpendingLevel  float64             `json:"spending_level,omitempty"`
	ApprovedStatus string              `json:"approved_status,omitempty"`
	MonthlyAlloc   float64             `json:"monthly_allocation,omitempty"`
	MonthlyActual  float64             `json:"monthly_actual,omitempty"`
	Profile        *ConsumptionProfile `json:"profile,omitempty"`
}

// PurchasingPattern represents purchasing habits
type PurchasingPattern struct {
	UID             string              `json:"uid,omitempty"`
	Type            string              `json:"dgraph.type,omitempty"`
	ID              string              `json:"id,omitempty"`
	PatternName     string              `json:"pattern_name,omitempty"`
	Frequency       string              `json:"frequency,omitempty"`
	TypicalVendors  []string            `json:"typical_vendors,omitempty"`
	TypicalAmounts  []float64           `json:"typical_amounts,omitempty"`
	ItemsOfInterest []string            `json:"items_of_interest,omitempty"`
	Profile         *ConsumptionProfile `json:"profile,omitempty"`
}

// IdeologicalProfile represents a citizen's ideological alignment
type IdeologicalProfile struct {
	UID              string        `json:"uid,omitempty"`
	Type             string        `json:"dgraph.type,omitempty"`
	ID               string        `json:"id,omitempty"`
	AlignmentScore   float64       `json:"alignment_score,omitempty"`
	DeviationMetrics []string      `json:"deviation_metrics,omitempty"`
	ExposureHistory  []string      `json:"exposure_history,omitempty"`
	CorrectionAtt    int           `json:"correction_attempts,omitempty"`
	ReeducationStat  string        `json:"reeducation_status,omitempty"`
	IdeologTraj      string        `json:"ideological_trajectory,omitempty"`
	Citizen          *GraphCitizen `json:"citizen,omitempty"`
}

// ComplianceHistory represents a citizen's compliance history
type ComplianceHistory struct {
	UID              string               `json:"uid,omitempty"`
	Type             string               `json:"dgraph.type,omitempty"`
	ID               string               `json:"id,omitempty"`
	OverallCompScore float64              `json:"overall_compliance_score,omitempty"`
	HistCompliance   []*ComplianceEvent   `json:"historical_compliance,omitempty"`
	ComplianceStreak int                  `json:"compliance_streak,omitempty"`
	LastViolation    time.Time            `json:"last_violation,omitempty"`
	InterventHist    []*InterventionEvent `json:"intervention_history,omitempty"`
	Citizen          *GraphCitizen        `json:"citizen,omitempty"`
}

// ComplianceEvent represents a compliance or violation event
type ComplianceEvent struct {
	UID             string             `json:"uid,omitempty"`
	Type            string             `json:"dgraph.type,omitempty"`
	ID              string             `json:"id,omitempty"`
	EventDate       time.Time          `json:"event_date,omitempty"`
	EventType       string             `json:"event_type,omitempty"`
	RuleReference   string             `json:"rule_reference,omitempty"`
	ImpactOnScore   int                `json:"impact_on_score,omitempty"`
	Location        *Location          `json:"location,omitempty"`
	AssociatedCitis []*GraphCitizen    `json:"associated_citizens,omitempty"`
	EvidenceRefs    []string           `json:"evidence_references,omitempty"`
	History         *ComplianceHistory `json:"history,omitempty"`
}

// InterventionEvent represents a corrective intervention
type InterventionEvent struct {
	UID              string             `json:"uid,omitempty"`
	Type             string             `json:"dgraph.type,omitempty"`
	ID               string             `json:"id,omitempty"`
	InterventionDate time.Time          `json:"intervention_date,omitempty"`
	InterventionType string             `json:"intervention_type,omitempty"`
	Authority        string             `json:"authority,omitempty"`
	Duration         int                `json:"duration,omitempty"`
	Outcome          string             `json:"outcome,omitempty"`
	ComplianceHist   *ComplianceHistory `json:"compliance_history,omitempty"`
}
