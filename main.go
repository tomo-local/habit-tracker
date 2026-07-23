package main

import (
	"fmt"
	"os"

	"habit-tracker/cmd"
)

func main() {
	args := os.Args[1:]

	var command string
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	var err error
	switch command {
	case "":
		err = cmd.RunView(args, true)
	case "view":
		err = cmd.RunView(args, false)
	case "auth":
		err = cmd.RunAuth(args)
	case "setup":
		err = cmd.RunSetup(args)
	case "add":
		err = cmd.RunAdd(args)
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
