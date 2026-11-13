package astro

import (
	"lunar-navigator/internal/dem"
	"math"
)

// CalculateTopocentric computes the topocentric apparent Azimuth (0°–360° clockwise from North)
// and Elevation (-90° to +90° above the local tangent plane) for a target body
// given the observer's selenographic coordinates and target's sub-body coordinates.
func CalculateTopocentric(
	obsLatDeg, obsLonDeg, obsElevM float64,
	tgtLatDeg, tgtLonDeg, tgtDistKm float64,
) (azimuthDeg, elevationDeg float64) {
	obsLatRad := obsLatDeg * dem.DegToRad
	obsLonRad := obsLonDeg * dem.DegToRad

	tgtLatRad := tgtLatDeg * dem.DegToRad
	tgtLonRad := tgtLonDeg * dem.DegToRad

	rObsM := dem.LunarMeanRadiusM + obsElevM
	rTgtM := tgtDistKm * 1000.0

	// 1. Observer position vector in selenographic Cartesian frame
	cosObsLat := math.Cos(obsLatRad)
	sinObsLat := math.Sin(obsLatRad)
	cosObsLon := math.Cos(obsLonRad)
	sinObsLon := math.Sin(obsLonRad)

	rObsX := rObsM * cosObsLat * cosObsLon
	rObsY := rObsM * cosObsLat * sinObsLon
	rObsZ := rObsM * sinObsLat

	// 2. Target position vector in selenographic Cartesian frame
	cosTgtLat := math.Cos(tgtLatRad)
	sinTgtLat := math.Sin(tgtLatRad)
	cosTgtLon := math.Cos(tgtLonRad)
	sinTgtLon := math.Sin(tgtLonRad)

	rTgtX := rTgtM * cosTgtLat * cosTgtLon
	rTgtY := rTgtM * cosTgtLat * sinTgtLon
	rTgtZ := rTgtM * sinTgtLat

	// 3. Relative vector from observer to target
	rhoX := rTgtX - rObsX
	rhoY := rTgtY - rObsY
	rhoZ := rTgtZ - rObsZ
	rhoMag := math.Sqrt(rhoX*rhoX + rhoY*rhoY + rhoZ*rhoZ)
	if rhoMag < 1.0 {
		return 0, 0
	}

	// 4. Local Topocentric Reference Frame (Up/Zenith, North, East)
	// Up / Zenith unit vector
	uZ_X := cosObsLat * cosObsLon
	uZ_Y := cosObsLat * sinObsLon
	uZ_Z := sinObsLat

	// North and East unit vectors
	var uN_X, uN_Y, uN_Z float64
	var uE_X, uE_Y, uE_Z float64

	if obsLatDeg <= -89.9999 {
		// Degenerate case at exact South Pole:
		// Define North along Prime Meridian (lon = 0°): [1, 0, 0]
		// Define East along lon = 90°: [0, 1, 0]
		uN_X, uN_Y, uN_Z = 1.0, 0.0, 0.0
		uE_X, uE_Y, uE_Z = 0.0, 1.0, 0.0
	} else {
		// Standard local tangent frame:
		// North: derivative w.r.t latitude
		uN_X = -sinObsLat * cosObsLon
		uN_Y = -sinObsLat * sinObsLon
		uN_Z = cosObsLat

		// East: derivative w.r.t longitude
		uE_X = -sinObsLon
		uE_Y = cosObsLon
		uE_Z = 0.0
	}

	// 5. Project relative vector into Topocentric components
	zTopo := rhoX*uZ_X + rhoY*uZ_Y + rhoZ*uZ_Z // Vertical component (Zenith)
	xTopo := rhoX*uN_X + rhoY*uN_Y + rhoZ*uN_Z // North component
	yTopo := rhoX*uE_X + rhoY*uE_Y + rhoZ*uE_Z // East component

	// 6. Elevation angle: angle above the local horizontal plane
	sinElevation := zTopo / rhoMag
	// Clamp to [-1.0, +1.0] for numerical safety
	if sinElevation > 1.0 {
		sinElevation = 1.0
	} else if sinElevation < -1.0 {
		sinElevation = -1.0
	}
	elevationDeg = math.Asin(sinElevation) * dem.RadToDeg

	// 7. Azimuth angle: clockwise from North (0° = North, 90° = East, 180° = South, 270° = West)
	azimuthRad := math.Atan2(yTopo, xTopo)
	azimuthDeg = azimuthRad * dem.RadToDeg
	if azimuthDeg < 0.0 {
		azimuthDeg += 360.0
	}

	return azimuthDeg, elevationDeg
}
