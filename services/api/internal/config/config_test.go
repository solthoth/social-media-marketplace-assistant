package config

import "testing"

func TestLoadUsesDefaults(t *testing.T) {
	cfg := Load(func(string) string {
		return ""
	})

	if cfg.Port != "8080" {
		t.Fatalf("expected default port, got %q", cfg.Port)
	}
	if cfg.DatabasePath != "data/app.db" {
		t.Fatalf("expected default database path, got %q", cfg.DatabasePath)
	}
}

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	values := map[string]string{
		"PORT":          "9090",
		"DATABASE_PATH": "/tmp/marketplace.db",
	}

	cfg := Load(func(key string) string {
		return values[key]
	})

	if cfg.Port != "9090" {
		t.Fatalf("expected overridden port, got %q", cfg.Port)
	}
	if cfg.DatabasePath != "/tmp/marketplace.db" {
		t.Fatalf("expected overridden database path, got %q", cfg.DatabasePath)
	}
}
