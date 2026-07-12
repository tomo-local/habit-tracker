package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config は ~/.config/habit-tracker/config.json の内容。
// Calendars には習慣として扱うカレンダー名(summary)を列挙する。
// GroupByTitle が true の場合、カレンダー内のイベントタイトルごとに別習慣として扱う。
// Habits に正式な習慣名を登録すると、タイトルは正規化+部分一致でこの名前に集約され、
// add コマンドもこの中の名前しか受け付けなくなる(誤字対策)。
type Config struct {
	Calendars    []string `json:"calendars"`
	GroupByTitle bool     `json:"group_by_title"`
	Habits       []string `json:"habits"`
}

func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot resolve home dir:", err)
		os.Exit(1)
	}
	return filepath.Join(home, ".config", "habit-tracker")
}

func LoadConfig() (*Config, error) {
	b, err := os.ReadFile(filepath.Join(ConfigDir(), "config.json"))
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
