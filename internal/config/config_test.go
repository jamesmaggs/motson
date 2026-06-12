package config_test

import (
	"strings"
	"testing"

	"github.com/jamesmaggs/motson/internal/config"
)

func valid() map[string]string {
	return map[string]string{
		"DATABASE_URL":        "postgres://localhost/motson",
		"FOOTBALL_DATA_TOKEN": "secret",
	}
}

func TestLoadsWithDefaults(t *testing.T) {
	cfg, err := config.Load(lookup(valid()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseURL != "postgres://localhost/motson" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.FootballDataToken != "secret" {
		t.Errorf("FootballDataToken = %q", cfg.FootballDataToken)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want default 8080", cfg.Port)
	}
	if cfg.FeedHost != "motson.jamesmaggs.com" {
		t.Errorf("FeedHost = %q, want default motson.jamesmaggs.com (ADR 0011)", cfg.FeedHost)
	}
	if cfg.FootballDataURL != "https://api.football-data.org" {
		t.Errorf("FootballDataURL = %q", cfg.FootballDataURL)
	}
	if cfg.Competition != "WC" {
		t.Errorf("Competition = %q, want default WC", cfg.Competition)
	}
}

func TestOverridesFromEnvironment(t *testing.T) {
	env := valid()
	env["PORT"] = "9999"
	env["FEED_HOST"] = "example.com"

	cfg, err := config.Load(lookup(env))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "9999" || cfg.FeedHost != "example.com" {
		t.Errorf("overrides not applied: %+v", cfg)
	}
}

func TestMissingRequiredValuesAreReportedTogether(t *testing.T) {
	_, err := config.Load(lookup(map[string]string{}))
	if err == nil {
		t.Fatal("want error for missing required config")
	}
	for _, name := range []string{"DATABASE_URL", "FOOTBALL_DATA_TOKEN"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not name missing %s", err, name)
		}
	}
}

func lookup(env map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	}
}
