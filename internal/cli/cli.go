package cli

import (
	"fmt"
	"os"
)

// BaseURL returns the server URL to talk to, honoring GTZY_URL.
func BaseURL() string {
	if u := os.Getenv("GTZY_URL"); u != "" {
		return u
	}
	return "http://localhost:8420"
}

// Run dispatches a CLI subcommand (everything except "serve").
func Run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: gtzy <command> [args]")
		return 1
	}
	cmd := args[0]
	rest := args[1:]

	var err error
	switch cmd {
	case "next":
		err = cmdNext(rest)
	case "start":
		err = cmdStart(rest)
	case "pause":
		err = cmdPause(rest)
	case "complete":
		err = cmdComplete(rest)
	case "current":
		err = cmdCurrent(rest)
	case "add":
		err = cmdAdd(rest)
	case "list":
		err = cmdList(rest)
	case "waybar":
		err = cmdWaybar(rest)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		return 1
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}
