package config

import (
	"os"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	cfg := Load()

	if cfg.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Port)
	}
	if cfg.DefaultMastH != 2.5 {
		t.Errorf("expected default mast height 2.5m, got %f", cfg.DefaultMastH)
	}
	if cfg.FreezeHours != 36 {
		t.Errorf("expected default freeze hours 36, got %d", cfg.FreezeHours)
	}
	if !cfg.MockMode {
		t.Errorf("expected mock mode to be enabled by default")
	}
}

func TestConfigEnvironmentOverrides(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("DB_PATH", ":memory:")
	os.Setenv("MOCK_MODE", "false")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DB_PATH")
		os.Unsetenv("MOCK_MODE")
	}()

	cfg := Load()

	if cfg.Port != 9090 {
		t.Errorf("expected port override 9090, got %d", cfg.Port)
	}
	if cfg.DBPath != ":memory:" {
		t.Errorf("expected db override ':memory:', got %s", cfg.DBPath)
	}
	if cfg.MockMode {
		t.Errorf("expected mock mode false, got true")
	}
}
