package astro

import (
	"math"
	"testing"
	"time"
)

func TestJulianDateEpochJ2000(t *testing.T) {
	// J2000.0 is 2000-01-01 12:00:00 UTC = 2451545.0 JD
	epoch := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	jd := JulianDate(epoch)

	if math.Abs(jd-2451545.0) > 1e-4 {
		t.Errorf("expected JD 2451545.0 for J2000.0, got %f", jd)
	}

	d := DaysSinceJ2000(epoch)
	if math.Abs(d) > 1e-4 {
		t.Errorf("expected DaysSinceJ2000 = 0.0 for J2000.0, got %f", d)
	}
}

func TestSubSolarPointBounds(t *testing.T) {
	// Sample multiple dates across an entire year
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	for day := 0; day < 365; day += 5 {
		testTime := start.AddDate(0, 0, day)
		subSolar := GetSubSolarPoint(testTime)

		// Subsolar latitude MUST be bounded by Moon's obliquity ±1.5424°
		if math.Abs(subSolar.LatitudeDeg) > MoonObliquityDeg+0.01 {
			t.Errorf("subsolar latitude out of bounds at day %d: %f° (max: %f°)",
				day, subSolar.LatitudeDeg, MoonObliquityDeg)
		}

		// Longitude must be in [-180, +180]
		if subSolar.LongitudeDeg < -180.0 || subSolar.LongitudeDeg > 180.0 {
			t.Errorf("subsolar longitude out of range [-180, +180]: %f°", subSolar.LongitudeDeg)
		}

		// Distance to Sun ~ 1 AU (147 - 152 million km)
		if subSolar.DistanceKm < 147e6 || subSolar.DistanceKm > 153e6 {
			t.Errorf("subsolar distance out of realistic AU range: %f km", subSolar.DistanceKm)
		}
	}
}

func TestSubEarthPointLibrationBounds(t *testing.T) {
	// Sample dates over multiple lunar orbital periods (100 days)
	start := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

	maxObservedLat := 0.0
	maxObservedLon := 0.0

	for h := 0; h < 2400; h += 6 {
		testTime := start.Add(time.Duration(h) * time.Hour)
		subEarth := GetSubEarthPoint(testTime)

		if math.Abs(subEarth.LatitudeDeg) > maxObservedLat {
			maxObservedLat = math.Abs(subEarth.LatitudeDeg)
		}
		if math.Abs(subEarth.LongitudeDeg) > maxObservedLon {
			maxObservedLon = math.Abs(subEarth.LongitudeDeg)
		}

		// Earth distance should stay between perigee ~363,000 km and apogee ~406,000 km
		if subEarth.DistanceKm < 360000 || subEarth.DistanceKm > 410000 {
			t.Errorf("sub-Earth distance out of bounds: %f km", subEarth.DistanceKm)
		}
	}

	// Verify libration reaches expected theoretical amplitudes:
	// Latitude libration amplitude ~ 6.7°
	if maxObservedLat > 6.75 || maxObservedLat < 6.5 {
		t.Errorf("expected max latitude libration ~6.7°, got %f°", maxObservedLat)
	}

	// Longitude libration amplitude ~ 7.9°
	if maxObservedLon > 7.95 || maxObservedLon < 7.7 {
		t.Errorf("expected max longitude libration ~7.9°, got %f°", maxObservedLon)
	}

	t.Logf("Libration bounds verified: max lat = %.2f°, max lon = %.2f°", maxObservedLat, maxObservedLon)
}
