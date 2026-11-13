package astro

import (
	"testing"
	"time"
)

func TestEphemerisServicePositionsAndTimeline(t *testing.T) {
	service := NewEphemerisService()

	obsLat := -89.45 // Shackleton Ridge
	obsLon := 120.12
	obsElev := 1240.0

	start := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)

	// 1. Single timestamp position
	sun, earth := service.GetPositions(obsLat, obsLon, obsElev, start)

	if sun.AzimuthDeg < 0.0 || sun.AzimuthDeg >= 360.0 {
		t.Errorf("Sun azimuth out of [0, 360): %f", sun.AzimuthDeg)
	}
	// At polar regions, Sun elevation stays between approximately -1.5° and +1.5°
	if sun.ElevationDeg < -2.0 || sun.ElevationDeg > 2.0 {
		t.Errorf("Sun elevation unexpected at South Pole: %f°", sun.ElevationDeg)
	}

	// Earth elevation from Shackleton ridge oscillates around 0° to 7°
	if earth.ElevationDeg < -5.0 || earth.ElevationDeg > 15.0 {
		t.Errorf("Earth elevation unexpected at South Pole: %f°", earth.ElevationDeg)
	}

	t.Logf("At %s: Sun az=%.1f°, el=%.2f° | Earth az=%.1f°, el=%.2f°",
		start.Format(time.RFC3339), sun.AzimuthDeg, sun.ElevationDeg, earth.AzimuthDeg, earth.ElevationDeg)

	// 2. Mission Timeline (14 days = 336 hours)
	durationH := 336
	timeline := service.GetTimeline(obsLat, obsLon, obsElev, start, durationH, 1)

	expectedSteps := durationH + 1
	if len(timeline) != expectedSteps {
		t.Fatalf("expected %d timeline steps, got %d", expectedSteps, len(timeline))
	}

	// Verify step offsets and monotonicity of time
	for i, step := range timeline {
		if step.TimeOffsetH != i {
			t.Errorf("step %d offset mismatch: %d", i, step.TimeOffsetH)
		}
	}
}

func BenchmarkEphemerisHourlyTimeline(b *testing.B) {
	service := NewEphemerisService()
	obsLat := -89.45
	obsLon := 120.12
	obsElev := 1240.0
	start := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Compute full 14-day mission timeline (336 hourly steps)
		timeline := service.GetTimeline(obsLat, obsLon, obsElev, start, 336, 1)
		if len(timeline) != 337 {
			b.Fatalf("unexpected timeline length %d", len(timeline))
		}
	}
}
