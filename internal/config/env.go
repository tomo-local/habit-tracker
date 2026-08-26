package config

// env.go holds settings sourced from environment variables
// (HABIT_CONFIG_DIR, HABIT_CLIENT_ID, HABIT_CLIENT_SECRET) rather than config.json.

import (
	"encoding/json"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

var habitConfigDir = "HABIT_CONFIG_DIR"

// defaultOAuth holds the shared OAuth client used when HABIT_CLIENT_ID /
// HABIT_CLIENT_SECRET are not set. Values are empty in source control and
// injected at release-build time via:
//
//	go build -ldflags "-X habit-tracker/internal/config.defaultClientID=... -X habit-tracker/internal/config.defaultClientSecret=..."
var (
	defaultClientID     = ""
	defaultClientSecret = ""
)

func ConfigDir() string {
	if envPath := os.Getenv(habitConfigDir); envPath != "" {
		return envPath
	}

	d, _ := os.UserHomeDir()
	return filepath.Join(d, ".config", "habit")
}

func TokenPath() string {
	return filepath.Join(ConfigDir(), "token.json")
}

func (c *Config) ClientID() string {
	if v := os.Getenv("HABIT_CLIENT_ID"); v != "" {
		return v
	}
	return defaultClientID
}

func (c *Config) ClientSecret() string {
	if v := os.Getenv("HABIT_CLIENT_SECRET"); v != "" {
		return v
	}
	return defaultClientSecret
}

func (c *Config) LoadToken() (*oauth2.Token, error) {
	b, err := os.ReadFile(TokenPath())
	if err != nil {
		return nil, err
	}
	var token oauth2.Token
	if err := json.Unmarshal(b, &token); err != nil {
		return nil, err
	}
	return &token, nil
}
