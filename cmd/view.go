package cmd

import (
	"flag"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"habit-tracker/internal/calendar"
	"habit-tracker/internal/heatmap"
	"habit-tracker/internal/prompt"
)

const optionAllHabits = "All habits"

func (c *Cmd) RunView(args []string, isDefault bool) error {
	fs := flag.NewFlagSet("view", flag.ContinueOnError)
	weeks := fs.Int("w", 0, "number of weeks to display (default: config value or 52)")
	fs.Usage = func() {
		showViewHelp()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	if err := c.cfg.Read(); err != nil {
		return err
	}

	if len(c.cfg.Habits) == 0 {
		return fmt.Errorf("no habits configured (run setup first)")
	}
	if fs.NArg() > 1 {
		return fmt.Errorf("view takes at most one habit, got %d", fs.NArg())
	}

	if *weeks <= 0 {
		*weeks = c.cfg.Weeks()
	}

	var habitName string
	switch {
	case fs.NArg() == 1:
		habitName = fs.Arg(0)
		if !contains(c.cfg.Habits, habitName) {
			return fmt.Errorf("habit %q is not configured (run setup first)", habitName)
		}
	case isDefault:
		habitName = optionAllHabits
	default:
		habit, err := c.selectHabit(c.cfg.Habits)
		if err != nil {
			return err
		}
		habitName = habit
	}

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
	client, err := c.cal.GetClient(c.ctx, oauthCfg, token)
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
		if e.End == nil {
			continue
		}
		if habitName != optionAllHabits && e.Title != habitName {
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

func (c *Cmd) selectHabit(habits []string) (string, error) {
	if len(habits) == 0 {
		return "", fmt.Errorf("no habits configured (run setup first)")
	}
	options := make([]prompt.Option, 0, len(habits)+1)
	options = append(options, prompt.Option{Label: optionAllHabits, Value: optionAllHabits})
	for _, h := range habits {
		options = append(options, prompt.Option{Label: h, Value: h})
	}
	selected, err := c.prompt.Select("Select a habit:", options, "")
	if err != nil {
		return "", err
	}
	return selected.Value, nil
}

func contains(habits []string, name string) bool {
	for _, h := range habits {
		if h == name {
			return true
		}
	}
	return false
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

func showViewHelp() {
	fmt.Fprint(os.Stderr, `Usage: habit view [habit] [options]

Select a habit and show its heatmap. Choose "All habits" (or omit the
argument and pick it interactively) to show a combined heatmap and streak
across every tracked habit.

Arguments:
  habit  Habit name to show. If omitted, you'll be prompted to select one
         interactively.

Options:
  -w <weeks>  Number of weeks to display (default: config value, or 52)

Examples:
  habit view          # select a habit (or "All habits") and show its heatmap
  habit view Golang   # show a specific habit directly
  habit view -w 26    # show the last 26 weeks
`)
}
