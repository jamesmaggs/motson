// Package config loads Motson's runtime configuration from the
// environment: secrets and platform values only — domain defaults
// (sync cadence, durations, staleness) live in the fixtures package
// as the spec's config block.
package config

import (
	"fmt"
	"strings"
)

type Config struct {
	DatabaseURL       string
	FootballDataToken string
	FootballDataURL   string
	Competition       string
	Port              string
	FeedHost          string
}

// Load reads configuration via lookup (os.LookupEnv in production).
// Missing required values are reported together.
func Load(lookup func(string) (string, bool)) (Config, error) {
	get := func(key, fallback string) string {
		if v, ok := lookup(key); ok && v != "" {
			return v
		}
		return fallback
	}

	cfg := Config{
		DatabaseURL:       get("DATABASE_URL", ""),
		FootballDataToken: get("FOOTBALL_DATA_TOKEN", ""),
		FootballDataURL:   get("FOOTBALL_DATA_URL", "https://api.football-data.org"),
		Competition:       get("COMPETITION", "WC"),
		Port:              get("PORT", "8080"),
		FeedHost:          get("FEED_HOST", "motson.jamesmaggs.com"),
	}

	var missing []string
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.FootballDataToken == "" {
		missing = append(missing, "FOOTBALL_DATA_TOKEN")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}
