package evaluation

import (
	"lunar-navigator/internal/astro"
	"lunar-navigator/internal/dem"
	"math"
	"time"
)

// Service coordinates the terrain horizon raymarching and astrodynamic calculations.
type Service struct {
	matrix       *dem.ElevationMatrix
	raycaster    *dem.Raycaster
	astroService *astro.EphemerisService
}

// NewService creates a new mission evaluation service.
func NewService(
	matrix *dem.ElevationMatrix,
	raycaster *dem.Raycaster,
	astroService *astro.EphemerisService,
) *Service {
	return &Service{
		matrix:       matrix,
		raycaster:    raycaster,
		astroService: astroService,
	}
}

// Evaluate performs the complete multi-day landing site environmental evaluation.
func (s *Service) Evaluate(req Request) (*Result, error) {
	startTimer := time.Now()

	// 1. Apply defaults and sanitize input
	if req.MastHeightM <= 0 {
		req.MastHeightM = 2.5
	}
	if req.MaxShadowFreezeHours <= 0 {
		req.MaxShadowFreezeHours = 36
	}
	if req.DurationHours <= 0 {
		req.DurationHours = 336 // Default 14 Earth days (half a lunar synodic cycle)
	}
	if req.SolarPanelType == "" {
		req.SolarPanelType = PanelVerticalCylindrical
	}
	if req.MissionStartUTC.IsZero() {
		req.MissionStartUTC = time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
	}

	// Sanitize weights (ensure sum = 1.0)
	totalW := req.Weights.Solar + req.Weights.Comm + req.Weights.Slope + req.Weights.Science
	if totalW <= 0.001 {
		req.Weights = DefaultWeights()
	} else {
		req.Weights.Solar /= totalW
		req.Weights.Comm /= totalW
		req.Weights.Slope /= totalW
		req.Weights.Science /= totalW
	}

	// 2. Apply Launch Date Slip
	effectiveStartUTC := req.MissionStartUTC.AddDate(0, 0, req.LaunchSlipDays)
	effectiveEndUTC := effectiveStartUTC.Add(time.Duration(req.DurationHours) * time.Hour)

	// 3. Terrain analysis: Surface elevation and slope
	surfaceElev, err := s.matrix.GetElevationAtCoords(req.Latitude, req.Longitude)
	if err != nil {
		return nil, err
	}

	slopeDeg, err := s.matrix.CalculateSlope(req.Latitude, req.Longitude)
	if err != nil {
		return nil, err
	}

	// 4. Compute 360-degree local horizon profile
	horizonProfile, err := s.raycaster.ComputeHorizon(req.Latitude, req.Longitude, req.MastHeightM)
	if err != nil {
		return nil, err
	}

	// Fast horizon lookup table indexed by integer azimuth (0..359)
	horizonLookup := make([]float64, 360)
	for _, p := range horizonProfile.Points {
		idx := int(p.AzimuthDeg) % 360
		if idx < 0 {
			idx += 360
		}
		horizonLookup[idx] = p.HorizonElevationDeg
	}

	// 5. Compute hourly celestial positions for the mission timeline
	ephemTimeline := s.astroService.GetTimeline(
		req.Latitude, req.Longitude, surfaceElev+req.MastHeightM,
		effectiveStartUTC, req.DurationHours, 1,
	)

	// 6. Process each hour: check line-of-sight occlusion, solar flux, and blackout windows
	timeline := make([]HourlyStep, len(ephemTimeline))

	illuminatedHours := 0
	dteConnectedHours := 0

	currentSunlightStreak := 0
	maxSunlightStreak := 0

	currentShadowStreak := 0
	longestShadowStreak := 0
	thermalFailure := false

	currentBlackoutStreak := 0
	longestBlackoutStreak := 0

	for i, pt := range ephemTimeline {
		// Look up horizon obstruction at the azimuth of Sun and Earth
		sunAzIdx := int(math.Round(pt.Sun.AzimuthDeg)) % 360
		if sunAzIdx < 0 {
			sunAzIdx += 360
		}
		sunHorizonMask := horizonLookup[sunAzIdx]

		earthAzIdx := int(math.Round(pt.Earth.AzimuthDeg)) % 360
		if earthAzIdx < 0 {
			earthAzIdx += 360
		}
		earthHorizonMask := horizonLookup[earthAzIdx]

		// Line-of-sight visibility conditions
		isSunlit := pt.Sun.ElevationDeg > sunHorizonMask
		isDTE := pt.Earth.ElevationDeg > earthHorizonMask

		// Relative solar power flux based on lander solar panel geometry
		solarFlux := 0.0
		if isSunlit {
			sunElevRad := pt.Sun.ElevationDeg * dem.DegToRad
			switch req.SolarPanelType {
			case PanelVerticalCylindrical:
				// Vertical panels receive side-grazing sunlight: flux ~ cos(elevation)
				solarFlux = math.Max(0.0, math.Cos(sunElevRad))
			case PanelHorizontalFlat:
				// Flat horizontal panel: flux ~ sin(elevation)
				solarFlux = math.Max(0.0, math.Sin(sunElevRad))
			case PanelTiltedTracking:
				// Optimal tracking panel pointing at sun
				solarFlux = 1.0
			default:
				solarFlux = math.Max(0.0, math.Cos(sunElevRad))
			}
		}

		// Update illumination & thermal streaks
		if isSunlit {
			illuminatedHours++
			currentSunlightStreak++
			if currentSunlightStreak > maxSunlightStreak {
				maxSunlightStreak = currentSunlightStreak
			}
			currentShadowStreak = 0
		} else {
			currentSunlightStreak = 0
			currentShadowStreak++
			if currentShadowStreak > longestShadowStreak {
				longestShadowStreak = currentShadowStreak
			}
			if currentShadowStreak > req.MaxShadowFreezeHours {
				thermalFailure = true
			}
		}

		// Update communication & blackout streaks
		if isDTE {
			dteConnectedHours++
			currentBlackoutStreak = 0
		} else {
			currentBlackoutStreak++
			if currentBlackoutStreak > longestBlackoutStreak {
				longestBlackoutStreak = currentBlackoutStreak
			}
		}

		timeline[i] = HourlyStep{
			TimeOffsetHours:   pt.TimeOffsetH,
			TimestampUTC:      pt.TimestampUTC,
			SunAzimuthDeg:     math.Round(pt.Sun.AzimuthDeg*100) / 100,
			SunElevationDeg:   math.Round(pt.Sun.ElevationDeg*100) / 100,
			SunHorizonMask:    math.Round(sunHorizonMask*100) / 100,
			IsIlluminated:     isSunlit,
			SolarFluxRel:      math.Round(solarFlux*1000) / 1000,
			EarthAzimuthDeg:   math.Round(pt.Earth.AzimuthDeg*100) / 100,
			EarthElevationDeg: math.Round(pt.Earth.ElevationDeg*100) / 100,
			EarthHorizonMask:  math.Round(earthHorizonMask*100) / 100,
			IsVisibleDTE:      isDTE,
		}
	}

	totalHours := len(ephemTimeline)
	illumPct := (float64(illuminatedHours) / float64(totalHours)) * 100.0
	dtePct := (float64(dteConnectedHours) / float64(totalHours)) * 100.0

	// 7. Calculate Multi-Criteria Feasibility Score (0 - 100)
	sSolar := illumPct
	sComm := dtePct
	sSlope := math.Max(0.0, 100.0-(slopeDeg/15.0)*100.0)
	sScience := 80.0 // Base regional science score

	score := req.Weights.Solar*sSolar +
		req.Weights.Comm*sComm +
		req.Weights.Slope*sSlope +
		req.Weights.Science*sScience

	// Penalize score if thermal freeze threshold is exceeded
	if thermalFailure {
		score *= 0.5 // 50% severe penalty for freezing the lander
	}
	score = math.Max(0.0, math.Min(100.0, score))

	elapsedMs := float64(time.Since(startTimer).Microseconds()) / 1000.0

	return &Result{
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		ElevationM: math.Round(surfaceElev*10) / 10,
		SlopeDeg:   math.Round(slopeDeg*10) / 10,
		Summary: Summary{
			MissionStartUTC:            effectiveStartUTC,
			MissionEndUTC:              effectiveEndUTC,
			TotalHours:                 totalHours,
			IlluminatedHours:           illuminatedHours,
			SolarIlluminationPct:       math.Round(illumPct*10) / 10,
			ContinuousSunlightMaxHours: maxSunlightStreak,
			LongestShadowHours:         longestShadowStreak,
			ThermalFailureTriggered:    thermalFailure,
			DTEConnectedHours:          dteConnectedHours,
			DTECommPct:                 math.Round(dtePct*10) / 10,
			LongestCommBlackoutHours:   longestBlackoutStreak,
			FeasibilityScore:           math.Round(score*10) / 10,
		},
		Timeline:          timeline,
		ComputationTimeMs: elapsedMs,
	}, nil
}
