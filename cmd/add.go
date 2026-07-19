package cmd

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"habit-tracker/internal/config"
)

func RunAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	duration := fs.Int("d", 30, "duration in minutes")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx := context.Background()
	cal, err := newCalendarClient(ctx)
	if err != nil {
		return err
	}

	cfg, err := config.New()
	if err != nil {
		return err
	}

	var habits []string
	if fs.NArg() > 0 {
		habits = []string{fs.Arg(0)}
	} else {
		habits, err = selectHabits(cfg.Habits)
		if err != nil {
			return err
		}
	}

	for _, h := range habits {
		if err := cal.AddEvent(cfg.CalendarID, h, time.Duration(*duration)*time.Minute); err != nil {
			return fmt.Errorf("add event %q: %w", h, err)
		}
		fmt.Printf("Added: %s\n", h)
	}
	return nil
}

func selectHabits(habits []string) ([]string, error) {
	if len(habits) == 0 {
		return nil, fmt.Errorf("no habits configured (run setup first)")
	}
	for i, h := range habits {
		fmt.Printf("  %d) %s\n", i+1, h)
	}
	fmt.Print("Select (comma-separated): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	parts := strings.Split(scanner.Text(), ",")

	var selected []string
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 1 || n > len(habits) {
			return nil, fmt.Errorf("invalid selection: %q", p)
		}
		selected = append(selected, habits[n-1])
	}
	return selected, nil
}
