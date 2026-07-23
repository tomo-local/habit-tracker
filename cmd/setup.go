package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"habit-tracker/internal/config"
)

func (c *Cmd) RunSetup(args []string) error {
	cal, err := c.getClient()
	if err != nil {
		return err
	}

	entries, err := cal.ListCalendars()
	if err != nil {
		return fmt.Errorf("list calendars: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("no calendars found")
	}

	fmt.Println("Select a calendar:")
	for i, e := range entries {
		fmt.Printf("  %d) %s\n", i+1, e.Name)
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Number: ")
	scanner.Scan()
	n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || n < 1 || n > len(entries) {
		return fmt.Errorf("invalid selection")
	}
	selected := entries[n-1]

	fmt.Println("Enter habit names (empty line to finish):")
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
