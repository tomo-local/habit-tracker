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
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\nusage: habit-tracker [view|auth|setup|add]\n", command)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
