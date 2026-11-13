package mock

import (
	"context"
	"lunar-navigator/internal/database"
	"testing"
	"time"
)

func TestSeedDatabase(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory db: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Run SeedDatabase
	if err := SeedDatabase(ctx, db); err != nil {
		t.Fatalf("failed to seed database: %v", err)
	}

	// Verify all 13 official Artemis presets exist
	presets, err := db.ListPresets(ctx)
	if err != nil {
		t.Fatalf("failed to list presets: %v", err)
	}
	if len(presets) != 13 {
		t.Errorf("expected 13 Artemis presets, got %d", len(presets))
	}

	// Verify candidate sites were seeded (13 regions * 80 = 1040)
	count, err := db.CountCandidateSites(ctx)
	if err != nil {
		t.Fatalf("failed to count candidate sites: %v", err)
	}
	if count < 1000 {
		t.Errorf("expected at least 1000 candidate sites, got %d", count)
	}

	// Test idempotency: calling SeedDatabase again should not fail or duplicate candidates
	if err := SeedDatabase(ctx, db); err != nil {
		t.Fatalf("second call to SeedDatabase failed: %v", err)
	}
	countAfter, err := db.CountCandidateSites(ctx)
	if err != nil {
		t.Fatalf("failed to count candidate sites after second seed: %v", err)
	}
	if countAfter != count {
		t.Errorf("expected candidate count %d to remain unchanged after second seed, got %d", count, countAfter)
	}
}
