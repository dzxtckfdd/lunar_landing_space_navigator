package dem

import (
	"math"
	"sync"
	"time"
)

// HorizonPoint represents the maximum terrain elevation angle in degrees for a single azimuth direction.
type HorizonPoint struct {
	AzimuthDeg          float64 `json:"azimuth"`
	HorizonElevationDeg float64 `json:"horizon_elevation_deg"`
	DistanceM           float64 `json:"distance_m"`
	PeakElevationM      float64 `json:"peak_elevation_m"`
}

// HorizonProfile holds the complete 360-degree horizon elevation mask computed around a landing site.
type HorizonProfile struct {
	ObserverLat         float64        `json:"observer_lat"`
	ObserverLon         float64        `json:"observer_lon"`
	ObserverElevationM  float64        `json:"observer_elevation_m"`
	MastHeightM         float64        `json:"mast_height_m"`
	Points              []HorizonPoint `json:"horizon_profile"`
	MinElevationDeg     float64        `json:"min_horizon_elevation_deg"`
	MaxElevationDeg     float64        `json:"max_horizon_elevation_deg"`
	ComputationTimeMs   float64        `json:"computation_time_ms"`
}

// Raycaster performs fast parallel raymarching against the in-memory DEM matrix.
type Raycaster struct {
	matrix      *ElevationMatrix
	maxRadiusM  float64 // Maximum horizon search distance in meters (e.g. 35,000m = 35km)
	stepM       float64 // Radial step size in meters (e.g. 100m)
	workerCount int     // Number of concurrent goroutines
}

// NewRaycaster creates a new raycaster instance for an elevation matrix.
func NewRaycaster(m *ElevationMatrix) *Raycaster {
	return &Raycaster{
		matrix:      m,
		maxRadiusM:  35000.0, // 35 km search radius covers major lunar horizon features
		stepM:       100.0,   // 100m radial sampling step
		workerCount: 8,       // Multi-core parallel workers
	}
}

// ComputeHorizon calculates the 360-degree local horizon elevation mask around (lat, lon).
func (rc *Raycaster) ComputeHorizon(latDeg, lonDeg, mastHeightM float64) (*HorizonProfile, error) {
	start := time.Now()

	x0, y0, err := rc.matrix.Proj.Forward(latDeg, lonDeg)
	if err != nil {
		return nil, err
	}

	surfaceElev, err := rc.matrix.InterpolateElevation(x0, y0)
	if err != nil {
		return nil, err
	}

	observerElev := surfaceElev + mastHeightM

	// Determine local North and East unit vectors at (x0, y0)
	// On South Polar stereographic projection:
	// North points radially away from the South Pole (0,0) towards the equator.
	var uNorthX, uNorthY float64
	distFromPole := math.Hypot(x0, y0)
	if distFromPole < 1.0 {
		// At South Pole, define North towards longitude 0° (+Y in LOLA convention)
		uNorthX = 0.0
		uNorthY = -1.0
	} else {
		uNorthX = x0 / distFromPole
		uNorthY = y0 / distFromPole
	}

	// East is 90° clockwise from North: (uNx, uNy) rotated 90° CW -> (uNy, -uNx)
	uEastX := uNorthY
	uEastY := -uNorthX

	numRays := 360
	points := make([]HorizonPoint, numRays)

	// Worker Pool for parallel raymarching across 360 azimuth directions
	var wg sync.WaitGroup
	chunkSize := (numRays + rc.workerCount - 1) / rc.workerCount

	for w := 0; w < rc.workerCount; w++ {
		startAz := w * chunkSize
		endAz := startAz + chunkSize
		if endAz > numRays {
			endAz = numRays
		}
		if startAz >= endAz {
			continue
		}

		wg.Add(1)
		go func(fromAz, toAz int) {
			defer wg.Done()
			for azIdx := fromAz; azIdx < toAz; azIdx++ {
				azimuthDeg := float64(azIdx)
				azRad := azimuthDeg * DegToRad

				// Direction vector along azimuth: cos(az)*North + sin(az)*East
				dirX := math.Cos(azRad)*uNorthX + math.Sin(azRad)*uEastX
				dirY := math.Cos(azRad)*uNorthY + math.Sin(azRad)*uEastY

				maxAngleDeg := -90.0
				peakDist := 0.0
				peakElev := surfaceElev

				// Raymarch outwards from start radius to max radius
				for r := rc.stepM; r <= rc.maxRadiusM; r += rc.stepM {
					currX := x0 + r*dirX
					currY := y0 + r*dirY

					h, err := rc.matrix.InterpolateElevation(currX, currY)
					if err != nil {
						// Hit matrix boundary, stop this ray
						break
					}

					// Calculate angular elevation above local horizontal tangent plane:
					// angle = arctan((h - obsElev)/r) - (r / (2 * R_moon)) in radians
					dz := h - observerElev
					curvatureDropM := (r * r) / (2.0 * LunarMeanRadiusM)
					effectiveDz := dz - curvatureDropM
					angleDeg := math.Atan2(effectiveDz, r) * RadToDeg

					if angleDeg > maxAngleDeg {
						maxAngleDeg = angleDeg
						peakDist = r
						peakElev = h
					}
				}

				points[azIdx] = HorizonPoint{
					AzimuthDeg:          azimuthDeg,
					HorizonElevationDeg: math.Round(maxAngleDeg*100) / 100,
					DistanceM:           math.Round(peakDist),
					PeakElevationM:      math.Round(peakElev*10) / 10,
				}
			}
		}(startAz, endAz)
	}

	wg.Wait()

	// Compute min and max horizon elevation
	minElev := points[0].HorizonElevationDeg
	maxElev := points[0].HorizonElevationDeg
	for _, p := range points {
		if p.HorizonElevationDeg < minElev {
			minElev = p.HorizonElevationDeg
		}
		if p.HorizonElevationDeg > maxElev {
			maxElev = p.HorizonElevationDeg
		}
	}

	elapsed := float64(time.Since(start).Microseconds()) / 1000.0

	return &HorizonProfile{
		ObserverLat:         latDeg,
		ObserverLon:         lonDeg,
		ObserverElevationM:  surfaceElev,
		MastHeightM:         mastHeightM,
		Points:              points,
		MinElevationDeg:     minElev,
		MaxElevationDeg:     maxElev,
		ComputationTimeMs:   elapsed,
	}, nil
}
