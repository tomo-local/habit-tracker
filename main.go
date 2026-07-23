package main

import (
	"context"
	"fmt"
	"os"

	"habit-tracker/cmd"
	"habit-tracker/internal/calendar"
	"habit-tracker/internal/config"
)

func main() {
	ctx := context.Background()
	args := os.Args[1:]

	var command string
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	c := cmd.New(ctx, config.New(), calendar.New(ctx))

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
		fmt.Fprintf(os.Stderr, "unknown command: %q\nRun 'habit-tracker --help' for usage.\n", command)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Fprint(os.Stderr, `Usage: habit-tracker [command] [options]

Commands:
  (none)          Show heatmap for the first habit (no interactive selection)
  view            Select a habit and show its heatmap
  add [habit...]  Record today's habit
  setup           Select a calendar and register habit names
  auth login      Authenticate with your Google account
  help            Show this help message
  -h, --help      Show this help message
`)
}
