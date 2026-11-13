package astro

import "time"

// CelestialPosition represents the topocentric apparent position of a celestial body
// (Sun or Earth) as viewed from an observer's location on the lunar surface.
type CelestialPosition struct {
	AzimuthDeg   float64 `json:"azimuth_deg"`   // 0° to 360° clockwise from Lunar North
	ElevationDeg float64 `json:"elevation_deg"` // Degrees above local horizontal plane (-90° to +90°)
	DistanceKm   float64 `json:"distance_km"`   // Distance from observer to celestial body center
}

// SubBodyPoint holds the selenographic coordinates of the point directly beneath a celestial body.
type SubBodyPoint struct {
	LatitudeDeg  float64 `json:"latitude_deg"`
	LongitudeDeg float64 `json:"longitude_deg"`
	DistanceKm   float64 `json:"distance_km"`
}

// EphemerisPoint contains the state of Sun and Earth at a specific mission timestamp.
type EphemerisPoint struct {
	TimestampUTC time.Time         `json:"timestamp_utc"`
	TimeOffsetH  int               `json:"time_offset_hours"`
	Sun          CelestialPosition `json:"sun"`
	Earth        CelestialPosition `json:"earth"`
	SubSolar     SubBodyPoint      `json:"sub_solar"`
	SubEarth     SubBodyPoint      `json:"sub_earth"`
}
