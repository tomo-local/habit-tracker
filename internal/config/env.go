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

var defaultOAuth = struct {
	ClientID     string
	ClientSecret string
}{
	ClientID:     "357079410304-cb56g740mmp73o0qa8plopmutd6dbfvk.apps.googleusercontent.com",
	ClientSecret: "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",
}

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
	return defaultOAuth.ClientID
}

func (c *Config) ClientSecret() string {
	if v := os.Getenv("HABIT_CLIENT_SECRET"); v != "" {
		return v
	}
	return defaultOAuth.ClientSecret
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
