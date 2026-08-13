package cmd

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gcal "google.golang.org/api/calendar/v3"
)

func (c *Cmd) RunAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	duration := fs.Int("d", defaultDuration, "duration in minutes")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: habit-tracker add [habit...] [options]

Record today's habit on the configured Google Calendar.

Arguments:
  habit...  Habit name(s) to record. If omitted, you'll be prompted to
            select one interactively (or type a new one).

Options:
  -d <minutes>  Duration to record (default: 30). If omitted, you'll be
                prompted to enter a value interactively.

Examples:
  habit-tracker add                  # select a habit and duration interactively
  habit-tracker add Golang           # record "Golang" for 30 minutes
  habit-tracker add Golang -d 60     # record "Golang" for 60 minutes
`)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	if err := c.cfg.Read(); err != nil {
		return err
	}

	var habits []string
	var err error
	if fs.NArg() > 0 {
		habits = fs.Args()
	} else {
		habits, err = c.selectHabits(c.cfg.Habits)
		if err != nil {
			return err
		}
	}

	dSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "d" {
			dSet = true
		}
	})
	if !dSet {
		*duration, err = c.selectDuration()
		if err != nil {
			return err
		}
	}

	token, err := c.cfg.LoadToken()
	if err != nil {
		return fmt.Errorf("read token (run auth login first): %w", err)
	}
	oauthCfg := &oauth2.Config{
		ClientID:     c.cfg.ClientID(),
		ClientSecret: c.cfg.ClientSecret(),
		Endpoint:     google.Endpoint,
		Scopes:       []string{gcal.CalendarScope},
	}
	client, err := c.cal.GetClient(c.ctx, oauthCfg, token)
	if err != nil {
		return err
	}

	for _, h := range habits {
		if err := client.AddEvent(c.cfg.CalendarID, h, time.Duration(*duration)*time.Minute); err != nil {
			return fmt.Errorf("add event %q: %w", h, err)
		}
		fmt.Printf("Added: %s %dm\n", h, *duration)
	}
	return nil
}

func (c *Cmd) selectDuration() (int, error) {
	text, err := c.prompt.Input("Duration (minutes):", strconv.Itoa(defaultDuration))
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %q", text)
	}
	return n, nil
}

func (c *Cmd) selectHabits(habits []string) ([]string, error) {
	if len(habits) == 0 {
		return nil, fmt.Errorf("no habits configured (run setup first)")
	}

	options := append(append([]string{}, habits...), optionNewHabit)
	selected, err := c.prompt.Select("Select a habit:", options, "")
	if err != nil {
		return nil, err
	}
	if selected != optionNewHabit {
		return []string{selected}, nil
	}

	name, err := c.prompt.Input("Habit name:", "")
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("habit name is empty")
	}
	return []string{name}, nil
}
