package models

import "time"

// Graph model definitions for Dgraph

// GraphRelationship represents a relationship between citizens
type GraphRelationship struct {
	FromID                 string                 `json:"from_id"`
	ToID                   string                 `json:"to_id"`
	RelationshipType       string                 `json:"relationship_type"`
	Trust                  float64                `json:"trust"`
	MutualInfluence        float64                `json:"mutual_influence"`
	StartDate              time.Time              `json:"start_date"`
	EndDate                time.Time              `json:"end_date"`
	Interactions           int                    `json:"interactions"`
	LastInteraction        time.Time              `json:"last_interaction"`
	Strength               float64                `json:"strength"`
	StrengthScore          float64                `json:"strength_score"`
	CommunicationFrequency float64                `json:"communication_frequency"`
	Metadata               map[string]interface{} `json:"metadata"`
	IsPrimary              bool                   `json:"is_primary"`
	IsReported             bool                   `json:"is_reported"`
	MonitoringLevel        string                 `json:"monitoring_level"`
	Flagged                bool                   `json:"flagged"`
	ReportCount            int                    `json:"report_count"`
}

// GraphActivity represents a citizen activity in the graph
type GraphActivity struct {
	CitizenID            string    `json:"citizen_id"`
	Timestamp            time.Time `json:"timestamp"`
	Type                 string    `json:"type"`
	Description          string    `json:"description"`
	LocationID           string    `json:"location_id"`
	PointsImpact         int       `json:"points_impact"`
	EvidenceHash         string    `json:"evidence_hash"`
	VerificationStatus   string    `json:"verification_status"`
	AssociatedCitizenIDs []string  `json:"associated_citizen_ids"`
	Flags                []string  `json:"flags"`
}

// GraphRelationshipView represents a relationship from a citizen's perspective
type GraphRelationshipView struct {
	CitizenID          string    `json:"citizen_id"`
	RelatedCitizenID   string    `json:"related_citizen_id"`
	RelatedCitizenName string    `json:"related_citizen_name"`
	RelationshipType   string    `json:"relationship_type"`
	Strength           float64   `json:"strength"`
	Trust              float64   `json:"trust"`
	Influence          float64   `json:"influence"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	Interactions       int       `json:"interactions"`
	LastInteraction    time.Time `json:"last_interaction"`
	Direction          string    `json:"direction"` // "incoming" or "outgoing"
}

// GraphRelationshipPerson represents relationship data about a person
type GraphRelationshipPerson struct {
	UID             string    `json:"uid"`
	SimulatedID     string    `json:"simulatedID"`
	Name            string    `json:"name"`
	Type            string    `json:"type,omitempty"`
	Strength        float64   `json:"strength,omitempty"`
	Trust           float64   `json:"trust,omitempty"`
	Influence       float64   `json:"influence,omitempty"`
	StartDate       time.Time `json:"startDate,omitempty"`
	EndDate         time.Time `json:"endDate,omitempty"`
	Interactions    int       `json:"interactions,omitempty"`
	LastInteraction time.Time `json:"lastInteraction,omitempty"`
}

// GraphSocialPath represents a segment in a social connection path
type GraphSocialPath struct {
	FromID           string    `json:"from_id"`
	FromName         string    `json:"from_name"`
	ToID             string    `json:"to_id"`
	ToName           string    `json:"to_name"`
	RelationshipType string    `json:"relationship_type"`
	Strength         float64   `json:"strength"`
	Trust            float64   `json:"trust"`
	Influence        float64   `json:"influence"`
	StartDate        time.Time `json:"start_date"`
}

// GraphInfluentialCitizen represents a citizen with high social influence
type GraphInfluentialCitizen struct {
	CitizenID            string                  `json:"citizen_id"`
	Name                 string                  `json:"name"`
	InfluenceScore       float64                 `json:"influence_score"`
	RelationshipCount    int                     `json:"relationship_count"`
	IncomingTrustSum     float64                 `json:"incoming_trust_sum"`
	OutgoingCount        int                     `json:"outgoing_count"`
	TopConnectedCitizens []GraphConnectedCitizen `json:"top_connected_citizens"`
}

// GraphConnectedCitizen represents a connected citizen in the influence graph
type GraphConnectedCitizen struct {
	CitizenID string  `json:"citizen_id"`
	Name      string  `json:"name"`
	Influence float64 `json:"influence"`
	Trust     float64 `json:"trust"`
}

// GraphBehavioralPattern represents a citizen's behavioral pattern
type GraphBehavioralPattern struct {
	CitizenID       string    `json:"citizen_id"`
	PatternType     string    `json:"pattern_type"`
	Description     string    `json:"description"`
	Frequency       string    `json:"frequency"`
	FirstObserved   time.Time `json:"first_observed"`
	LastObserved    time.Time `json:"last_observed"`
	Confidence      float64   `json:"confidence"`
	PotentialMotivs []string  `json:"potential_motivations"`
	LocationIDs     []string  `json:"location_ids"`
	AssociatedIDs   []string  `json:"associated_citizen_ids"`
}

// GraphSurveillanceNote represents a surveillance observation
type GraphSurveillanceNote struct {
	CitizenID      string    `json:"citizen_id"`
	Timestamp      time.Time `json:"timestamp"`
	ObserverID     string    `json:"observer_id"`
	LocationID     string    `json:"location_id"`
	Content        string    `json:"content"`
	Classification string    `json:"classification"`
	AttachmentRefs []string  `json:"attachment_refs"`
	SentimentScore float64   `json:"sentiment_score"`
	RiskAssessment string    `json:"risk_assessment"`
}

// GraphTravelEvent represents a citizen travel event
type GraphTravelEvent struct {
	CitizenID          string    `json:"citizen_id"`
	DepartureTime      time.Time `json:"departure_time"`
	ArrivalTime        time.Time `json:"arrival_time"`
	OriginID           string    `json:"origin_id"`
	DestinationID      string    `json:"destination_id"`
	Purpose            string    `json:"purpose"`
	TransportMethod    string    `json:"transport_method"`
	AuthStatus         string    `json:"authorization_status"`
	VerificationStatus string    `json:"verification_status"`
	CompanionIDs       []string  `json:"companion_ids"`
}

// GraphIdeologicalAssessment represents an assessment of citizen ideology
type GraphIdeologicalAssessment struct {
	CitizenID          string    `json:"citizen_id"`
	AssessmentDate     time.Time `json:"assessment_date"`
	AlignmentScore     float64   `json:"alignment_score"`
	DeviationMetrics   []string  `json:"deviation_metrics"`
	ExposureHistory    []string  `json:"exposure_history"`
	CorrectionAttempts int       `json:"correction_attempts"`
	ReeducationStatus  string    `json:"reeducation_status"`
	Trajectory         string    `json:"trajectory"`
}
