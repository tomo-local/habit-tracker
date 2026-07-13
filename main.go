package main

import "fmt"

func main() {
	fmt.Println("Hello World!")

	config := Config{
		googleCalendarTokenPath: ".",
		habitTrackerConfigPath:  ".",
	}

	config.GetGoogleCalendarToken()
	config.GetHabitTrackerConfig()
}

type Config struct {
	googleCalendarTokenPath string
	habitTrackerConfigPath  string
}

type HabitTrackerConfig struct {
	Calendars    []string `json:"calendars"`
	GroupByTitle bool     `json:"group_by_title"`
	Habits       []string `json:"habits"`
}

func (c *Config) GetGoogleCalendarToken() ([]byte, error) {
	fmt.Println("GetGoogleCalendarToken")
	return nil, nil
}

func (c *Config) GetHabitTrackerConfig() (*HabitTrackerConfig, error) {
	fmt.Println("GetHabitTrackerConfig")
	return nil, nil
}
