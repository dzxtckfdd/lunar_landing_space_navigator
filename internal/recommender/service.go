package recommender

import (
	"context"
	"fmt"
	"lunar-navigator/internal/database"
	"lunar-navigator/internal/domain"
	"lunar-navigator/internal/evaluation"
	"sort"
	"sync"
	"time"
)

// Service coordinates the spatial filtering and multi-objective ranking of candidate landing sites.
type Service struct {
	db          *database.DB
	evalService *evaluation.Service
	workerCount int
}

// NewService creates a new recommender service instance.
func NewService(db *database.DB, evalService *evaluation.Service) *Service {
	return &Service{
		db:          db,
		evalService: evalService,
		workerCount: 8, // Multi-core worker pool
	}
}

// Recommend queries the candidate database and ranks the Top-N landing sites based on user criteria.
func (s *Service) Recommend(ctx context.Context, req Request) (*Response, error) {
	startTimer := time.Now()

	// 1. Sanitize defaults
	if req.Limit <= 0 {
		req.Limit = 5
	}
	if req.MaxSlopeDeg <= 0 {
		req.MaxSlopeDeg = 10.0 // Default maximum safe slope: 10°
	}
	if req.MaxDistanceToPSRkm <= 0 {
		req.MaxDistanceToPSRkm = 5.0 // Default search radius to PSR: 5 km
	}
	if req.MastHeightM <= 0 {
		req.MastHeightM = 2.5
	}
	if req.MaxShadowFreezeHours <= 0 {
		req.MaxShadowFreezeHours = 36
	}
	if req.DurationHours <= 0 {
		req.DurationHours = 336 // Default 14 days
	}
	if req.MissionStartUTC.IsZero() {
		req.MissionStartUTC = time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
	}

	// 2. Query Candidate Sites meeting Hard Constraints (Slope & Distance to PSR)
	candidates, err := s.db.ListFilteredCandidateSites(
		ctx, req.MaxSlopeDeg, req.MaxDistanceToPSRkm, req.PreferredRegion,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query filtered candidate sites: %w", err)
	}

	totalScanned := len(candidates)
	if totalScanned == 0 {
		return &Response{
			CandidatesScanned: 0,
			MatchesFound:      0,
			Recommendations:   []RecommendedSite{},
			ComputationTimeMs: float64(time.Since(startTimer).Microseconds()) / 1000.0,
		}, nil
	}

	// 3. Parallel Evaluation of Candidates across Worker Goroutines
	type evaluatedCandidate struct {
		site   domain.CandidateSite
		eval   *evaluation.Result
		passed bool
	}

	resultsChan := make(chan evaluatedCandidate, totalScanned)
	jobsChan := make(chan domain.CandidateSite, totalScanned)

	var wg sync.WaitGroup
	for w := 0; w < s.workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for cand := range jobsChan {
				evalReq := evaluation.Request{
					Latitude:             cand.Latitude,
					Longitude:            cand.Longitude,
					MastHeightM:          req.MastHeightM,
					MissionStartUTC:      req.MissionStartUTC,
					DurationHours:        req.DurationHours,
					LaunchSlipDays:       req.LaunchSlipDays,
					SolarPanelType:       req.SolarPanelType,
					MaxShadowFreezeHours: req.MaxShadowFreezeHours,
					Weights:              evaluation.DefaultWeights(),
				}

				res, err := s.evalService.Evaluate(evalReq)
				if err != nil {
					continue
				}

				// Check soft criteria thresholds
				passed := res.Summary.SolarIlluminationPct >= req.MinSolarIlluminationPct &&
					res.Summary.DTECommPct >= req.MinDTECommPct &&
					!res.Summary.ThermalFailureTriggered

				resultsChan <- evaluatedCandidate{
					site:   cand,
					eval:   res,
					passed: passed,
				}
			}
		}()
	}

	// Send jobs
	for _, cand := range candidates {
		jobsChan <- cand
	}
	close(jobsChan)

	wg.Wait()
	close(resultsChan)

	// Collect evaluated candidates
	var passedList []evaluatedCandidate
	var allEvaluated []evaluatedCandidate

	for ec := range resultsChan {
		allEvaluated = append(allEvaluated, ec)
		if ec.passed {
			passedList = append(passedList, ec)
		}
	}

	// 4. Sort Candidates by Feasibility Score descending
	sortSlice := func(slice []evaluatedCandidate) {
		sort.Slice(slice, func(i, j int) bool {
			return slice[i].eval.Summary.FeasibilityScore > slice[j].eval.Summary.FeasibilityScore
		})
	}

	sortSlice(passedList)
	sortSlice(allEvaluated)

	// Selection list: prefer candidates that passed all filters, fall back to best available if needed
	var selected []evaluatedCandidate
	if len(passedList) >= req.Limit {
		selected = passedList[:req.Limit]
	} else if len(passedList) > 0 {
		selected = passedList
	} else {
		// If no site strictly met all high thresholds, pick top available
		topN := req.Limit
		if len(allEvaluated) < topN {
			topN = len(allEvaluated)
		}
		selected = allEvaluated[:topN]
	}

	// 5. Construct Recommended Site objects with Badges
	recommendations := make([]RecommendedSite, len(selected))
	for i, item := range selected {
		rank := i + 1
		badge := assignBadge(rank, item.site, item.eval)

		title := fmt.Sprintf("%s - Landing Spot #%d", item.site.ClusterRegion, rank)

		recommendations[i] = RecommendedSite{
			Rank:                    rank,
			Title:                   title,
			ClusterRegion:           item.site.ClusterRegion,
			Latitude:                item.site.Latitude,
			Longitude:               item.site.Longitude,
			ElevationM:              item.site.ElevationM,
			SlopeDeg:                item.site.SlopeDeg,
			DistanceToPSRkm:         item.site.DistanceToPSRkm,
			SolarIlluminationPct:    item.eval.Summary.SolarIlluminationPct,
			ContinuousSunlightHours: item.eval.Summary.ContinuousSunlightMaxHours,
			DTECommPct:              item.eval.Summary.DTECommPct,
			LongestBlackoutHours:    item.eval.Summary.LongestCommBlackoutHours,
			FeasibilityScore:        item.eval.Summary.FeasibilityScore,
			Badge:                   badge,
		}
	}

	elapsedMs := float64(time.Since(startTimer).Microseconds()) / 1000.0

	return &Response{
		CandidatesScanned: totalScanned,
		MatchesFound:      len(passedList),
		Recommendations:   recommendations,
		ComputationTimeMs: elapsedMs,
	}, nil
}

// assignBadge generates a contextual badge highlighting the unique merit of each candidate.
func assignBadge(rank int, site domain.CandidateSite, res *evaluation.Result) string {
	if rank == 1 {
		return "Top Recommended - Optimal Power & Comms"
	}
	if site.DistanceToPSRkm <= 1.0 {
		return "Prime Water Ice Reservoir Access"
	}
	if res.Summary.SolarIlluminationPct >= 80.0 {
		return "Maximum Solar Sunlight Exposure"
	}
	if res.Summary.DTECommPct >= 70.0 {
		return "High Direct-to-Earth Link Window"
	}
	if site.SlopeDeg <= 5.0 {
		return "Ultra-Safe Flat Terrain"
	}
	return fmt.Sprintf("Artemis Alternate #%d", rank)
}
