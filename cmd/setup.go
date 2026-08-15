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
)

const (
	defaultNewCalendarName = "habit"
	optionCreateCalendar   = "Create a new calendar"
	optionSelectCalendar   = "Select an existing calendar"
)

func (c *Cmd) RunSetup(args []string) error {
	token, err := c.cfg.LoadToken()
	if err != nil {
		return fmt.Errorf("read token (run auth login first): %w", err)
	}
	oauthCfg := &oauth2.Config{
		ClientID:     c.cfg.ClientID(),
		ClientSecret: c.cfg.ClientSecret(),
		Endpoint:     google.Endpoint,
		Scopes:       calendar.Scopes,
	}
	cal, err := c.cal.GetClient(c.ctx, oauthCfg, token)
	if err != nil {
		return err
	}

	selected, err := c.selectOrCreateCalendar(cal)
	if err != nil {
		return err
	}

	fmt.Println("Enter habit names (empty line to finish):")
	scanner := bufio.NewScanner(os.Stdin)
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

	cfg := &config.Config{
		CalendarID:   selected.ID,
		CalendarName: selected.Name,
		Habits:       habits,
	}
	if err := cfg.Write(); err != nil {
		return err
	}
	fmt.Printf("Saved: calendar=%q habits=%v\n", selected.Name, habits)
	return nil
}

func (c *Cmd) selectOrCreateCalendar(cal calendar.Client) (*calendar.CalendarEntry, error) {
	choice, err := c.prompt.Select("Calendar:", []string{optionCreateCalendar, optionSelectCalendar}, optionCreateCalendar)
	if err != nil {
		return nil, err
	}

	if choice == optionCreateCalendar {
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

	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}
	selectedName, err := c.prompt.Select("Select a calendar:", names, "")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.Name == selectedName {
			return e, nil
		}
	}
	return nil, fmt.Errorf("calendar not found: %q", selectedName)
}
