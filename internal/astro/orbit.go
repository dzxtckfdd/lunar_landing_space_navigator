package astro

import (
	"lunar-navigator/internal/dem"
	"math"
	"time"
)

const (
	// SynodicMonthDays is the mean period of lunar phases (29d 12h 44m 2.9s)
	SynodicMonthDays = 29.530588853
	// TropicalMonthDays is the mean period between crossings of the equinox (27.321582 days)
	TropicalMonthDays = 27.321582
	// DraconicMonthDays is the period between nodal crossings (27.212220 days)
	DraconicMonthDays = 27.212220
	// AnomalisticMonthDays is the period between perigees (27.554550 days)
	AnomalisticMonthDays = 27.554550

	// MoonObliquityDeg is the inclination of the lunar equator to the ecliptic plane
	MoonObliquityDeg = 1.5424

	// MeanEarthMoonDistanceKm is the semi-major axis of Moon's orbit around Earth
	MeanEarthMoonDistanceKm = 384400.0
	// LunarOrbitalEccentricity
	LunarEccentricity = 0.0549
	// MeanSunDistanceKm (1 AU)
	MeanSunDistanceKm = 149597870.7
)

// JulianDate converts a standard Go time.Time to Julian Date (Days since January 1, 4713 BC 12:00 UTC).
func JulianDate(t time.Time) float64 {
	utc := t.UTC()
	y := utc.Year()
	m := int(utc.Month())
	d := utc.Day()

	if m <= 2 {
		y -= 1
		m += 12
	}

	a := y / 100
	b := 2 - a + (a / 4)

	dayFraction := (float64(utc.Hour()) + float64(utc.Minute())/60.0 + float64(utc.Second())/3600.0) / 24.0

	jd := math.Floor(365.25*float64(y+4716)) + math.Floor(30.6001*float64(m+1)) + float64(d) + dayFraction + float64(b) - 1524.5
	return jd
}

// DaysSinceJ2000 returns the fractional days elapsed since the standard epoch J2000.0 (2000-01-01 12:00:00 UTC).
func DaysSinceJ2000(t time.Time) float64 {
	return JulianDate(t) - 2451545.0
}

// GetSubSolarPoint calculates the selenographic coordinates (latitude, longitude, distance)
// of the subsolar point on the Moon at the given timestamp.
func GetSubSolarPoint(t time.Time) SubBodyPoint {
	d := DaysSinceJ2000(t)

	// Mean solar longitude on the ecliptic (degrees)
	// L = 280.460 + 0.9856474 * d
	meanSunLonDeg := math.Mod(280.460+0.9856474*d, 360.0)
	if meanSunLonDeg < 0 {
		meanSunLonDeg += 360.0
	}

	// Longitude of the Moon's ascending node (degrees)
	// Omega = 125.04 - 0.0529538083 * d
	nodeLonDeg := math.Mod(125.04-0.0529538083*d, 360.0)
	if nodeLonDeg < 0 {
		nodeLonDeg += 360.0
	}

	// Phase angle of the Sun relative to lunar equatorial node
	lambdaPrimeRad := (meanSunLonDeg - nodeLonDeg) * dem.DegToRad

	// Subsolar latitude: oscillates by ±1.5424°
	subSolarLatDeg := MoonObliquityDeg * math.Sin(lambdaPrimeRad)

	// Subsolar longitude: travels 360° westward every synodic month (~29.53 days)
	// At J2000.0 (d=0), the subsolar longitude was approximately 0.0°
	subSolarLonDeg := math.Mod(180.0-360.0*(d/SynodicMonthDays), 360.0)
	if subSolarLonDeg > 180.0 {
		subSolarLonDeg -= 360.0
	} else if subSolarLonDeg < -180.0 {
		subSolarLonDeg += 360.0
	}

	// Distance variation due to Earth-Sun orbital eccentricity
	meanAnomalySun := (357.528 + 0.9856003*d) * dem.DegToRad
	sunDistKm := MeanSunDistanceKm * (1.0 - 0.0167*math.Cos(meanAnomalySun))

	return SubBodyPoint{
		LatitudeDeg:  subSolarLatDeg,
		LongitudeDeg: subSolarLonDeg,
		DistanceKm:   sunDistKm,
	}
}

// GetSubEarthPoint calculates the selenographic coordinates (latitude, longitude, distance)
// of the sub-Earth point on the Moon, modeling optical and physical lunar libration.
func GetSubEarthPoint(t time.Time) SubBodyPoint {
	d := DaysSinceJ2000(t)

	// Mean anomaly of the Moon (period: anomalistic month = 27.55455 days)
	meanAnomalyMoonRad := (134.963 + 13.064993*d) * dem.DegToRad

	// Argument of latitude (period: draconic month = 27.21222 days)
	argumentLatitudeRad := (93.272 + 13.229350*d) * dem.DegToRad

	// 1. Optical Libration in Longitude (amplitude ~ 7.9°)
	// Dominated by the equation of the center due to orbital eccentricity
	librationLonDeg := 7.9 * math.Sin(meanAnomalyMoonRad)

	// 2. Optical Libration in Latitude (amplitude ~ 6.7°)
	// Dominated by the inclination of Moon's equator to its orbital plane
	librationLatDeg := 6.7 * math.Sin(argumentLatitudeRad)

	// 3. Earth-Moon distance (varies between ~363,300 km at perigee and ~405,500 km at apogee)
	earthDistKm := MeanEarthMoonDistanceKm * (1.0 - LunarEccentricity*math.Cos(meanAnomalyMoonRad))

	return SubBodyPoint{
		LatitudeDeg:  librationLatDeg,
		LongitudeDeg: librationLonDeg,
		DistanceKm:   earthDistKm,
	}
}
