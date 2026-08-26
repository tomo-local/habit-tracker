package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"habit-tracker/internal/calendar"
	"habit-tracker/internal/config"
	"habit-tracker/internal/prompt"
)

const (
	defaultNewCalendarName = "habit"
	defaultViewWeek        = 12
	optionCreateCalendar   = "Create a new calendar"
	optionSelectCalendar   = "Select an existing calendar"
)

func (c *Cmd) RunSetup(args []string) error {
	oauthCfg := &oauth2.Config{
		ClientID:     c.cfg.ClientID(),
		ClientSecret: c.cfg.ClientSecret(),
		Endpoint:     google.Endpoint,
		Scopes:       calendar.Scopes,
	}
	token, err := loadOrAuthorizeToken(c.cfg, oauthCfg)
	if err != nil {
		return err
	}
	cal, err := c.cal.GetClient(c.ctx, oauthCfg, token)
	if err != nil {
		return err
	}

	fmt.Println("Enter habit names (empty line to finish):")
	scanner := bufio.NewScanner(os.Stdin)
	defer scanner.Err() // ignore error on exit

	var habits []string
	for {
		fmt.Print("> ")
		scanner.Scan()
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			break
		}
		habits = append(habits, line)
	}
	if len(habits) == 0 {
		return fmt.Errorf("no habits entered")
	}

	selected, err := c.selectOrCreateCalendar(cal)
	if err != nil {
		return err
	}

	cfg := &config.Config{
		CalendarID:   selected.ID,
		CalendarName: selected.Name,
		Habits:       habits,
	}
	if err := cfg.Write(); err != nil {
		return fmt.Errorf("write config (calendar %q was already created; delete it manually if unwanted): %w", selected.Name, err)
	}
	fmt.Printf("Saved: calendar=%q habits=%v\n", selected.Name, habits)
	return nil
}

func (c *Cmd) selectOrCreateCalendar(cal calendar.Client) (*calendar.CalendarEntry, error) {
	choice, err := c.prompt.Select("Calendar:", []prompt.Option{
		{Label: optionCreateCalendar, Value: optionCreateCalendar},
		{Label: optionSelectCalendar, Value: optionSelectCalendar},
	}, optionCreateCalendar)
	if err != nil {
		return nil, err
	}

	if choice.Value == optionCreateCalendar {
		name, err := c.prompt.Input("Calendar name:", defaultNewCalendarName)
		if err != nil {
			return nil, err
		}
		entry, err := cal.CreateCalendar(name)
		if err != nil {
			return nil, fmt.Errorf("create calendar: %w", err)
		}
		return entry, nil
	}

	entries, err := cal.ListCalendars()
	if err != nil {
		return nil, fmt.Errorf("list calendars: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no calendars found")
	}

	options := make([]prompt.Option, len(entries))
	for i, e := range entries {
		options[i] = prompt.Option{Label: e.Name, Value: e.ID}
	}
	selected, err := c.prompt.Select("Select a calendar:", options, "")
	if err != nil {
		return nil, err
	}
	return &calendar.CalendarEntry{ID: selected.Value, Name: selected.Label}, nil
}
