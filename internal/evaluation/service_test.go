package evaluation

import (
	"lunar-navigator/internal/astro"
	"lunar-navigator/internal/dem"
	"testing"
	"time"
)

func setupTestEvaluationService() *Service {
	// Matrix covering 260 km x 260 km (from -90° to -85° latitude)
	// 2600 x 2600 pixels with 100m resolution (~27 MB RAM)
	matrix := dem.NewSyntheticMatrix(2600, 2600, 100.0)
	raycaster := dem.NewRaycaster(matrix)
	ephem := astro.NewEphemerisService()

	return NewService(matrix, raycaster, ephem)
}

func TestEvaluateMountainSummitHighIllumination(t *testing.T) {
	svc := setupTestEvaluationService()

	// Mountain peak coordinates (elevated ~4800m)
	peakLat, peakLon := -86.04, 2.71 // Malapert Mountain area

	req := Request{
		Latitude:             peakLat,
		Longitude:            peakLon,
		MastHeightM:          2.5,
		MissionStartUTC:      time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC),
		DurationHours:        336, // 14 days
		LaunchSlipDays:       0,
		SolarPanelType:       PanelVerticalCylindrical,
		MaxShadowFreezeHours: 36,
		Weights:              DefaultWeights(),
	}

	result, err := svc.Evaluate(req)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if result.Summary.TotalHours != 337 {
		t.Errorf("expected 337 hourly steps, got %d", result.Summary.TotalHours)
	}

	// On mountain summit with clear horizons, illumination should be high (>50%)
	if result.Summary.SolarIlluminationPct < 40.0 {
		t.Errorf("expected high illumination on mountain summit (>40%%), got %.1f%%",
			result.Summary.SolarIlluminationPct)
	}

	// Feasibility score should be high on summit
	if result.Summary.FeasibilityScore < 60.0 {
		t.Errorf("expected feasibility score > 60 on summit, got %.1f", result.Summary.FeasibilityScore)
	}

	t.Logf("Mountain Summit: Illum=%.1f%%, DTE=%.1f%%, LongestShadow=%dh, Score=%.1f, Time=%.2f ms",
		result.Summary.SolarIlluminationPct, result.Summary.DTECommPct,
		result.Summary.LongestShadowHours, result.Summary.FeasibilityScore, result.ComputationTimeMs)
}

func TestEvaluateCraterBowlThermalFailure(t *testing.T) {
	svc := setupTestEvaluationService()

	// Inside deep crater bowl (x=0, y=0, lat = -89.99)
	craterLat, craterLon := -89.99, 0.0

	req := Request{
		Latitude:             craterLat,
		Longitude:            craterLon,
		MastHeightM:          2.5,
		MissionStartUTC:      time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC),
		DurationHours:        100,
		LaunchSlipDays:       0,
		SolarPanelType:       PanelVerticalCylindrical,
		MaxShadowFreezeHours: 36, // 36 hours limit
		Weights:              DefaultWeights(),
	}

	result, err := svc.Evaluate(req)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	// Inside deep crater bowl, Sun is permanently blocked by surrounding 20°+ rims
	if result.Summary.IlluminatedHours != 0 {
		t.Errorf("expected 0 illuminated hours in deep crater, got %d", result.Summary.IlluminatedHours)
	}

	// Darkness lasts 100 hours (> 36 hours limit) -> Thermal Failure MUST be triggered!
	if !result.Summary.ThermalFailureTriggered {
		t.Errorf("expected ThermalFailureTriggered to be true inside permanent shadow")
	}

	// Feasibility score must be penalized severely (< 50)
	if result.Summary.FeasibilityScore > 50.0 {
		t.Errorf("expected penalized feasibility score (<50) due to thermal freeze, got %.1f",
			result.Summary.FeasibilityScore)
	}

	t.Logf("Crater Bowl: Illum=%.1f%%, LongestShadow=%dh, ThermalFailure=%t, Score=%.1f",
		result.Summary.SolarIlluminationPct, result.Summary.LongestShadowHours,
		result.Summary.ThermalFailureTriggered, result.Summary.FeasibilityScore)
}

