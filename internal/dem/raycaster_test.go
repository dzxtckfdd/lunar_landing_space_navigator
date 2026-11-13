package dem

import (
	"testing"
)

func TestRaycasterInsideCraterBowl(t *testing.T) {
	// 50m resolution synthetic matrix: 500x500 pixels (25 km x 25 km)
	matrix := NewSyntheticMatrix(500, 500, 50.0)
	raycaster := NewRaycaster(matrix)

	// Observer inside the crater bowl near center (lat = -89.99, lon = 0)
	lat, lon := -89.99, 0.0
	profile, err := raycaster.ComputeHorizon(lat, lon, 2.5)
	if err != nil {
		t.Fatalf("ComputeHorizon failed: %v", err)
	}

	if len(profile.Points) != 360 {
		t.Fatalf("expected 360 horizon points, got %d", len(profile.Points))
	}

	// Inside the deep crater, the surrounding rim should obstruct the horizon upwards:
	// All horizon elevation angles should be positive (> 5°)
	if profile.MinElevationDeg < 2.0 {
		t.Errorf("expected all horizon angles to be positive inside crater bowl, got min %f°", profile.MinElevationDeg)
	}

	t.Logf("Inside crater: min horizon angle = %.2f°, max horizon angle = %.2f°, time = %.2f ms",
		profile.MinElevationDeg, profile.MaxElevationDeg, profile.ComputationTimeMs)
}

func TestRaycasterOnMountainPeak(t *testing.T) {
	// Synthetic matrix containing high mountain peak at x=-45000, y=35000
	matrix := NewSyntheticMatrix(2000, 2000, 50.0) // 100 km x 100 km grid
	raycaster := NewRaycaster(matrix)

	// Observer at peak of mountain (x = -45000, y = 35000)
	peakLat, peakLon := matrix.Proj.Inverse(-45000.0, 35000.0)
	profile, err := raycaster.ComputeHorizon(peakLat, peakLon, 2.5)
	if err != nil {
		t.Fatalf("ComputeHorizon on mountain peak failed: %v", err)
	}

	// On a mountain summit looking out over lower plains with lunar curvature,
	// the horizon angle should drop into negative values (< 0°)
	if profile.MinElevationDeg >= 0.0 {
		t.Errorf("expected horizon on high mountain summit to drop below 0° due to depression/curvature, got min %f°", profile.MinElevationDeg)
	}

	t.Logf("On mountain summit: min horizon angle = %.2f°, max horizon angle = %.2f°, time = %.2f ms",
		profile.MinElevationDeg, profile.MaxElevationDeg, profile.ComputationTimeMs)
}

func BenchmarkRaycaster360Azimuths(b *testing.B) {
	matrix := NewSyntheticMatrix(800, 800, 50.0)
	raycaster := NewRaycaster(matrix)
	lat, lon := -89.45, 120.12

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := raycaster.ComputeHorizon(lat, lon, 2.5)
		if err != nil {
			b.Fatalf("raycaster benchmark failed: %v", err)
		}
	}
}
