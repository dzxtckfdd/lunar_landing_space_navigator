package main

import (
	"context"
	"fmt"
	"log"
	"lunar-navigator/internal/api"
	"lunar-navigator/internal/astro"
	"lunar-navigator/internal/config"
	"lunar-navigator/internal/database"
	"lunar-navigator/internal/dem"
	"lunar-navigator/internal/evaluation"
	"lunar-navigator/internal/mock"
	"lunar-navigator/internal/recommender"
	"net/http"
	"time"
)

func main() {
	cfg := config.Load()
	log.Printf("[LUNAR-NAVIGATOR] Initializing Lunar Landing Site Navigator (Environment: %s)...", cfg.Environment)

	// 1. Initialize SQLite Database
	db, err := database.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("[FATAL] Failed to initialize database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Apply Migrations
	log.Println("[DB] Applying database schema migrations...")
	if err := db.Migrate(ctx); err != nil {
		log.Fatalf("[FATAL] Database migration failed: %v", err)
	}

	// Seed Mock Data if enabled
	if cfg.MockMode {
		log.Println("[SEED] Seeding official Artemis III candidate regions and candidate grid...")
		if err := mock.SeedDatabase(ctx, db); err != nil {
			log.Fatalf("[FATAL] Mock database seeding failed: %v", err)
		}
		count, _ := db.CountCandidateSites(ctx)
		presets, _ := db.ListPresets(ctx)
		log.Printf("[SEED] Seeding complete: %d presets and %d candidate landing spots available.", len(presets), count)
	}

	// 2. Initialize DEM Matrix & Raycaster
	log.Println("[DEM] Initializing South Pole elevation grid (50m resolution, 260km extent)...")
	demMatrix := dem.NewSyntheticMatrix(2600, 2600, 100.0)
	raycaster := dem.NewRaycaster(demMatrix)

	// 3. Initialize Astrodynamics & Core Services
	log.Println("[ASTRO] Initializing ephemeris service (JPL Horizons / SPICE models)...")
	ephemService := astro.NewEphemerisService()
	evalService := evaluation.NewService(demMatrix, raycaster, ephemService)
	recService := recommender.NewService(db, evalService)

	// 4. Initialize HTTP Handlers & Router
	handler := api.NewHandler(cfg, db, demMatrix, raycaster, evalService, recService)
	router := api.NewRouter(handler)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("🚀 Lunar Landing Site Navigator Monolith ready at http://localhost%s", addr)
	log.Printf("📡 REST API endpoints active under /api/v1/*")

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[FATAL] Server failed: %v", err)
	}
}
