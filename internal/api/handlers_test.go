package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"lunar-navigator/internal/astro"
	"lunar-navigator/internal/config"
	"lunar-navigator/internal/database"
	"lunar-navigator/internal/dem"
	"lunar-navigator/internal/domain"
	"lunar-navigator/internal/evaluation"
	"lunar-navigator/internal/mock"
	"lunar-navigator/internal/recommender"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupTestServer(t *testing.T) (*httptest.Server, func()) {
	cfg := &config.Config{
		Port:         8080,
		DBPath:       ":memory:",
		MockMode:     true,
		DefaultMastH: 2.5,
		FreezeHours:  36,
	}

	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	if err := mock.SeedDatabase(ctx, db); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	matrix := dem.NewSyntheticMatrix(2600, 2600, 100.0)
	raycaster := dem.NewRaycaster(matrix)
	ephem := astro.NewEphemerisService()
	evalService := evaluation.NewService(matrix, raycaster, ephem)
	recService := recommender.NewService(db, evalService)

	handler := NewHandler(cfg, db, matrix, raycaster, evalService, recService)
	router := NewRouter(handler)
	server := httptest.NewServer(router)

	cleanup := func() {
		server.Close()
		db.Close()
	}

	return server, cleanup
}

func TestHealthEndpoint(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Get(server.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}

	if health.Status != "ready" {
		t.Errorf("expected status 'ready', got '%s'", health.Status)
	}
	if health.Datasets["candidate_sites_count"].(float64) < 1000 {
		t.Errorf("expected at least 1000 candidate sites, got %v", health.Datasets["candidate_sites_count"])
	}
}

func TestPresetsEndpoints(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. List Presets
	resp, err := http.Get(server.URL + "/api/v1/presets")
	if err != nil {
		t.Fatalf("GET /presets failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var presets []domain.Preset
	if err := json.NewDecoder(resp.Body).Decode(&presets); err != nil {
		t.Fatalf("failed to decode presets: %v", err)
	}
	if len(presets) != 13 {
		t.Errorf("expected 13 presets, got %d", len(presets))
	}

	// 2. Get Single Preset
	respSingle, err := http.Get(server.URL + "/api/v1/presets/malapert-mountain")
	if err != nil {
		t.Fatalf("GET /presets/malapert-mountain failed: %v", err)
	}
	defer respSingle.Body.Close()

	if respSingle.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", respSingle.StatusCode)
	}
	var preset domain.Preset
	_ = json.NewDecoder(respSingle.Body).Decode(&preset)
	if preset.Name != "Malapert Mountain" {
		t.Errorf("expected 'Malapert Mountain', got '%s'", preset.Name)
	}

	// 3. Get Nonexistent Preset -> 404
	resp404, _ := http.Get(server.URL + "/api/v1/presets/nonexistent-crater")
	defer resp404.Body.Close()
	if resp404.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for nonexistent preset, got %d", resp404.StatusCode)
	}
}

func TestHorizonEndpoint(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Valid Request
	body := []byte(`{"latitude": -89.45, "longitude": 120.12, "mast_height_m": 2.5}`)
	resp, err := http.Post(server.URL+"/api/v1/horizon", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("POST /horizon failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d. Body: %s", resp.StatusCode, string(b))
	}

	var profile dem.HorizonProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		t.Fatalf("failed to decode horizon profile: %v", err)
	}
	if len(profile.Points) != 360 {
		t.Errorf("expected 360 horizon points, got %d", len(profile.Points))
	}

	// 2. Invalid Latitude (Out of bounds) -> 422
	badBody := []byte(`{"latitude": -50.0, "longitude": 0.0}`)
	respBad, _ := http.Post(server.URL+"/api/v1/horizon", "application/json", bytes.NewBuffer(badBody))
	defer respBad.Body.Close()

	if respBad.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422 for lat -50.0, got %d", respBad.StatusCode)
	}
}

