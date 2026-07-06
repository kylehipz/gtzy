package cli

import "github.com/spf13/cobra"

// NewRootCommand builds the "gtzy" root command with every subcommand
// attached except "serve", which lives in package main (it depends on
// internal/api, internal/db, internal/ai and the embedded web frontend, so
// wiring it here would pull those into internal/cli unnecessarily). The
// caller (main.go) is expected to do:
//
//	root := cli.NewRootCommand()
//	root.AddCommand(newServeCmd())
//	root.Execute()
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "gtzy",
		Short: "A local task manager and one-task-at-a-time timer",
		Long: `gtzy is a single-user, locally-run productivity app: a task manager with a
strict one-task-at-a-time timer, a calendar-based progress tracker, a
journal, and an analytics dashboard with an optional AI growth summary.

One binary is both the HTTP server ("gtzy serve") and this CLI. Every
subcommand other than "serve" is a thin HTTP client that talks to a running
"gtzy serve" instance (default http://localhost:8420, override with the
GTZY_URL environment variable).`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(
		newNextCmd(),
		newStartCmd(),
		newPauseCmd(),
		newCompleteCmd(),
		newCurrentCmd(),
		newAddCmd(),
		newListCmd(),
		newWaybarCmd(),
	)
	return root
}
