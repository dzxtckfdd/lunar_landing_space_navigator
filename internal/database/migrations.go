package database

import (
	"context"
	"database/sql"
	"fmt"
	"lunar-navigator/internal/domain"
)

// Migrate applies the schema migrations to the database.
func (db *DB) Migrate(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS presets (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		sub_region TEXT,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		elevation_m REAL NOT NULL,
		average_slope_deg REAL NOT NULL,
		distance_to_psr_km REAL NOT NULL,
		description TEXT,
		scientific_significance TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS candidate_sites (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cluster_region TEXT NOT NULL,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		elevation_m REAL NOT NULL,
		slope_deg REAL NOT NULL,
		distance_to_psr_km REAL NOT NULL,
		is_safe_slope BOOLEAN NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_candidates_filtering 
	ON candidate_sites (slope_deg, distance_to_psr_km, latitude, longitude);

	CREATE TABLE IF NOT EXISTS saved_scenarios (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		antenna_mast_height_m REAL DEFAULT 2.5,
		mission_start_utc DATETIME NOT NULL,
		mission_duration_hours INTEGER NOT NULL,
		launch_slip_days INTEGER DEFAULT 0,
		solar_illumination_pct REAL NOT NULL,
		dte_comm_pct REAL NOT NULL,
		feasibility_score REAL NOT NULL,
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}

// InsertPreset inserts or updates an official Artemis landing preset.
func (db *DB) InsertPreset(ctx context.Context, p *domain.Preset) error {
	query := `
	INSERT INTO presets (
		id, name, sub_region, latitude, longitude, elevation_m, 
		average_slope_deg, distance_to_psr_km, description, scientific_significance
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		sub_region = excluded.sub_region,
		latitude = excluded.latitude,
		longitude = excluded.longitude,
		elevation_m = excluded.elevation_m,
		average_slope_deg = excluded.average_slope_deg,
		distance_to_psr_km = excluded.distance_to_psr_km,
		description = excluded.description,
		scientific_significance = excluded.scientific_significance;
	`
	_, err := db.ExecContext(ctx, query,
		p.ID, p.Name, p.SubRegion, p.Latitude, p.Longitude, p.ElevationM,
		p.AverageSlopeDeg, p.DistanceToPSRkm, p.Description, p.ScientificSignificance,
	)
	return err
}

// ListPresets returns all official landing presets sorted by name.
func (db *DB) ListPresets(ctx context.Context) ([]domain.Preset, error) {
	query := `
	SELECT id, name, sub_region, latitude, longitude, elevation_m, 
	       average_slope_deg, distance_to_psr_km, description, scientific_significance, created_at
	FROM presets
	ORDER BY latitude DESC, name ASC;
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var presets []domain.Preset
	for rows.Next() {
		var p domain.Preset
		if err := rows.Scan(
			&p.ID, &p.Name, &p.SubRegion, &p.Latitude, &p.Longitude, &p.ElevationM,
			&p.AverageSlopeDeg, &p.DistanceToPSRkm, &p.Description, &p.ScientificSignificance, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		presets = append(presets, p)
	}
	return presets, rows.Err()
}

// GetPreset retrieves a single preset by ID.
func (db *DB) GetPreset(ctx context.Context, id string) (*domain.Preset, error) {
	query := `
	SELECT id, name, sub_region, latitude, longitude, elevation_m, 
	       average_slope_deg, distance_to_psr_km, description, scientific_significance, created_at
	FROM presets
	WHERE id = ?;
	`
	var p domain.Preset
	err := db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.SubRegion, &p.Latitude, &p.Longitude, &p.ElevationM,
		&p.AverageSlopeDeg, &p.DistanceToPSRkm, &p.Description, &p.ScientificSignificance, &p.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// InsertCandidateSitesBatch inserts candidate points in a single transaction.
func (db *DB) InsertCandidateSitesBatch(ctx context.Context, sites []domain.CandidateSite) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO candidate_sites (
			cluster_region, latitude, longitude, elevation_m, slope_deg, distance_to_psr_km, is_safe_slope
		) VALUES (?, ?, ?, ?, ?, ?, ?);
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range sites {
		if _, err := stmt.ExecContext(ctx,
			s.ClusterRegion, s.Latitude, s.Longitude, s.ElevationM, s.SlopeDeg, s.DistanceToPSRkm, s.IsSafeSlope,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ListFilteredCandidateSites queries candidate sites matching slope and distance constraints.
func (db *DB) ListFilteredCandidateSites(ctx context.Context, maxSlope float64, maxPSRDist float64, region string) ([]domain.CandidateSite, error) {
	query := `
	SELECT id, cluster_region, latitude, longitude, elevation_m, slope_deg, distance_to_psr_km, is_safe_slope
	FROM candidate_sites
	WHERE slope_deg <= ? AND distance_to_psr_km <= ?
	`
	var args []interface{}
	args = append(args, maxSlope, maxPSRDist)

	if region != "" && region != "all" {
		query += " AND cluster_region = ?"
		args = append(args, region)
	}
	query += " ORDER BY slope_deg ASC, distance_to_psr_km ASC LIMIT 1000;"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []domain.CandidateSite
	for rows.Next() {
		var s domain.CandidateSite
		if err := rows.Scan(
			&s.ID, &s.ClusterRegion, &s.Latitude, &s.Longitude, &s.ElevationM,
			&s.SlopeDeg, &s.DistanceToPSRkm, &s.IsSafeSlope,
		); err != nil {
			return nil, err
		}
		sites = append(sites, s)
	}
	return sites, rows.Err()
}

// CountCandidateSites returns the total number of candidate sites currently stored.
func (db *DB) CountCandidateSites(ctx context.Context) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM candidate_sites;").Scan(&count)
	return count, err
}

// SaveScenario persists a mission scenario bookmark.
func (db *DB) SaveScenario(ctx context.Context, s *domain.SavedScenario) error {
	query := `
	INSERT INTO saved_scenarios (
		id, title, latitude, longitude, antenna_mast_height_m,
		mission_start_utc, mission_duration_hours, launch_slip_days,
		solar_illumination_pct, dte_comm_pct, feasibility_score, notes
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err := db.ExecContext(ctx, query,
		s.ID, s.Title, s.Latitude, s.Longitude, s.AntennaMastHeightM,
		s.MissionStartUTC, s.MissionDurationHours, s.LaunchSlipDays,
		s.SolarIlluminationPct, s.DTECommPct, s.FeasibilityScore, s.Notes,
	)
	return err
}

// GetScenario retrieves a saved scenario by ID.
func (db *DB) GetScenario(ctx context.Context, id string) (*domain.SavedScenario, error) {
	query := `
	SELECT id, title, latitude, longitude, antenna_mast_height_m,
	       mission_start_utc, mission_duration_hours, launch_slip_days,
	       solar_illumination_pct, dte_comm_pct, feasibility_score, notes, created_at
	FROM saved_scenarios
	WHERE id = ?;
	`
	var s domain.SavedScenario
	err := db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.Title, &s.Latitude, &s.Longitude, &s.AntennaMastHeightM,
		&s.MissionStartUTC, &s.MissionDurationHours, &s.LaunchSlipDays,
		&s.SolarIlluminationPct, &s.DTECommPct, &s.FeasibilityScore, &s.Notes, &s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
