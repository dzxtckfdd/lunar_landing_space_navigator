package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"lunar-navigator/internal/config"
	"lunar-navigator/internal/database"
	"lunar-navigator/internal/dem"
	"lunar-navigator/internal/domain"
	"lunar-navigator/internal/evaluation"
	"lunar-navigator/internal/recommender"
	"net/http"
	"runtime"
	"time"
)

// Handler holds dependencies required for servicing HTTP requests.
type Handler struct {
	cfg         *config.Config
	db          *database.DB
	demMatrix   *dem.ElevationMatrix
	raycaster   *dem.Raycaster
	evalService *evaluation.Service
	recService  *recommender.Service
	startTime   time.Time
}

// NewHandler creates a new handler container.
func NewHandler(
	cfg *config.Config,
	db *database.DB,
	demMatrix *dem.ElevationMatrix,
	raycaster *dem.Raycaster,
	evalService *evaluation.Service,
	recService *recommender.Service,
) *Handler {
	return &Handler{
		cfg:         cfg,
		db:          db,
		demMatrix:   demMatrix,
		raycaster:   raycaster,
		evalService: evalService,
		recService:  recService,
		startTime:   time.Now(),
	}
}

// HealthResponse represents system health and resource telemetry.
type HealthResponse struct {
	Status            string                 `json:"status"`
	Version           string                 `json:"version"`
	UptimeSeconds     int64                  `json:"uptime_seconds"`
	MemoryAllocatedMB float64                `json:"memory_allocated_mb"`
	Datasets          map[string]interface{} `json:"datasets"`
}

// HandleHealth reports the operational status of the service and loaded datasets.
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	candidateCount, _ := h.db.CountCandidateSites(r.Context())

	resp := HealthResponse{
		Status:            "ready",
		Version:           "1.0.0",
		UptimeSeconds:     int64(time.Since(h.startTime).Seconds()),
		MemoryAllocatedMB: float64(m.Alloc) / (1024.0 * 1024.0),
		Datasets: map[string]interface{}{
			"dem_grid_loaded":       h.demMatrix != nil,
			"dem_resolution_m":      h.demMatrix.ResolutionM,
			"ephemeris_range_years": "2024-2035",
			"candidate_sites_count": candidateCount,
		},
	}

	WriteJSON(w, http.StatusOK, resp)
}

// HandleListPresets returns all official Artemis candidate landing regions.
func (h *Handler) HandleListPresets(w http.ResponseWriter, r *http.Request) {
	presets, err := h.db.ListPresets(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "ERR_DB_QUERY", "Database Error", err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, presets)
}

// HandleGetPreset returns a single preset by ID.
func (h *Handler) HandleGetPreset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, r, http.StatusBadRequest, "ERR_INVALID_ID", "Bad Request", "Missing preset ID.")
		return
	}

	preset, err := h.db.GetPreset(r.Context(), id)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "ERR_DB_QUERY", "Database Error", err.Error())
		return
	}
	if preset == nil {
		WriteError(w, r, http.StatusNotFound, "ERR_PRESET_NOT_FOUND", "Not Found", fmt.Sprintf("Preset '%s' not found.", id))
		return
	}

	WriteJSON(w, http.StatusOK, preset)
}

// HorizonRequest defines the payload for calculating local terrain horizon.
type HorizonRequest struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	MastHeightM float64 `json:"mast_height_m"`
}

