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

	"habit-tracker/internal/calendar"
	"habit-tracker/internal/prompt"
)

const (
	defaultDuration = 30
	optionNewHabit  = "New habit"
	minDuration     = 1
	maxDuration     = 24 * 60
)

func (c *Cmd) RunAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	duration := fs.Int("d", defaultDuration, "duration in minutes")
	description := fs.String("D", "", "event description (Markdown)")
	fs.Usage = func() {
		showAddHelp()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return fmt.Errorf("add takes at most one habit, got %d", fs.NArg())
	}
	if durationErr := validateDuration(*duration); durationErr != nil {
		return durationErr
	}

	if err := c.cfg.Read(); err != nil {
		return err
	}

	var habit string
	var err error
	if fs.NArg() == 1 {
		habit = fs.Arg(0)
	} else {
		habit, err = c.selectOrCreateHabit(c.cfg.Habits)
		if err != nil {
			return err
		}
	}
	durationSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "d" {
			durationSet = true
		}
	})
	if !durationSet {
		*duration, err = c.selectDuration()
		if err != nil {
			return err
		}
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
	if err := client.AddEvent(c.cfg.CalendarID, habit, *description, time.Duration(*duration)*time.Minute); err != nil {
		return fmt.Errorf("add event %q: %w", habit, err)
	}
	fmt.Printf("Added: %s %dm\n", habit, *duration)
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
	if err := validateDuration(n); err != nil {
		return 0, err
	}
	return n, nil
}

func validateDuration(n int) error {
	if n < minDuration || n > maxDuration {
		return fmt.Errorf("duration must be between %d and %d minutes, got %d", minDuration, maxDuration, n)
	}
	return nil
}

func (c *Cmd) selectOrCreateHabit(habits []string) (string, error) {
	if len(habits) == 0 {
		return "", fmt.Errorf("no habits configured (run setup first)")
	}

	options := make([]prompt.Option, 0, len(habits)+1)
	for _, h := range habits {
		options = append(options, prompt.Option{Label: h, Value: h})
	}
	options = append(options, prompt.Option{Label: optionNewHabit, Value: optionNewHabit})
	selected, err := c.prompt.Select("Select a habit:", options, "")
	if err != nil {
		return "", err
	}
	if selected.Value != optionNewHabit {
		return selected.Value, nil
	}

	name, err := c.prompt.Input("Habit name:", "")
	if err != nil {
		return "", err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("habit name is empty")
	}
	return name, nil
}

func showAddHelp() {
	fmt.Fprint(os.Stderr, `Usage: habit-tracker add [habit] [options]

Record today's habit on the configured Google Calendar.

Arguments:
  habit  Habit name to record. If omitted, you'll be prompted to
         select one interactively1 (or type a new one).

Options:
  -d <minutes>      Duration to record (default: 30). If omitted, you'll be
                    prompted to enter a value interactively.
  -D <description>  Event description in Markdown (default: none). Not
                    prompted for interactively; only settable via this flag.

Examples:
  habit-tracker add                  # select a habit and duration interactively
  habit-tracker add Golang           # record "Golang" for 30 minutes
  habit-tracker add -d 60 Golang     # record "Golang" for 60 minutes
  habit-tracker add -D "**done**" Golang  # record "Golang" with a Markdown description
`)
}
