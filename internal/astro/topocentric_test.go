package astro

import (
	"math"
	"testing"
)

func TestCalculateTopocentricZenithAndNadir(t *testing.T) {
	// Observer at Shackleton Ridge: lat = -89.45°, lon = 120.0°, elev = 1000m
	obsLat := -89.45
	obsLon := 120.0
	obsElev := 1000.0

	// 1. Target directly at observer's zenith (same lat and lon)
	azimuth, elevation := CalculateTopocentric(obsLat, obsLon, obsElev, obsLat, obsLon, 384400.0)
	if math.Abs(elevation-90.0) > 0.01 {
		t.Errorf("expected zenith elevation ~90.0°, got %f°", elevation)
	}
	t.Logf("Zenith target: elevation = %.2f°, azimuth = %.2f°", elevation, azimuth)

	// 2. Target at observer's nadir (opposite side of the Moon: lat = +89.45°, lon = -60.0°)
	nadirLat := -obsLat
	nadirLon := obsLon - 180.0
	_, nadirElevation := CalculateTopocentric(obsLat, obsLon, obsElev, nadirLat, nadirLon, 384400.0)
	if math.Abs(nadirElevation - (-90.0)) > 0.5 {
		t.Errorf("expected nadir elevation ~ -90.0°, got %f°", nadirElevation)
	}
	t.Logf("Nadir target: elevation = %.2f°", nadirElevation)
}

func TestCalculateTopocentricFromSouthPole(t *testing.T) {
	// Observer at South Pole (-90.0°)
	obsLat := -90.0
	obsLon := 0.0
	obsElev := 0.0

	// Target 1: on equator at lon 0° (due North)
	az1, el1 := CalculateTopocentric(obsLat, obsLon, obsElev, 0.0, 0.0, MeanSunDistanceKm)
	if math.Abs(el1-0.0) > 0.01 {
		t.Errorf("expected elevation ~0.0° at pole for equatorial target, got %f°", el1)
	}
	if math.Abs(az1-0.0) > 0.01 {
		t.Errorf("expected azimuth ~0.0° (North), got %f°", az1)
	}

	// Target 2: on equator at lon 90° (due East)
	az2, el2 := CalculateTopocentric(obsLat, obsLon, obsElev, 0.0, 90.0, MeanSunDistanceKm)
	if math.Abs(el2-0.0) > 0.01 {
		t.Errorf("expected elevation ~0.0°, got %f°", el2)
	}
	if math.Abs(az2-90.0) > 0.01 {
		t.Errorf("expected azimuth ~90.0° (East), got %f°", az2)
	}

	// Target 3: on equator at lon 180° (opposite meridian)
	az3, _ := CalculateTopocentric(obsLat, obsLon, obsElev, 0.0, 180.0, MeanSunDistanceKm)
	if math.Abs(az3-180.0) > 0.01 {
		t.Errorf("expected azimuth ~180.0°, got %f°", az3)
	}
}
