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
	return os.Getenv("HABIT_CLIENT_ID")
}

func (c *Config) ClientSecret() string {
	return os.Getenv("HABIT_CLIENT_SECRET")
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
