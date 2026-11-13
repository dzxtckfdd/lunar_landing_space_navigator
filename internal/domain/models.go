package domain

import "time"

// Preset represents an official NASA candidate lunar landing site (e.g. Artemis III regions).
type Preset struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	SubRegion              string    `json:"sub_region"`
	Latitude               float64   `json:"latitude"`
	Longitude              float64   `json:"longitude"`
	ElevationM             float64   `json:"elevation_m"`
	AverageSlopeDeg        float64   `json:"average_slope_deg"`
	DistanceToPSRkm        float64   `json:"distance_to_psr_km"`
	Description            string    `json:"description"`
	ScientificSignificance string    `json:"scientific_significance"`
	CreatedAt              time.Time `json:"created_at"`
}

// CandidateSite represents a discrete point in the spatial candidate grid used by the Recommender Engine.
type CandidateSite struct {
	ID              int64   `json:"id"`
	ClusterRegion   string  `json:"cluster_region"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	ElevationM      float64 `json:"elevation_m"`
	SlopeDeg        float64 `json:"slope_deg"`
	DistanceToPSRkm float64 `json:"distance_to_psr_km"`
	IsSafeSlope     bool    `json:"is_safe_slope"` // True if slope <= 10.0 degrees
}

// SavedScenario represents a user-saved evaluation of a landing site with specific mission parameters.
type SavedScenario struct {
	ID                   string    `json:"id"`
	Title                string    `json:"title"`
	Latitude             float64   `json:"latitude"`
	Longitude            float64   `json:"longitude"`
	AntennaMastHeightM   float64   `json:"antenna_mast_height_m"`
	MissionStartUTC      time.Time `json:"mission_start_utc"`
	MissionDurationHours int       `json:"mission_duration_hours"`
	LaunchSlipDays       int       `json:"launch_slip_days"`
	SolarIlluminationPct float64   `json:"solar_illumination_pct"`
	DTECommPct           float64   `json:"dte_comm_pct"`
	FeasibilityScore     float64   `json:"feasibility_score"`
	Notes                string    `json:"notes,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}
