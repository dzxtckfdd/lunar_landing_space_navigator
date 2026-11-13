package dem

import (
	"errors"
	"math"
)

const (
	// LunarMeanRadiusM is the mean volumetric radius of the Moon in meters (IAU/NASA standard).
	LunarMeanRadiusM = 1737400.0
	// DegToRad converts degrees to radians.
	DegToRad = math.Pi / 180.0
	// RadToDeg converts radians to degrees.
	RadToDeg = 180.0 / math.Pi
)

var (
	ErrLatitudeOutOfRange = errors.New("latitude must be in the south polar region (-90.0 to -75.0 degrees)")
	ErrCoordinatesOutOfBounds = errors.New("projected coordinates fall outside DEM matrix bounds")
)

// PolarStereographic handles conversion between Selenographic coordinates (lat, lon)
// and South Polar Stereographic projection coordinates (X, Y) in meters.
// Tangent point is the South Pole (lat = -90.0°).
type PolarStereographic struct {
	radius float64
}

// NewPolarStereographic creates a new projection instance with the standard Moon radius.
func NewPolarStereographic() *PolarStereographic {
	return &PolarStereographic{
		radius: LunarMeanRadiusM,
	}
}

// Forward converts selenographic (lat, lon in degrees) to projected coordinates (x, y in meters).
// Standard South Polar Stereographic:
// Latitude must be negative (South Pole: -90° to -75°).
func (p *PolarStereographic) Forward(latDeg, lonDeg float64) (float64, float64, error) {
	if latDeg < -90.0 || latDeg > -75.0 {
		return 0, 0, ErrLatitudeOutOfRange
	}

	// At exact south pole (-90°), x = 0, y = 0
	if latDeg <= -89.999999 {
		return 0, 0, nil
	}

	latRad := latDeg * DegToRad
	lonRad := lonDeg * DegToRad

	// Colatitude angle from south pole (0 at -90°, pi/2 at equator)
	colatitude := math.Pi/2.0 + latRad

	// True stereographic radius from South Pole
	rho := 2.0 * p.radius * math.Tan(colatitude/2.0)

	// In NASA LOLA polar stereographic standard:
	// X = rho * sin(lon)
	// Y = -rho * cos(lon)
	x := rho * math.Sin(lonRad)
	y := -rho * math.Cos(lonRad)

	return x, y, nil
}

// Inverse converts projected (x, y in meters) back to selenographic (lat, lon in degrees).
func (p *PolarStereographic) Inverse(x, y float64) (float64, float64) {
	rho := math.Hypot(x, y)
	if rho < 1e-6 {
		return -90.0, 0.0
	}

	// Colatitude from rho
	colatitude := 2.0 * math.Atan(rho/(2.0*p.radius))
	latRad := colatitude - math.Pi/2.0
	latDeg := latRad * RadToDeg

	// Longitude in [-180, +180]
	// lon = atan2(x, -y)
	lonRad := math.Atan2(x, -y)
	lonDeg := lonRad * RadToDeg

	return latDeg, lonDeg
}
