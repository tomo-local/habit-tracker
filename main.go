package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"habit-tracker/cmd"
	"habit-tracker/internal/calendar"
	"habit-tracker/internal/config"
	"habit-tracker/internal/prompt/huh"
)

func main() {
	ctx := context.Background()
	args := os.Args[1:]

	var command string
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	c := cmd.New(ctx, config.New(), calendar.New(), huh.New())

	var err error
	switch command {
	case "":
		err = c.RunView(args, true)
	case "view":
		err = c.RunView(args, false)
	case "auth":
		err = c.RunAuth(args)
	case "setup":
		err = c.RunSetup(args)
	case "add":
		err = c.RunAdd(args)
	case "-h", "--help", "help":
		showHelp()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\nRun 'habit --help' for usage.\n", command)
		os.Exit(1)
	}

	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			// Usage was already printed by the subcommand's FlagSet.
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Fprint(os.Stderr, `habit: track daily habits via Google Calendar events

Usage: habit [command] [options]

Commands:
  (none)              Show combined heatmap for all habits (no interactive selection)
  view [habit]        Select a habit (or show one directly) and show its heatmap
  add [habit...]      Record today's habit
  setup               Create/select a calendar and register habit names
  auth login          Authenticate with your Google account
  help                Show this help message
  -h, --help          Show this help message

Run 'habit <command> -h' for details and options of each command.

Examples:
  habit                 # show today's combined heatmap for all habits
  habit add             # record today's habit interactively
  habit add Golang      # record "Golang" for the default duration
  habit view -w 26      # show the last 26 weeks
`)
}
