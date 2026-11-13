package dem

import (
	"math"
	"testing"
)

func TestElevationMatrixBilinearInterpolation(t *testing.T) {
	// 4x4 matrix with 100m resolution, centered at (0, 0)
	m := NewElevationMatrix(4, 4, 100.0, 0.0, 0.0)

	// Set known values: corner (col=1, row=1) = 1000m, (col=2, row=1) = 2000m
	m.SetElevation(1, 1, 1000.0)
	m.SetElevation(2, 1, 2000.0)
	m.SetElevation(1, 2, 1000.0)
	m.SetElevation(2, 2, 2000.0)

	// Coordinates between (col 1.5, row 1.5)
	x15, y15 := m.GridToWorld(1.5, 1.5)
	interpolated, err := m.InterpolateElevation(x15, y15)
	if err != nil {
		t.Fatalf("InterpolateElevation failed: %v", err)
	}

	expected := 1500.0
	if math.Abs(interpolated-expected) > 0.01 {
		t.Errorf("expected interpolated elevation %f, got %f", expected, interpolated)
	}
}

func TestCalculateSlopeFlatVsTilted(t *testing.T) {
	// Create synthetic matrix with 50m resolution
	matrix := NewSyntheticMatrix(1500, 1500, 50.0)

	// Flat undulating plain far from crater rim (at x = 25000m, y = 25000m where slope is low)
	plainLat, plainLon := matrix.Proj.Inverse(25000.0, 25000.0)
	slope, err := matrix.CalculateSlope(plainLat, plainLon)
	if err != nil {
		t.Fatalf("CalculateSlope on plain failed: %v", err)
	}

	// Slopes on undulating plains should be low (< 10 degrees)
	if slope > 10.0 {
		t.Errorf("expected gentle slope on plain (<10°), got %f°", slope)
	}

	// Inside the steep crater wall (x = 7000m, y = 7000m, dist ~ 9.9 km, near the 10.5 km rim)
	craterWallLat, craterWallLon := matrix.Proj.Inverse(7000.0, 7000.0)
	wallSlope, err := matrix.CalculateSlope(craterWallLat, craterWallLon)
	if err != nil {
		t.Fatalf("CalculateSlope on crater wall failed: %v", err)
	}

	// Wall slope should be significantly steep (> 20 degrees)
	if wallSlope < 20.0 {
		t.Errorf("expected steep crater wall slope (>20°), got %f°", wallSlope)
	}
}
