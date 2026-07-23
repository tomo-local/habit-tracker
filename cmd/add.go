package cmd

import (
	"bufio"
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
	duration := fs.Int("d", 30, "duration in minutes")
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
		habits, err = selectHabits(c.cfg.Habits)
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