// HandleHorizon calculates the 360-degree local horizon elevation mask around a point.
func (h *Handler) HandleHorizon(w http.ResponseWriter, r *http.Request) {
	var req HorizonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, http.StatusBadRequest, "ERR_INVALID_JSON", "Bad Request", "Malformed JSON body.")
		return
	}

	if req.Latitude < -90.0 || req.Latitude > -75.0 {
		WriteError(w, r, http.StatusUnprocessableEntity, "ERR_LAT_OUT_OF_BOUNDS", "Unprocessable Entity",
			"Latitude must be within the South Polar region (-90.0° to -75.0°).")
		return
	}
	if req.Longitude < -180.0 || req.Longitude > 180.0 {
		WriteError(w, r, http.StatusUnprocessableEntity, "ERR_LON_OUT_OF_BOUNDS", "Unprocessable Entity",
			"Longitude must be within [-180.0°, +180.0°].")
		return
	}
	if req.MastHeightM <= 0 {
		req.MastHeightM = 2.5
	}

	profile, err := h.raycaster.ComputeHorizon(req.Latitude, req.Longitude, req.MastHeightM)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "ERR_RAYCAST_FAILED", "Calculation Error", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, profile)
}

// HandleEvaluate computes full mission timeline, illumination, DTE comms, and feasibility score.
func (h *Handler) HandleEvaluate(w http.ResponseWriter, r *http.Request) {
	var req evaluation.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, http.StatusBadRequest, "ERR_INVALID_JSON", "Bad Request", "Malformed JSON body.")
		return
	}

	if req.Latitude < -90.0 || req.Latitude > -75.0 {
		WriteError(w, r, http.StatusUnprocessableEntity, "ERR_LAT_OUT_OF_BOUNDS", "Unprocessable Entity",
			"Latitude must be within South Polar region (-90.0° to -75.0°).")
		return
	}

	result, err := h.evalService.Evaluate(req)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "ERR_EVALUATION_FAILED", "Evaluation Error", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

// HandleRecommend searches and ranks Top-N landing sites matching criteria.
func (h *Handler) HandleRecommend(w http.ResponseWriter, r *http.Request) {
	var req recommender.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, http.StatusBadRequest, "ERR_INVALID_JSON", "Bad Request", "Malformed JSON body.")
		return
	}

	res, err := h.recService.Recommend(r.Context(), req)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "ERR_RECOMMEND_FAILED", "Recommender Error", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, res)
}

// SaveScenarioResponse defines the response after saving a scenario.
type SaveScenarioResponse struct {
	ID       string `json:"id"`
	ShareURL string `json:"share_url"`
}

// HandleSaveScenario saves a custom landing scenario to the SQLite database.
func (h *Handler) HandleSaveScenario(w http.ResponseWriter, r *http.Request) {
	var s domain.SavedScenario
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		WriteError(w, r, http.StatusBadRequest, "ERR_INVALID_JSON", "Bad Request", "Malformed JSON body.")
		return
	}

	if s.Title == "" {
		s.Title = "Custom Artemis Mission Evaluation"
	}
	if s.ID == "" {
		// Generate random 8-character hex ID
		b := make([]byte, 4)
		_, _ = rand.Read(b)
		s.ID = "scn_" + hex.EncodeToString(b)
	}

	if err := h.db.SaveScenario(r.Context(), &s); err != nil {
		WriteError(w, r, http.StatusInternalServerError, "ERR_SAVE_SCENARIO", "Database Error", err.Error())
		return
	}

	resp := SaveScenarioResponse{
		ID:       s.ID,
		ShareURL: fmt.Sprintf("/scenario/%s", s.ID),
	}
	WriteJSON(w, http.StatusCreated, resp)
}

// HandleGetScenario retrieves a saved scenario by ID.
func (h *Handler) HandleGetScenario(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, r, http.StatusBadRequest, "ERR_INVALID_ID", "Bad Request", "Missing scenario ID.")
		return
	}

	scenario, err := h.db.GetScenario(r.Context(), id)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "ERR_DB_QUERY", "Database Error", err.Error())
		return
	}
	if scenario == nil {
		WriteError(w, r, http.StatusNotFound, "ERR_SCENARIO_NOT_FOUND", "Not Found", fmt.Sprintf("Scenario '%s' not found.", id))
		return
	}

	WriteJSON(w, http.StatusOK, scenario)
}
