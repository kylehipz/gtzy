package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gtzy/internal/ai"
	"gtzy/internal/api"
	"gtzy/internal/cli"
	"gtzy/internal/db"
	"gtzy/internal/meter"
	"gtzy/web"

	"github.com/spf13/cobra"
)

func main() {
	root := cli.NewRootCommand()
	root.AddCommand(newServeCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func newServeCmd() *cobra.Command {
	var port int
	var dbPath string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the gtzy HTTP server and web UI",
		Long: `Serves the gtzy REST API and, if the frontend has been built into web/dist,
the embedded React single-page app, backed by a local SQLite database at
--db (or $GTZY_DB, default ~/.local/share/gtzy/gtzy.db).

Leave this running; every other gtzy subcommand talks to it over HTTP.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(port, dbPath)
		},
	}
	cmd.Flags().IntVar(&port, "port", 8420, "port to listen on")
	cmd.Flags().StringVar(&dbPath, "db", "", "path to sqlite database file (default: $GTZY_DB or ~/.local/share/gtzy/gtzy.db)")
	return cmd
}

func runServe(port int, dbPath string) error {
	path := dbPath
	if path == "" {
		p, err := db.DefaultPath()
		if err != nil {
			return fmt.Errorf("resolve db path: %w", err)
		}
		path = p
	}

	conn, err := db.Open(path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()

	server := &api.Server{DB: conn, AI: ai.New()}
	handler := api.NewRouter(server, spaFS())

	// Optional background auto-sync from a Bluetooth glucose meter. Opt-in so
	// headless / no-Bluetooth hosts are unaffected.
	if watchEnabled() {
		w := &meter.Watcher{DB: conn, Interval: watchInterval(), OnImport: meter.NotifyImport}
		w.Start(context.Background())
		log.Printf("meter watch enabled (auto-sync on advertise, every %s)", watchInterval())
	}

	addr := fmt.Sprintf(":%d", port)
	log.Printf("gtzy listening on %s (db: %s)", addr, path)
	if err := http.ListenAndServe(addr, handler); err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// watchEnabled reports whether background meter auto-sync is turned on via
// GTZY_METER_WATCH.
func watchEnabled() bool {
	switch strings.ToLower(os.Getenv("GTZY_METER_WATCH")) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// watchInterval is the cooldown between meter scan cycles, from
// GTZY_METER_WATCH_INTERVAL (seconds), defaulting to 10s.
func watchInterval() time.Duration {
	if s := os.Getenv("GTZY_METER_WATCH_INTERVAL"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 10 * time.Second
}

// spaFS returns the embedded frontend build, or nil if it hasn't been built
// yet (dist/ contains only the placeholder .gitkeep), in which case the
// server runs API-only.
func spaFS() fs.FS {
	sub, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		return nil
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil
	}
	return sub
}
