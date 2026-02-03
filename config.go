package main

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	URL   string
	Token string
}

func LoadConfig() (Config, error) {
	cfg := Config{
		URL:   os.Getenv("JOT_URL"),
		Token: os.Getenv("JOT_TOKEN"),
	}

	if cfg.URL == "" || cfg.Token == "" {
		return Config{}, fmt.Errorf("missing env vars — set:\n\n  export JOT_URL=\"https://your-db.turso.io\"\n  export JOT_TOKEN=\"your-auth-token\"\n")
	}

	// Turso CLI gives libsql:// URLs, but the HTTP API needs https://
	cfg.URL = strings.Replace(cfg.URL, "libsql://", "https://", 1)

	return cfg, nil
}
