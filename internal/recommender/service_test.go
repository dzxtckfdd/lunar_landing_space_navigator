package recommender

import (
	"context"
	"lunar-navigator/internal/astro"
	"lunar-navigator/internal/database"
	"lunar-navigator/internal/dem"
	"lunar-navigator/internal/evaluation"
	"lunar-navigator/internal/mock"
	"testing"
	"time"
)

func setupTestRecommenderService(t *testing.T) (*Service, func()) {
	// 1. Setup in-memory SQLite database
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Seed Artemis presets and ~1040 candidate sites
	if err := mock.SeedDatabase(ctx, db); err != nil {
		t.Fatalf("failed to seed database: %v", err)
	}

	// 2. Setup DEM, Raycaster and Ephemeris
	matrix := dem.NewSyntheticMatrix(2600, 2600, 100.0)
	raycaster := dem.NewRaycaster(matrix)
	ephem := astro.NewEphemerisService()
	evalService := evaluation.NewService(matrix, raycaster, ephem)

	svc := NewService(db, evalService)
	cleanup := func() {
		db.Close()
	}

	return svc, cleanup
}

func TestRecommendTop5StandardCriteria(t *testing.T) {
	svc, cleanup := setupTestRecommenderService(t)
	defer cleanup()

	ctx := context.Background()

	req := Request{
		MinSolarIlluminationPct: 30.0,
		MinDTECommPct:           20.0,
		MaxSlopeDeg:             10.0,
		MaxDistanceToPSRkm:      3.5,
		MissionStartUTC:         time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC),
		DurationHours:           168, // 7 days
		LaunchSlipDays:          0,
		Limit:                   5,
		PreferredRegion:         "all",
	}

	res, err := svc.Recommend(ctx, req)
	if err != nil {
		t.Fatalf("Recommend failed: %v", err)
	}

	if res.CandidatesScanned == 0 {
		t.Fatalf("expected candidates to be scanned, got 0")
	}

	if len(res.Recommendations) != 5 {
		t.Fatalf("expected 5 recommendations, got %d", len(res.Recommendations))
	}

	// Check sorted in descending order of feasibility score
	for i := 0; i < len(res.Recommendations)-1; i++ {
		current := res.Recommendations[i]
		next := res.Recommendations[i+1]

		if current.Rank != i+1 {
			t.Errorf("expected rank %d, got %d", i+1, current.Rank)
		}
		if current.FeasibilityScore < next.FeasibilityScore {
			t.Errorf("recommendations not sorted: #%d (score %.1f) < #%d (score %.1f)",
				current.Rank, current.FeasibilityScore, next.Rank, next.FeasibilityScore)
		}
	}

	// First recommendation should have Top Recommended badge
	if res.Recommendations[0].Badge != "Top Recommended - Optimal Power & Comms" {
		t.Errorf("expected Rank 1 badge, got: %s", res.Recommendations[0].Badge)
	}

	t.Logf("Recommender: Scanned=%d, Matches=%d, Top 1 Score=%.1f (%s), Time=%.2f ms",
		res.CandidatesScanned, res.MatchesFound, res.Recommendations[0].FeasibilityScore,
		res.Recommendations[0].Title, res.ComputationTimeMs)
}

func TestRecommendFilterBySpecificRegion(t *testing.T) {
	svc, cleanup := setupTestRecommenderService(t)
	defer cleanup()

	ctx := context.Background()
	targetRegion := "Shackleton Connecting Ridge"

	req := Request{
		MinSolarIlluminationPct: 10.0,
		MinDTECommPct:           10.0,
		MaxSlopeDeg:             12.0,
		MaxDistanceToPSRkm:      4.0,
		MissionStartUTC:         time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC),
		DurationHours:           168,
		PreferredRegion:         targetRegion,
		Limit:                   3,
	}

	res, err := svc.Recommend(ctx, req)
	if err != nil {
		t.Fatalf("Recommend failed: %v", err)
	}

	if len(res.Recommendations) == 0 {
		t.Fatalf("expected recommendations for %s, got 0", targetRegion)
	}

	// Verify all returned recommendations belong exclusively to the requested cluster region
	for _, rec := range res.Recommendations {
		if rec.ClusterRegion != targetRegion {
			t.Errorf("expected cluster %s, got %s", targetRegion, rec.ClusterRegion)
		}
	}
}

func TestRecommendNoMatchingSites(t *testing.T) {
	svc, cleanup := setupTestRecommenderService(t)
	defer cleanup()

	ctx := context.Background()

	// Impossible criteria: slope <= 0.01 degrees and distance <= 0.01 km
	req := Request{
		MaxSlopeDeg:        0.01,
		MaxDistanceToPSRkm: 0.01,
		Limit:              5,
	}

	res, err := svc.Recommend(ctx, req)
	if err != nil {
		t.Fatalf("Recommend failed: %v", err)
	}

	if res.CandidatesScanned != 0 {
		t.Errorf("expected 0 scanned candidates for impossible slope, got %d", res.CandidatesScanned)
	}
	if len(res.Recommendations) != 0 {
		t.Errorf("expected 0 recommendations, got %d", len(res.Recommendations))
	}
}

func BenchmarkRecommendTop5(b *testing.B) {
	db, _ := database.New(":memory:")
	defer db.Close()
	ctx := context.Background()
	_ = db.Migrate(ctx)
	_ = mock.SeedDatabase(ctx, db)

	matrix := dem.NewSyntheticMatrix(2600, 2600, 100.0)
	raycaster := dem.NewRaycaster(matrix)
	ephem := astro.NewEphemerisService()
	evalService := evaluation.NewService(matrix, raycaster, ephem)
	svc := NewService(db, evalService)

	req := Request{
		MinSolarIlluminationPct: 30.0,
		MinDTECommPct:           20.0,
		MaxSlopeDeg:             10.0,
		MaxDistanceToPSRkm:      3.0,
		MissionStartUTC:         time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC),
		DurationHours:           168,
		Limit:                   5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.Recommend(ctx, req)
		if err != nil {
			b.Fatalf("benchmark recommend failed: %v", err)
		}
	}
}
