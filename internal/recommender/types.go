package recommender

import (
	"lunar-navigator/internal/evaluation"
	"time"
)

// Request defines the search criteria and lander parameters for the Smart Site Recommender.
type Request struct {
	MinSolarIlluminationPct float64              `json:"min_solar_illumination_pct"`
	MinDTECommPct           float64              `json:"min_dte_comm_pct"`
	MaxSlopeDeg             float64              `json:"max_slope_deg"`
	MaxDistanceToPSRkm      float64              `json:"max_distance_to_psr_km"`
	MissionStartUTC         time.Time            `json:"mission_start_utc"`
	DurationHours           int                  `json:"duration_hours"`
	LaunchSlipDays          int                  `json:"launch_slip_days"`
	PreferredRegion         string               `json:"preferred_region"` // "all" or specific cluster name
	Limit                   int                  `json:"limit"`            // Default: 5
	MastHeightM             float64              `json:"mast_height_m"`
	SolarPanelType          evaluation.PanelType `json:"solar_panel_type"`
	MaxShadowFreezeHours    int                  `json:"max_shadow_freeze_hours"`
}

// RecommendedSite represents a ranked candidate landing location with its scores and telemetry.
type RecommendedSite struct {
	Rank                    int     `json:"rank"`
	Title                   string  `json:"title"`
	ClusterRegion           string  `json:"cluster_region"`
	Latitude                float64 `json:"latitude"`
	Longitude               float64 `json:"longitude"`
	ElevationM              float64 `json:"elevation_m"`
	SlopeDeg                float64 `json:"slope_deg"`
	DistanceToPSRkm         float64 `json:"distance_to_psr_km"`
	SolarIlluminationPct    float64 `json:"solar_illumination_pct"`
	ContinuousSunlightHours int     `json:"continuous_sunlight_hours"`
	DTECommPct              float64 `json:"dte_comm_pct"`
	LongestBlackoutHours    int     `json:"longest_blackout_hours"`
	FeasibilityScore        float64 `json:"feasibility_score"`
	Badge                   string  `json:"badge"`
}

// Response holds the ranked candidate recommendations and execution statistics.
type Response struct {
	CandidatesScanned int               `json:"candidates_scanned"`
	MatchesFound      int               `json:"matches_found"`
	Recommendations   []RecommendedSite `json:"recommendations"`
	ComputationTimeMs float64           `json:"computation_time_ms"`
}
