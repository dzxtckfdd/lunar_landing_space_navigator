package astro

import (
	"sync"
	"time"
)

// EphemerisService coordinates the calculation and in-memory caching of celestial positions.
type EphemerisService struct {
	mu    sync.RWMutex
	cache map[int64]subBodySnapshot // Keyed by unix hour timestamp
}

type subBodySnapshot struct {
	sun   SubBodyPoint
	earth SubBodyPoint
}

// NewEphemerisService creates a new instance of the ephemeris service.
func NewEphemerisService() *EphemerisService {
	return &EphemerisService{
		cache: make(map[int64]subBodySnapshot),
	}
}

// getOrComputeSubBodies retrieves cached sub-body coordinates or calculates them.
func (s *EphemerisService) getOrComputeSubBodies(t time.Time) (sun, earth SubBodyPoint) {
	// Round down to the nearest hour for caching
	hourKey := t.Truncate(time.Hour).Unix()

	s.mu.RLock()
	cached, found := s.cache[hourKey]
	s.mu.RUnlock()

	if found {
		return cached.sun, cached.earth
	}

	// Compute sub-body points
	sun = GetSubSolarPoint(t)
	earth = GetSubEarthPoint(t)

	s.mu.Lock()
	s.cache[hourKey] = subBodySnapshot{sun: sun, earth: earth}
	s.mu.Unlock()

	return sun, earth
}

// GetPositions calculates the apparent topocentric azimuth and elevation of both the Sun and Earth
// at a specific timestamp for a given location on the Moon.
func (s *EphemerisService) GetPositions(obsLat, obsLon, obsElevM float64, t time.Time) (sun, earth CelestialPosition) {
	subSun, subEarth := s.getOrComputeSubBodies(t)

	sunAz, sunEl := CalculateTopocentric(
		obsLat, obsLon, obsElevM,
		subSun.LatitudeDeg, subSun.LongitudeDeg, subSun.DistanceKm,
	)

	earthAz, earthEl := CalculateTopocentric(
		obsLat, obsLon, obsElevM,
		subEarth.LatitudeDeg, subEarth.LongitudeDeg, subEarth.DistanceKm,
	)

	sun = CelestialPosition{
		AzimuthDeg:   sunAz,
		ElevationDeg: sunEl,
		DistanceKm:   subSun.DistanceKm,
	}

	earth = CelestialPosition{
		AzimuthDeg:   earthAz,
		ElevationDeg: earthEl,
		DistanceKm:   subEarth.DistanceKm,
	}

	return sun, earth
}

// GetTimeline computes the celestial state vectors for each hourly step across the mission duration.
func (s *EphemerisService) GetTimeline(
	obsLat, obsLon, obsElevM float64,
	startUTC time.Time,
	durationHours int,
	stepHours int,
) []EphemerisPoint {
	if stepHours <= 0 {
		stepHours = 1
	}

	totalSteps := (durationHours / stepHours) + 1
	timeline := make([]EphemerisPoint, 0, totalSteps)

	for h := 0; h <= durationHours; h += stepHours {
		currentT := startUTC.Add(time.Duration(h) * time.Hour)
		subSun, subEarth := s.getOrComputeSubBodies(currentT)

		sunAz, sunEl := CalculateTopocentric(
			obsLat, obsLon, obsElevM,
			subSun.LatitudeDeg, subSun.LongitudeDeg, subSun.DistanceKm,
		)
		earthAz, earthEl := CalculateTopocentric(
			obsLat, obsLon, obsElevM,
			subEarth.LatitudeDeg, subEarth.LongitudeDeg, subEarth.DistanceKm,
		)

		timeline = append(timeline, EphemerisPoint{
			TimestampUTC: currentT,
			TimeOffsetH:  h,
			Sun: CelestialPosition{
				AzimuthDeg:   sunAz,
				ElevationDeg: sunEl,
				DistanceKm:   subSun.DistanceKm,
			},
			Earth: CelestialPosition{
				AzimuthDeg:   earthAz,
				ElevationDeg: earthEl,
				DistanceKm:   subEarth.DistanceKm,
			},
			SubSolar: subSun,
			SubEarth: subEarth,
		})
	}

	return timeline
}
