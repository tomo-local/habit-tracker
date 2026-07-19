package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	mu           sync.RWMutex
	CalendarID   string   `json:"calendar_id"`
	CalendarName string   `json:"calendar_name,omitempty"`
	Habits       []string `json:"habits"`
}

func New() (*Config, error) {
	c := &Config{}
	if err := c.Read(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) Read() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.ReadFile(filepath.Join(ConfigDir(), "config.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(file, c)
}

func (c *Config) Write() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := os.MkdirAll(ConfigDir(), 0700); err != nil {
		return err
	}
	tmpFile := filepath.Join(ConfigDir(), "config.json.tmp")
	file, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmpFile, file, 0600); err != nil {
		return err
	}
	return os.Rename(tmpFile, filepath.Join(ConfigDir(), "config.json"))
}