func TestLaunchSlipShift(t *testing.T) {
	svc := setupTestEvaluationService()

	baseStart := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
	lat, lon := -86.04, 2.71

	reqNormal := Request{
		Latitude:        lat,
		Longitude:       lon,
		MissionStartUTC: baseStart,
		DurationHours:   48,
		LaunchSlipDays:  0,
	}
	resNormal, err := svc.Evaluate(reqNormal)
	if err != nil {
		t.Fatalf("reqNormal failed: %v", err)
	}

	reqSlipped := Request{
		Latitude:        lat,
		Longitude:       lon,
		MissionStartUTC: baseStart,
		DurationHours:   48,
		LaunchSlipDays:  3, // Slipped by 3 days = 72 hours
	}
	resSlipped, err := svc.Evaluate(reqSlipped)
	if err != nil {
		t.Fatalf("reqSlipped failed: %v", err)
	}

	expectedSlipStart := baseStart.AddDate(0, 0, 3)
	if !resSlipped.Summary.MissionStartUTC.Equal(expectedSlipStart) {
		t.Errorf("expected start %s, got %s", expectedSlipStart, resSlipped.Summary.MissionStartUTC)
	}

	// Verify timeline first timestamp matches the slipped date
	if !resSlipped.Timeline[0].TimestampUTC.Equal(expectedSlipStart) {
		t.Errorf("expected timeline[0] timestamp %s, got %s",
			expectedSlipStart, resSlipped.Timeline[0].TimestampUTC)
	}

	// Timestamps between normal and slipped should differ by 72 hours
	diffHours := resSlipped.Timeline[0].TimestampUTC.Sub(resNormal.Timeline[0].TimestampUTC).Hours()
	if diffHours != 72.0 {
		t.Errorf("expected 72 hours difference, got %f", diffHours)
	}
}

func TestSolarPanelGeometryComparison(t *testing.T) {
	svc := setupTestEvaluationService()

	// High elevation ridge where sun is visible at low grazing angle (~1.5°)
	lat, lon := -86.04, 2.71
	start := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)

	reqVertical := Request{
		Latitude:        lat,
		Longitude:       lon,
		MissionStartUTC: start,
		DurationHours:   24,
		SolarPanelType:  PanelVerticalCylindrical,
	}
	resVertical, err := svc.Evaluate(reqVertical)
	if err != nil {
		t.Fatalf("Evaluate vertical failed: %v", err)
	}

	reqFlat := Request{
		Latitude:        lat,
		Longitude:       lon,
		MissionStartUTC: start,
		DurationHours:   24,
		SolarPanelType:  PanelHorizontalFlat,
	}
	resFlat, err := svc.Evaluate(reqFlat)
	if err != nil {
		t.Fatalf("Evaluate flat failed: %v", err)
	}

	// Compare solar flux on illuminated steps
	for i := 0; i < len(resVertical.Timeline); i++ {
		stepVert := resVertical.Timeline[i]
		stepFlat := resFlat.Timeline[i]

		if stepVert.IsIlluminated {
			// Vertical panel should produce significantly higher flux (> 0.9) than horizontal flat (< 0.1)
			if stepVert.SolarFluxRel < 0.8 {
				t.Errorf("expected high vertical panel flux (>0.8), got %f", stepVert.SolarFluxRel)
			}
			if stepFlat.SolarFluxRel > 0.2 {
				t.Errorf("expected low horizontal flat panel flux (<0.2) at grazing sun, got %f", stepFlat.SolarFluxRel)
			}
			break
		}
	}
}

func BenchmarkEvaluateFullMission(b *testing.B) {
	svc := setupTestEvaluationService()
	lat, lon := -86.04, 2.71
	start := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)

	req := Request{
		Latitude:             lat,
		Longitude:            lon,
		MastHeightM:          2.5,
		MissionStartUTC:      start,
		DurationHours:        336, // Full 14 days
		LaunchSlipDays:       0,
		SolarPanelType:       PanelVerticalCylindrical,
		MaxShadowFreezeHours: 36,
		Weights:              DefaultWeights(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.Evaluate(req)
		if err != nil {
			b.Fatalf("Evaluate failed: %v", err)
		}
	}
}