func TestEvaluateEndpoint(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	evalPayload := []byte(`{
		"latitude": -86.04,
		"longitude": 2.71,
		"mast_height_m": 2.5,
		"mission_start_utc": "2026-11-15T00:00:00Z",
		"duration_hours": 48,
		"launch_slip_days": 0,
		"solar_panel_type": "vertical_cylindrical",
		"max_shadow_freeze_hours": 36
	}`)

	resp, err := http.Post(server.URL+"/api/v1/evaluate", "application/json", bytes.NewBuffer(evalPayload))
	if err != nil {
		t.Fatalf("POST /evaluate failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d. Body: %s", resp.StatusCode, string(b))
	}

	var result evaluation.Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode evaluation result: %v", err)
	}

	if result.Summary.TotalHours != 49 { // 48 hours + offset 0 = 49 steps
		t.Errorf("expected 49 steps, got %d", result.Summary.TotalHours)
	}
	if result.Summary.FeasibilityScore <= 0 {
		t.Errorf("expected positive feasibility score, got %f", result.Summary.FeasibilityScore)
	}
}

func TestRecommendEndpoint(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	recPayload := []byte(`{
		"min_solar_illumination_pct": 30.0,
		"min_dte_comm_pct": 20.0,
		"max_slope_deg": 10.0,
		"max_distance_to_psr_km": 4.0,
		"duration_hours": 72,
		"limit": 5
	}`)

	resp, err := http.Post(server.URL+"/api/v1/recommend", "application/json", bytes.NewBuffer(recPayload))
	if err != nil {
		t.Fatalf("POST /recommend failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d. Body: %s", resp.StatusCode, string(b))
	}

	var recResp recommender.Response
	if err := json.NewDecoder(resp.Body).Decode(&recResp); err != nil {
		t.Fatalf("failed to decode recommend response: %v", err)
	}

	if len(recResp.Recommendations) != 5 {
		t.Errorf("expected 5 recommendations, got %d", len(recResp.Recommendations))
	}
	if recResp.Recommendations[0].Rank != 1 {
		t.Errorf("expected Rank 1 for first item, got %d", recResp.Recommendations[0].Rank)
	}
}

func TestScenarioSaveAndGet(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Save Scenario
	scenarioPayload := []byte(`{
		"title": "Artemis III Test Run",
		"latitude": -89.45,
		"longitude": 120.12,
		"antenna_mast_height_m": 2.5,
		"mission_start_utc": "2026-11-15T00:00:00Z",
		"mission_duration_hours": 168,
		"solar_illumination_pct": 91.5,
		"dte_comm_pct": 75.0,
		"feasibility_score": 92.4,
		"notes": "E2E API Test"
	}`)

	resp, err := http.Post(server.URL+"/api/v1/scenarios", "application/json", bytes.NewBuffer(scenarioPayload))
	if err != nil {
		t.Fatalf("POST /scenarios failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", resp.StatusCode)
	}

	var saveResp SaveScenarioResponse
	_ = json.NewDecoder(resp.Body).Decode(&saveResp)
	if saveResp.ID == "" {
		t.Fatalf("expected non-empty scenario ID")
	}

	// 2. Get Saved Scenario by ID
	respGet, err := http.Get(server.URL + "/api/v1/scenarios/" + saveResp.ID)
	if err != nil {
		t.Fatalf("GET /scenarios/:id failed: %v", err)
	}
	defer respGet.Body.Close()

	if respGet.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", respGet.StatusCode)
	}

	var fetched domain.SavedScenario
	_ = json.NewDecoder(respGet.Body).Decode(&fetched)
	if fetched.Title != "Artemis III Test Run" {
		t.Errorf("expected title 'Artemis III Test Run', got '%s'", fetched.Title)
	}
	if fetched.FeasibilityScore != 92.4 {
		t.Errorf("expected score 92.4, got %f", fetched.FeasibilityScore)
	}
}

func TestCORSPreflight(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req, _ := http.NewRequest(http.MethodOptions, server.URL+"/api/v1/health", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS preflight failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204 No Content for CORS OPTIONS, got %d", resp.StatusCode)
	}
	if allowOrigin := resp.Header.Get("Access-Control-Allow-Origin"); allowOrigin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin '*', got '%s'", allowOrigin)
	}
}

func TestStaticFrontendServing(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	bodyStr := string(body)
	if !bytes.Contains(body, []byte("Lunar Landing Site Navigator")) && !bytes.Contains(body, []byte("Lunar Navigator by Gemini")) {
		t.Errorf("expected HTML containing app title, got: %s", bodyStr[:min(200, len(bodyStr))])
	}
}
