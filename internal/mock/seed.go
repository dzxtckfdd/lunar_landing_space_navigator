package mock

import (
	"context"
	"fmt"
	"lunar-navigator/internal/database"
	"lunar-navigator/internal/domain"
	"math"
	"math/rand"
)

// Official 13 Artemis III Candidate Regions defined by NASA
var ArtemisPresets = []domain.Preset{
	{
		ID:                     "shackleton-connecting-ridge",
		Name:                   "Shackleton Connecting Ridge",
		SubRegion:              "South Pole Peak",
		Latitude:               -89.45,
		Longitude:              120.12,
		ElevationM:             1240.0,
		AverageSlopeDeg:        4.5,
		DistanceToPSRkm:        0.8,
		Description:            "Elevated ridge between Shackleton and de Gerlache craters. Features high cumulative illumination.",
		ScientificSignificance: "Immediate access to cryogenic volatiles in Shackleton interior.",
	},
	{
		ID:                     "malapert-mountain",
		Name:                   "Malapert Mountain",
		SubRegion:              "Malapert Massif",
		Latitude:               -86.04,
		Longitude:              2.71,
		ElevationM:             5050.0,
		AverageSlopeDeg:        5.8,
		DistanceToPSRkm:        3.2,
		Description:            "Massive volcanic/impact uplift plateau. One of the prime candidates for Peak of Eternal Light.",
		ScientificSignificance: "Ideal communications vantage point and regional geology stratigraphy.",
	},
	{
		ID:                     "faustini-rim-a",
		Name:                   "Faustini Rim A",
		SubRegion:              "Faustini Crater Rim",
		Latitude:               -87.15,
		Longitude:              84.30,
		ElevationM:             2100.0,
		AverageSlopeDeg:        6.1,
		DistanceToPSRkm:        1.4,
		Description:            "Sunlit ridge overlooking deep permanently shadowed cold traps in Faustini.",
		ScientificSignificance: "Direct exploration of ancient pre-Nectarian volatile reservoirs.",
	},
	{
		ID:                     "de-gerlache-rim-1",
		Name:                   "de Gerlache Rim 1",
		SubRegion:              "de Gerlache Crater",
		Latitude:               -88.50,
		Longitude:              -88.20,
		ElevationM:             1800.0,
		AverageSlopeDeg:        5.2,
		DistanceToPSRkm:        1.1,
		Description:            "Broad ridge with consistent seasonal illumination and gentle descent slopes into the crater.",
		ScientificSignificance: "Impact melt sheets and volatile stratigraphy.",
	},
	{
		ID:                     "de-gerlache-rim-2",
		Name:                   "de Gerlache Rim 2",
		SubRegion:              "de Gerlache Crater",
		Latitude:               -88.35,
		Longitude:              -87.50,
		ElevationM:             1750.0,
		AverageSlopeDeg:        6.8,
		DistanceToPSRkm:        1.6,
		Description:            "Northern boundary plateau of de Gerlache crater.",
		ScientificSignificance: "Surface roughness boundary analysis and volatile tracking.",
	},
	{
		ID:                     "haworth-crater-rim",
		Name:                   "Haworth Crater Rim",
		SubRegion:              "Haworth Crater",
		Latitude:               -87.40,
		Longitude:              -5.10,
		ElevationM:             2400.0,
		AverageSlopeDeg:        7.2,
		DistanceToPSRkm:        2.1,
		Description:            "Rugged high-elevation rim with expansive line-of-sight to Earth.",
		ScientificSignificance: "Investigation of ultra-cold (<30K) organic compounds in shadow.",
	},
	{
		ID:                     "nobile-crater-rim-1",
		Name:                   "Nobile Crater Rim 1",
		SubRegion:              "Nobile Crater",
		Latitude:               -85.20,
		Longitude:              53.50,
		ElevationM:             3200.0,
		AverageSlopeDeg:        4.9,
		DistanceToPSRkm:        2.5,
		Description:            "Broad ridge selected for VIPER rover target science, moderate slopes.",
		ScientificSignificance: "Direct rover traversability into volatile-rich terrain.",
	},
	{
		ID:                     "nobile-crater-rim-2",
		Name:                   "Nobile Crater Rim 2",
		SubRegion:              "Nobile Crater",
		Latitude:               -85.35,
		Longitude:              54.80,
		ElevationM:             3150.0,
		AverageSlopeDeg:        5.4,
		DistanceToPSRkm:        2.8,
		Description:            "Secondary ridge flanking the eastern wall of Nobile.",
		ScientificSignificance: "Lateral heterogeneity of lunar polar regolith.",
	},
	{
		ID:                     "nobile-ridge",
		Name:                   "Nobile Ridge",
		SubRegion:              "Nobile Complex",
		Latitude:               -85.50,
		Longitude:              48.20,
		ElevationM:             2900.0,
		AverageSlopeDeg:        5.1,
		DistanceToPSRkm:        1.9,
		Description:            "High elevation traverse pathway connecting Nobile and Amundsen.",
		ScientificSignificance: "Multi-crater ejecta blanket cross-section sampling.",
	},
	{
		ID:                     "amundsen-rim",
		Name:                   "Amundsen Rim",
		SubRegion:              "Amundsen Basin",
		Latitude:               -84.50,
		Longitude:              82.80,
		ElevationM:             3600.0,
		AverageSlopeDeg:        7.8,
		DistanceToPSRkm:        3.5,
		Description:            "High mountain rim of the giant Amundsen impact basin.",
		ScientificSignificance: "Deep crustal lunar material brought up from South Pole-Aitken basin.",
	},
	{
		ID:                     "leibniz-beta-plateau",
		Name:                   "Leibniz Beta Plateau",
		SubRegion:              "Leibniz Beta Massif",
		Latitude:               -85.80,
		Longitude:              23.60,
		ElevationM:             4200.0,
		AverageSlopeDeg:        5.9,
		DistanceToPSRkm:        2.9,
		Description:            "Elevated mesa-like plateau offering long-duration direct-to-Earth communication.",
		ScientificSignificance: "Ancient highland crust intact exposures.",
	},
	{
		ID:                     "connecting-ridge-extension",
		Name:                   "Connecting Ridge Extension",
		SubRegion:              "South Pole Ridge",
		Latitude:               -89.20,
		Longitude:              118.50,
		ElevationM:             1150.0,
		AverageSlopeDeg:        4.2,
		DistanceToPSRkm:        0.9,
		Description:            "Northern spur of the Shackleton connecting ridge system.",
		ScientificSignificance: "Ultra-close proximity to multiple micro-cold traps.",
	},
	{
		ID:                     "peak-near-shackleton",
		Name:                   "Peak near Shackleton",
		SubRegion:              "Shackleton Massif",
		Latitude:               -89.70,
		Longitude:              130.40,
		ElevationM:             1400.0,
		AverageSlopeDeg:        6.5,
		DistanceToPSRkm:        0.5,
		Description:            "Immediate rim peak closest to true lunar South Pole (-90°).",
		ScientificSignificance: "Historic polar landing target with near-continuous illumination.",
	},
}

