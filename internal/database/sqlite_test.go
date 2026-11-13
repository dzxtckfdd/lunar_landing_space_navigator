package database

import (
	"context"
	"lunar-navigator/internal/domain"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *DB {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	return db
}

func TestDatabaseMigrationAndPresetsCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	// 1. Insert Preset
	preset := &domain.Preset{
		ID:                     "test-shackleton",
		Name:                   "Shackleton Test Peak",
		SubRegion:              "Peak A",
		Latitude:               -89.5,
		Longitude:              120.0,
		ElevationM:             1500.0,
		AverageSlopeDeg:        4.2,
		DistanceToPSRkm:        0.9,
		Description:            "Test description for Shackleton",
		ScientificSignificance: "High ice volatile potential",
	}

	if err := db.InsertPreset(ctx, preset); err != nil {
		t.Fatalf("failed to insert preset: %v", err)
	}

	// 2. Get Preset
	fetched, err := db.GetPreset(ctx, "test-shackleton")
	if err != nil {
		t.Fatalf("failed to get preset: %v", err)
	}
	if fetched == nil || fetched.Name != "Shackleton Test Peak" {
		t.Fatalf("expected 'Shackleton Test Peak', got %+v", fetched)
	}

	// 3. List Presets
	list, err := db.ListPresets(ctx)
	if err != nil {
		t.Fatalf("failed to list presets: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 preset in list, got %d", len(list))
	}

	// 4. Upsert (update)
	preset.Description = "Updated description"
	if err := db.InsertPreset(ctx, preset); err != nil {
		t.Fatalf("failed to update preset: %v", err)
	}
	updated, err := db.GetPreset(ctx, "test-shackleton")
	if err != nil {
		t.Fatalf("failed to get updated preset: %v", err)
	}
	if updated.Description != "Updated description" {
		t.Fatalf("expected updated description, got %s", updated.Description)
	}
}

func TestCandidateSitesBatchAndFiltering(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	candidates := []domain.CandidateSite{
		{
			ClusterRegion:   "Malapert",
			Latitude:        -86.0,
			Longitude:       2.5,
			ElevationM:      5000,
			SlopeDeg:        4.0,
			DistanceToPSRkm: 1.5,
			IsSafeSlope:     true,
		},
		{
			ClusterRegion:   "Malapert",
			Latitude:        -86.1,
			Longitude:       2.6,
			ElevationM:      4800,
			SlopeDeg:        12.5, // Hazardous slope
			DistanceToPSRkm: 2.0,
			IsSafeSlope:     false,
		},
		{
			ClusterRegion:   "Shackleton",
			Latitude:        -89.5,
			Longitude:       120.0,
			ElevationM:      1200,
			SlopeDeg:        5.0,
			DistanceToPSRkm: 0.5,
			IsSafeSlope:     true,
		},
	}

	if err := db.InsertCandidateSitesBatch(ctx, candidates); err != nil {
		t.Fatalf("failed to batch insert candidates: %v", err)
	}

	count, err := db.CountCandidateSites(ctx)
	if err != nil {
		t.Fatalf("failed to count candidates: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 candidates, got %d", count)
	}

	// Filter max slope <= 6.0 and distance <= 1.0 km
	filtered, err := db.ListFilteredCandidateSites(ctx, 6.0, 1.0, "all")
	if err != nil {
		t.Fatalf("failed to filter candidates: %v", err)
	}
	if len(filtered) != 1 || filtered[0].ClusterRegion != "Shackleton" {
		t.Fatalf("expected 1 Shackleton candidate, got %+v", filtered)
	}
}

func TestSavedScenarios(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	scenario := &domain.SavedScenario{
		ID:                   "scn_test123",
		Title:                "Test Artemis Scenario",
		Latitude:             -89.45,
		Longitude:            120.12,
		AntennaMastHeightM:   2.5,
		MissionStartUTC:      time.Now().UTC().Truncate(time.Second),
		MissionDurationHours: 336,
		LaunchSlipDays:       2,
		SolarIlluminationPct: 92.5,
		DTECommPct:           78.0,
		FeasibilityScore:     91.2,
		Notes:                "Unit test scenario record",
	}

	if err := db.SaveScenario(ctx, scenario); err != nil {
		t.Fatalf("failed to save scenario: %v", err)
	}

	fetched, err := db.GetScenario(ctx, "scn_test123")
	if err != nil {
		t.Fatalf("failed to get scenario: %v", err)
	}
	if fetched == nil || fetched.Title != "Test Artemis Scenario" {
		t.Fatalf("expected 'Test Artemis Scenario', got %+v", fetched)
	}
	if fetched.AntennaMastHeightM != 2.5 {
		t.Fatalf("expected mast height 2.5, got %f", fetched.AntennaMastHeightM)
	}
}
