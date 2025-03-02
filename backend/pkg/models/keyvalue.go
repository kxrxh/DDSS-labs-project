package models

// TiKV constants for key prefixes
const (
	// Blocklist/Whitelist
	PrefixBlacklist = "blacklist:"
	PrefixWhitelist = "whitelist:"

	// Citizen Status
	PrefixCitizenStatus = "citizen_status:"

	// Scores
	PrefixCurrentScore    = "current_score:"
	PrefixScoreBreakdown  = "score_breakdown:"
	PrefixScoreTrajectory = "score_trajectory:"

	// Classifications
	PrefixTier              = "tier:"
	PrefixSecurityClearance = "security_clearance:"

	// Activity Tracking
	PrefixLastActivity = "last_activity:"
	PrefixLastLocation = "last_location:"

	// Device Management
	PrefixDeviceMAC  = "device:MAC_"
	PrefixDeviceIMEI = "device:IMEI_"

	// Biometric Data
	PrefixBiometricFacial      = "biometric:facial:"
	PrefixBiometricFingerprint = "biometric:fingerprint:"
	PrefixBiometricVoice       = "biometric:voice:"

	// System Configuration
	PrefixConfig = "config:"

	// Rate Limiting
	PrefixRateLimit = "rate_limit:api_calls:"

	// Feature Flags
	PrefixFeatureEnabled = "feature_enabled:"

	// Cache Management
	PrefixCacheToken = "cache_token:"

	// Emergency Actions
	PrefixEmergencyOverride = "emergency_override:"

	// Travel Restrictions
	PrefixGeoRestriction = "geo_restriction:"

	// Access Control
	PrefixZoneAccess = "zone_access:"
	PrefixPOIAccess  = "poi_access:"

	// Surveillance
	PrefixSurveillancePriority = "surveillance_priority:"

	// Behavioral Analysis
	PrefixBehaviorPattern = "behavior_pattern:"

	// Ideological Monitoring
	PrefixIdeologyAlignment = "ideology_alignment:"

	// Compliance Tracking
	PrefixComplianceStreak = "compliance_streak:"

	// Transaction Tracking
	PrefixDailyTxCount = "daily_tx_count:"

	// Authentication
	PrefixAuth = "auth:"

	// Notification Settings
	PrefixNotificationChannels = "notification_channels:"

	// Alerts
	PrefixAlert = "alert:"

	// Social Graph Analysis
	PrefixAssociationNetwork = "association_network:"

	// Assessment Scheduling
	PrefixNextAssessment = "next_assessment:"

	// Resource Usage
	PrefixResourceUsage = "resource_usage:"

	// Reeducation Programs
	PrefixReeducation = "reeducation:"

	// Content Analysis
	PrefixContentAnalysis = "content_analysis:"
)

// Status values
const (
	StatusActive    = "active"
	StatusProbation = "probation"
	StatusSuspended = "suspended"
	StatusWatchlist = "watchlist"
)

// Tier values
const (
	TierAPlus = "A+"
	TierA     = "A"
	TierB     = "B"
	TierC     = "C"
	TierD     = "D"
	TierF     = "F"
)

// Security clearance levels
const (
	ClearanceHigh       = "high"
	ClearanceMedium     = "medium"
	ClearanceLow        = "low"
	ClearanceRestricted = "restricted"
)

// Trajectory values
const (
	TrajectoryRising           = "rising"
	TrajectoryRapidlyRising    = "rapidly_rising"
	TrajectoryStable           = "stable"
	TrajectoryDeclining        = "slowly_declining"
	TrajectoryRapidlyDeclining = "declining"
)

// Surveillance priority levels
const (
	SurvPriorityLow      = "low"
	SurvPriorityMedium   = "medium"
	SurvPriorityElevated = "elevated"
	SurvPriorityHigh     = "high"
)

// Surveillance types
const (
	SurvTypeRoutine    = "routine"
	SurvTypePeriodic   = "periodic"
	SurvTypeTargeted   = "targeted"
	SurvTypeContinuous = "continuous"
)

// IdeologyAlignment trajectories
const (
	IdeologyTrajectoryStable     = "stable"
	IdeologyTrajectoryImproving  = "improving"
	IdeologyTrajectoryConcerning = "concerning"
	IdeologyTrajectoryDangerous  = "dangerous"
)

// Reeducation statuses
const (
	ReeducationRequired   = "required"
	ReeducationInProgress = "in_progress"
	ReeducationCompleted  = "completed"
	ReeducationVoluntary  = "voluntary"
	ReeducationCompulsory = "compulsory"
)
