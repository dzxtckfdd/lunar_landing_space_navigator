package dem

import (
	"math"
)

// ElevationMatrix holds an in-memory 2D raster of lunar surface elevations in meters.
type ElevationMatrix struct {
	Width       int                 // Number of columns
	Height      int                 // Number of rows
	ResolutionM float64             // Spatial resolution in meters per pixel (e.g. 50.0m)
	MinX        float64             // Left boundary in projected meters
	MaxY        float64             // Top boundary in projected meters
	Data        []float32           // Flat elevation buffer in row-major order: index = row * Width + col
	Proj        *PolarStereographic // Coordinate projection
}

// NewElevationMatrix initializes an empty elevation matrix with the given dimensions and resolution.
func NewElevationMatrix(width, height int, resolutionM float64, centerX, centerY float64) *ElevationMatrix {
	halfWidthM := (float64(width) * resolutionM) / 2.0
	halfHeightM := (float64(height) * resolutionM) / 2.0

	return &ElevationMatrix{
		Width:       width,
		Height:      height,
		ResolutionM: resolutionM,
		MinX:        centerX - halfWidthM,
		MaxY:        centerY + halfHeightM,
		Data:        make([]float32, width*height),
		Proj:        NewPolarStereographic(),
	}
}

// WorldToGrid converts projected (x, y in meters) to continuous floating-point matrix coordinates (col, row).
func (m *ElevationMatrix) WorldToGrid(x, y float64) (col, row float64) {
	col = (x - m.MinX) / m.ResolutionM
	row = (m.MaxY - y) / m.ResolutionM
	return col, row
}

// GridToWorld converts matrix indices (col, row) to projected coordinates (x, y in meters).
func (m *ElevationMatrix) GridToWorld(col, row float64) (x, y float64) {
	x = m.MinX + col*m.ResolutionM
	y = m.MaxY - row*m.ResolutionM
	return x, y
}

// GetElevation returns the raw elevation in meters at integer grid indices (col, row).
// If indices are out of bounds, returns the nearest edge pixel (clamping).
func (m *ElevationMatrix) GetElevation(col, row int) float32 {
	if col < 0 {
		col = 0
	} else if col >= m.Width {
		col = m.Width - 1
	}
	if row < 0 {
		row = 0
	} else if row >= m.Height {
		row = m.Height - 1
	}
	return m.Data[row*m.Width+col]
}

// SetElevation sets the elevation at integer grid indices (col, row).
func (m *ElevationMatrix) SetElevation(col, row int, elev float32) {
	if col >= 0 && col < m.Width && row >= 0 && row < m.Height {
		m.Data[row*m.Width+col] = elev
	}
}

// InterpolateElevation calculates elevation at continuous projected coordinates (x, y) using Bilinear Interpolation.
func (m *ElevationMatrix) InterpolateElevation(x, y float64) (float64, error) {
	col, row := m.WorldToGrid(x, y)

	if col < 0 || col >= float64(m.Width-1) || row < 0 || row >= float64(m.Height-1) {
		return 0, ErrCoordinatesOutOfBounds
	}

	c0 := int(math.Floor(col))
	r0 := int(math.Floor(row))
	c1 := c0 + 1
	r1 := r0 + 1

	dx := col - float64(c0)
	dy := row - float64(r0)

	// Sample 4 surrounding grid points
	z00 := float64(m.GetElevation(c0, r0))
	z10 := float64(m.GetElevation(c1, r0))
	z01 := float64(m.GetElevation(c0, r1))
	z11 := float64(m.GetElevation(c1, r1))

	// Bilinear interpolation formula
	zTop := (1.0-dx)*z00 + dx*z10
	zBottom := (1.0-dx)*z01 + dx*z11
	z := (1.0-dy)*zTop + dy*zBottom

	return z, nil
}

// GetElevationAtCoords converts (lat, lon) to projected space and interpolates surface elevation.
func (m *ElevationMatrix) GetElevationAtCoords(latDeg, lonDeg float64) (float64, error) {
	x, y, err := m.Proj.Forward(latDeg, lonDeg)
	if err != nil {
		return 0, err
	}
	return m.InterpolateElevation(x, y)
}

// CalculateSlope computes the surface slope angle in degrees at the given selenographic coordinates.
// Uses central difference gradients over the DEM resolution.
func (m *ElevationMatrix) CalculateSlope(latDeg, lonDeg float64) (float64, error) {
	x, y, err := m.Proj.Forward(latDeg, lonDeg)
	if err != nil {
		return 0, err
	}

	step := m.ResolutionM
	zRight, err := m.InterpolateElevation(x+step, y)
	if err != nil {
		return 0, err
	}
	zLeft, err := m.InterpolateElevation(x-step, y)
	if err != nil {
		return 0, err
	}
	zTop, err := m.InterpolateElevation(x, y+step)
	if err != nil {
		return 0, err
	}
	zBottom, err := m.InterpolateElevation(x, y-step)
	if err != nil {
		return 0, err
	}

	dzdx := (zRight - zLeft) / (2.0 * step)
	dzdy := (zTop - zBottom) / (2.0 * step)

	gradientMag := math.Hypot(dzdx, dzdy)
	slopeDeg := math.Atan(gradientMag) * RadToDeg

	return slopeDeg, nil
}

// NewSyntheticMatrix creates a realistic test DEM containing synthetic craters and massifs
// for unit testing, offline development, and demonstration.
func NewSyntheticMatrix(width, height int, resolutionM float64) *ElevationMatrix {
	m := NewElevationMatrix(width, height, resolutionM, 0, 0)

	// Base elevation: South Pole Aitken basin plateau (~ -2000 m relative to reference sphere)
	baseElev := float64(-2000.0)

	// Synthetic Shackleton crater at center (x=0, y=0, radius ~10.5 km = 10,500m, depth ~4200m, rim elevation +1200m)
	craterRadius := 10500.0
	rimHeight := 1200.0
	floorDepth := -4200.0

	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			x, y := m.GridToWorld(float64(c), float64(r))
			dist := math.Hypot(x, y)

			var elev float64
			if dist < craterRadius {
				// Inside crater bowl: parabolic descent to floor
				factor := dist / craterRadius
				elev = floorDepth + (rimHeight-floorDepth)*math.Pow(factor, 2.0)
			} else if dist < craterRadius*1.5 {
				// Raised rim and outer ejecta blanket decay
				factor := (dist - craterRadius) / (craterRadius * 0.5)
				elev = rimHeight*math.Exp(-factor*2.5) + baseElev*(1.0-math.Exp(-factor*2.5))
			} else {
				// Surrounding undulating terrain with low hills
				elev = baseElev + 150.0*math.Sin(x*1e-4)*math.Cos(y*1e-4)
			}

			// Add a synthetic mountain ridge (like Malapert Massif) to the North-West
			mountainDist := math.Hypot(x+45000, y-35000)
			if mountainDist < 25000 {
				mountainHeight := 4800.0 * math.Exp(-math.Pow(mountainDist/12000.0, 2.0))
				elev += mountainHeight
			}

			m.SetElevation(c, r, float32(elev))
		}
	}

	return m
}
