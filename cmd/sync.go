package cmd

import (
	"flag"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gcal "google.golang.org/api/calendar/v3"

	"habit-tracker/internal/sessionlog"
)

func (c *Cmd) RunSync(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	tag := fs.String("tag", "", "habit name to record the sessions under")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: habit-tracker sync [options]

Read Claude Code session logs for the current directory
(~/.claude/projects/<slug>) and back-fill calendar events for each block of
work, split whenever a session goes quiet for more than 15 minutes. Each
event's description lists the prompts you sent during that block.

Options:
  -tag <habit>  Habit name to record the sessions under. If omitted, you'll
                be prompted to select one interactively (or type a new one).

Examples:
  habit-tracker sync                    # pick a habit interactively and sync
  habit-tracker sync -tag habit-tracker # sync under the "habit-tracker" habit
`)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	if err := c.cfg.Read(); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	dir, err := sessionlog.ProjectDir(cwd)
	if err != nil {
		return err
	}

	blocks, latest, err := sessionlog.Blocks(dir, c.cfg.LastSyncedAt)
	if err != nil {
		return fmt.Errorf("read session logs: %w", err)
	}
	if len(blocks) == 0 {
		fmt.Println("Nothing to sync.")
		return nil
	}

	habitTag := *tag
	if habitTag == "" {
		habitTag, err = c.selectOrNewHabit(c.cfg.Habits)
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

	for _, b := range blocks {
		if err := client.AddEventAt(c.cfg.CalendarID, habitTag, b.Description(), b.Start, b.End); err != nil {
			return fmt.Errorf("add event %s: %w", b.Start.Format(time.RFC3339), err)
		}
		fmt.Printf("Added: %s  %s - %s (%d notes)\n", habitTag, b.Start.Format("2006-01-02 15:04"), b.End.Format("15:04"), len(b.Notes))
	}

	c.cfg.LastSyncedAt = latest
	if err := c.cfg.Write(); err != nil {
		return err
	}
	fmt.Printf("Synced %d event(s).\n", len(blocks))
	return nil
}
