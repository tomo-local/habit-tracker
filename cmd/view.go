package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"habit-tracker/internal/heatmap"
)

func (c *Cmd) RunView(args []string, isDefault bool) error {
	fs := flag.NewFlagSet("view", flag.ContinueOnError)
	weeks := fs.Int("w", 0, "number of weeks to display (default: config value or 52)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if err := c.cfg.Read(); err != nil {
		return err
	}

	if len(c.cfg.Habits) == 0 {
		return fmt.Errorf("no habits configured (run setup first)")
	}

	if *weeks <= 0 {
		*weeks = c.cfg.Weeks()
	}

	var habitName string
	if isDefault {
		habitName = c.cfg.Habits[0]
	} else {
		habit, err := selectHabit(c.cfg.Habits)
		if err != nil {
			return err
		}
		habitName = habit
	}

	client, err := c.getClient()
	if err != nil {
		return err
	}

	now := time.Now()
	events, err := client.GetEvents(c.cfg.CalendarID, now.AddDate(0, 0, -*weeks*7), now.AddDate(0, 0, 1))
	if err != nil {
		return fmt.Errorf("get events: %w", err)
	}

	counts := make(map[string]int)
	for _, e := range events {
		if e.Title != habitName || e.End == nil {
			continue
		}
		minutes := int(e.End.Sub(e.Start).Minutes())
		counts[e.Start.Format("2006-01-02")] += minutes
	}

	streak := calcStreak(counts, now)
	fmt.Printf("%s  🔥 %d day streak\n\n", habitName, streak)
	heatmap.Render(counts, now, *weeks)
	return nil
}

func selectHabit(habits []string) (string, error) {
	if len(habits) == 0 {
		return "", fmt.Errorf("no habits configured (run setup first)")
	}
	for i, h := range habits {
		fmt.Printf("  %d) %s\n", i+1, h)
	}
	fmt.Print("Select (comma-separated): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	part := scanner.Text()

	n, err := strconv.Atoi(strings.TrimSpace(part))
	if err != nil || n < 1 || n > len(habits) {
		return "", fmt.Errorf("invalid selection: %q", n)
	}

	return habits[n-1], nil
}

func calcStreak(counts map[string]int, now time.Time) int {
	d := now
	if counts[d.Format("2006-01-02")] == 0 {
		d = d.AddDate(0, 0, -1)
	}
	streak := 0
	for counts[d.Format("2006-01-02")] > 0 {
		streak++
		d = d.AddDate(0, 0, -1)
	}
	return streak
}