// SeedDatabase populates presets and generates a dense candidate grid for the recommender engine.
func SeedDatabase(ctx context.Context, db *database.DB) error {
	// 1. Seed Presets
	for _, p := range ArtemisPresets {
		if err := db.InsertPreset(ctx, &p); err != nil {
			return fmt.Errorf("failed to seed preset %s: %w", p.ID, err)
		}
	}

	// 2. Check if candidate sites already exist
	count, err := db.CountCandidateSites(ctx)
	if err != nil {
		return fmt.Errorf("failed to check candidate sites count: %w", err)
	}
	if count > 0 {
		return nil // Already seeded
	}

	// 3. Generate candidate points around each of the 13 Artemis regions
	var candidates []domain.CandidateSite
	rng := rand.New(rand.NewSource(42)) // Deterministic seed for reproducible testing

	for _, preset := range ArtemisPresets {
		// Generate 80 candidate landing spots clustered around each preset center (approx ~1040 total candidates)
		for i := 0; i < 80; i++ {
			// Offset by small delta (-0.25 to +0.25 deg latitude, -1.5 to +1.5 deg longitude)
			latOffset := (rng.Float64() - 0.5) * 0.4
			lonOffset := (rng.Float64() - 0.5) * 3.0

			lat := math.Max(-89.9, math.Min(-80.0, preset.Latitude+latOffset))
			lon := preset.Longitude + lonOffset
			if lon > 180 {
				lon -= 360
			} else if lon < -180 {
				lon += 360
			}

			// Varied realistic slopes (3.0 to 14.0 deg) and distances to PSR (0.3 to 4.5 km)
			slope := math.Max(2.5, preset.AverageSlopeDeg+(rng.Float64()-0.5)*4.0)
			distPSR := math.Max(0.2, preset.DistanceToPSRkm+(rng.Float64()-0.5)*1.5)
			elev := preset.ElevationM + (rng.Float64()-0.5)*300.0

			candidates = append(candidates, domain.CandidateSite{
				ClusterRegion:   preset.Name,
				Latitude:        math.Round(lat*1000) / 1000,
				Longitude:       math.Round(lon*1000) / 1000,
				ElevationM:      math.Round(elev*10) / 10,
				SlopeDeg:        math.Round(slope*10) / 10,
				DistanceToPSRkm: math.Round(distPSR*100) / 100,
				IsSafeSlope:     slope <= 10.0,
			})
		}
	}

	// Batch insert candidates
	if err := db.InsertCandidateSitesBatch(ctx, candidates); err != nil {
		return fmt.Errorf("failed to batch insert candidate sites: %w", err)
	}

	return nil
}
