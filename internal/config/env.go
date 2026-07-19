package config

import (
	"os"
	"path/filepath"
)

var (
	habitConfigDir = "HABIT_CONFIG_DIR"
	credentialsEnv = "HABIT_CREDENTIALS_DIR"
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

func CredentialsPath() (string, error) {
	if envPath := os.Getenv(credentialsEnv); envPath != "" {
		return filepath.Join(envPath, "credentials.json"), nil
	}

	return filepath.Join(ConfigDir(), "credentials.json"), nil
}
