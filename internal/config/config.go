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
	ViewWeek     int      `json:"view_week"`
}

func New() *Config {
	return &Config{}
}

func (c *Config) Weeks() int {
	if c.ViewWeek <= 0 {
		return 52
	}
	return c.ViewWeek
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
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(ConfigDir(), "config.json.*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, filepath.Join(ConfigDir(), "config.json"))
}
