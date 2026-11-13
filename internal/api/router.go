package api

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// NewRouter registers all API endpoints and applies the middleware stack.
func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	// 1. Health & Telemetry
	mux.HandleFunc("GET /api/v1/health", h.HandleHealth)

	// 2. Official Artemis Presets
	mux.HandleFunc("GET /api/v1/presets", h.HandleListPresets)
	mux.HandleFunc("GET /api/v1/presets/{id}", h.HandleGetPreset)

	// 3. Local Horizon Elevation Raymarching
	mux.HandleFunc("POST /api/v1/horizon", h.HandleHorizon)

	// 4. Mission Environmental Evaluation
	mux.HandleFunc("POST /api/v1/evaluate", h.HandleEvaluate)

	// 5. Smart Site Recommender Engine (Top-5)
	mux.HandleFunc("POST /api/v1/recommend", h.HandleRecommend)

	// 6. Saved Mission Scenarios & Bookmarks
	mux.HandleFunc("POST /api/v1/scenarios", h.HandleSaveScenario)
	mux.HandleFunc("GET /api/v1/scenarios/{id}", h.HandleGetScenario)

	// 7. Serve Static Frontend Web App (Production React dist or fallback)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		distCandidates := []string{"web/dist", "../../web/dist", "dist"}
		var distDir string
		for _, cand := range distCandidates {
			if info, err := os.Stat(cand); err == nil && info.IsDir() {
				distDir = cand
				break
			}
		}

		if distDir != "" {
			fs := http.FileServer(http.Dir(distDir))
			cleanPath := path.Clean(r.URL.Path)
			targetFile := filepath.Join(distDir, cleanPath)
			if stat, err := os.Stat(targetFile); err == nil && !stat.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
			http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
			return
		}

		// Fallback for root index.html if web/dist not built
		rootCandidates := []string{"index.html", "../../index.html"}
		for _, cand := range rootCandidates {
			if _, err := os.Stat(cand); err == nil {
				if r.URL.Path != "/" && r.URL.Path != "/index.html" {
					http.NotFound(w, r)
					return
				}
				http.ServeFile(w, r, cand)
				return
			}
		}
		http.NotFound(w, r)
	})

	// Apply middleware stack
	middlewareStack := CreateMiddlewareStack(
		RecoveryMiddleware,
		CORSMiddleware,
		LoggingMiddleware,
	)

	return middlewareStack(mux)
}
