package config

import (
	"flag"
	"os"
	"strconv"
)

// Config holds the application configuration parameters.
type Config struct {
	Port         int
	DBPath       string
	DEMPath      string
	MockMode     bool
	Environment  string
	DefaultMastH float64
	FreezeHours  int
}

// Load loads configuration from command-line flags and environment variables.
func Load() *Config {
	cfg := &Config{
		Port:         8080,
		DBPath:       "data/lunar_navigator.db",
		DEMPath:      "data/lola_southpole_50m.bin",
		MockMode:     true, // Default to true for ease of development & demo
		Environment:  "development",
		DefaultMastH: 2.5,
		FreezeHours:  36,
	}

	// Environment variable overrides
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = p
		}
	}
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		cfg.DBPath = dbPath
	}
	if demPath := os.Getenv("DEM_PATH"); demPath != "" {
		cfg.DEMPath = demPath
	}
	if mockStr := os.Getenv("MOCK_MODE"); mockStr != "" {
		cfg.MockMode = mockStr == "true" || mockStr == "1"
	}
	if env := os.Getenv("APP_ENV"); env != "" {
		cfg.Environment = env
	}

	// Flag overrides (only parse if not already parsed, e.g. in tests)
	if !flag.Parsed() {
		flag.IntVar(&cfg.Port, "port", cfg.Port, "HTTP server port")
		flag.StringVar(&cfg.DBPath, "db", cfg.DBPath, "Path to SQLite database file")
		flag.StringVar(&cfg.DEMPath, "dem", cfg.DEMPath, "Path to LOLA DEM binary grid")
		flag.BoolVar(&cfg.MockMode, "mock", cfg.MockMode, "Run in mock/seed data mode")
		flag.StringVar(&cfg.Environment, "env", cfg.Environment, "Runtime environment (development/production)")
		flag.Float64Var(&cfg.DefaultMastH, "mast", cfg.DefaultMastH, "Default instrument mast height in meters")
		flag.IntVar(&cfg.FreezeHours, "freeze-hours", cfg.FreezeHours, "Default critical shadow freeze threshold in hours")
		flag.Parse()
	}

	return cfg
}
