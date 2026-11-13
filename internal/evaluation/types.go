package evaluation

import (
	"time"
)

// PanelType defines the geometric orientation of the lander's solar arrays.
type PanelType string

const (
	PanelVerticalCylindrical PanelType = "vertical_cylindrical" // Optimal for polar grazing sunlight
	PanelHorizontalFlat      PanelType = "horizontal_flat"      // Poor for polar grazing sunlight
	PanelTiltedTracking      PanelType = "tilted_tracking"      // Actively tracks sun
)

// Weights defines user-configurable multi-criteria scoring weights (must sum to 1.0).
type Weights struct {
	Solar   float64 `json:"solar"`
	Comm    float64 `json:"comm"`
	Slope   float64 `json:"slope"`
	Science float64 `json:"science"`
}

// DefaultWeights returns standard balanced weights for mission evaluation.
func DefaultWeights() Weights {
	return Weights{
		Solar:   0.40,
		Comm:    0.30,
		Slope:   0.15,
		Science: 0.15,
	}
}

// Request holds all user-specified and default parameters for a landing site evaluation.
type Request struct {
	Latitude             float64   `json:"latitude"`
	Longitude            float64   `json:"longitude"`
	MastHeightM          float64   `json:"mast_height_m"`
	MissionStartUTC      time.Time `json:"mission_start_utc"`
	DurationHours        int       `json:"duration_hours"`
	LaunchSlipDays       int       `json:"launch_slip_days"`
	SolarPanelType       PanelType `json:"solar_panel_type"`
	MaxShadowFreezeHours int       `json:"max_shadow_freeze_hours"`
	Weights              Weights   `json:"weights"`
}

// HourlyStep contains environmental and celestial telemetry for a single hour of the mission.
type HourlyStep struct {
	TimeOffsetHours int       `json:"time_offset_hours"`
	TimestampUTC    time.Time `json:"timestamp_utc"`
	SunAzimuthDeg   float64   `json:"sun_azimuth_deg"`
	SunElevationDeg float64   `json:"sun_elevation_deg"`
	SunHorizonMask  float64   `json:"sun_horizon_mask_deg"`
	IsIlluminated   bool      `json:"is_illuminated"`
	SolarFluxRel    float64   `json:"solar_flux_rel"` // 0.0 to 1.0 relative power output

	EarthAzimuthDeg   float64 `json:"earth_azimuth_deg"`
	EarthElevationDeg float64 `json:"earth_elevation_deg"`
	EarthHorizonMask  float64 `json:"earth_horizon_mask_deg"`
	IsVisibleDTE      bool    `json:"is_visible_dte"`
}

// Summary provides aggregate statistics and score for the landing site over the mission timeline.
type Summary struct {
	MissionStartUTC             time.Time `json:"mission_start_utc"`
	MissionEndUTC               time.Time `json:"mission_end_utc"`
	TotalHours                  int       `json:"total_hours"`
	IlluminatedHours            int       `json:"illuminated_hours"`
	SolarIlluminationPct        float64   `json:"solar_illumination_pct"`
	ContinuousSunlightMaxHours  int       `json:"continuous_sunlight_max_hours"`
	LongestShadowHours          int       `json:"longest_shadow_hours"`
	ThermalFailureTriggered     bool      `json:"thermal_failure_triggered"` // True if shadow exceeded freeze threshold
	DTEConnectedHours           int       `json:"dte_connected_hours"`
	DTECommPct                  float64   `json:"dte_comm_pct"`
	LongestCommBlackoutHours    int       `json:"longest_comm_blackout_hours"`
	FeasibilityScore            float64   `json:"feasibility_score"` // 0 to 100
}

// Result represents the full evaluation response returned to clients.
type Result struct {
	Latitude          float64      `json:"latitude"`
	Longitude         float64      `json:"longitude"`
	ElevationM        float64      `json:"elevation_m"`
	SlopeDeg          float64      `json:"slope_deg"`
	Summary           Summary      `json:"mission_summary"`
	Timeline          []HourlyStep `json:"timeline_hourly"`
	ComputationTimeMs float64      `json:"computation_time_ms"`
}
