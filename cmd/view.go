package cmd

import (
	"context"
	"fmt"
	"time"

	"habit-tracker/internal/config"
	"habit-tracker/internal/heatmap"
)

func RunView(args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}
	if len(cfg.Habits) == 0 {
		return fmt.Errorf("no habits configured (run setup first)")
	}

	habitName := cfg.Habits[0]
	if len(args) > 0 {
		habitName = args[0]
	}

	ctx := context.Background()
	cal, err := newCalendarClient(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	events, err := cal.GetEvents(cfg.CalendarID, now.AddDate(0, 0, -52*7), now)
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
	heatmap.Render(counts, now)
	return nil
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
