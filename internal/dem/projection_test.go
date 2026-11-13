package dem

import (
	"math"
	"testing"
)

func TestPolarStereographicRoundtrip(t *testing.T) {
	proj := NewPolarStereographic()

	testPoints := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"South Pole Center", -90.0, 0.0},
		{"Shackleton Ridge", -89.45, 120.12},
		{"Malapert Massif", -86.04, 2.71},
		{"Faustini Rim", -87.15, 84.30},
		{"de Gerlache Rim", -88.50, -88.20},
		{"Polar Cap Boundary", -75.50, -45.00},
	}

	for _, tc := range testPoints {
		t.Run(tc.name, func(t *testing.T) {
			x, y, err := proj.Forward(tc.lat, tc.lon)
			if err != nil {
				t.Fatalf("Forward(%f, %f) failed: %v", tc.lat, tc.lon, err)
			}

			invLat, invLon := proj.Inverse(x, y)

			latDiff := math.Abs(invLat - tc.lat)
			if latDiff > 1e-5 {
				t.Errorf("Latitude mismatch: expected %f, got %f (diff: %e)", tc.lat, invLat, latDiff)
			}

			// At the exact pole (-90°), longitude is degenerate, check only for other points
			if tc.lat > -89.99 {
				lonDiff := math.Abs(invLon - tc.lon)
				if lonDiff > 180.0 {
					lonDiff = math.Abs(lonDiff - 360.0)
				}
				if lonDiff > 1e-4 {
					t.Errorf("Longitude mismatch: expected %f, got %f (diff: %e)", tc.lon, invLon, lonDiff)
				}
			}
		})
	}
}

func TestPolarStereographicInvalidLatitude(t *testing.T) {
	proj := NewPolarStereographic()

	// Latitude > -75° (Equator / mid-latitudes)
	_, _, err := proj.Forward(-50.0, 0.0)
	if err != ErrLatitudeOutOfRange {
		t.Errorf("expected ErrLatitudeOutOfRange for lat -50.0, got %v", err)
	}

	// Latitude < -90° (Impossible south of South Pole)
	_, _, err = proj.Forward(-95.0, 0.0)
	if err != ErrLatitudeOutOfRange {
		t.Errorf("expected ErrLatitudeOutOfRange for lat -95.0, got %v", err)
	}
}
