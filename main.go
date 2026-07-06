package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"gtzy/internal/ai"
	"gtzy/internal/api"
	"gtzy/internal/cli"
	"gtzy/internal/db"
	"gtzy/web"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gtzy <serve|next|start|pause|complete|current|add|list|waybar> [args]")
		os.Exit(1)
	}

	if os.Args[1] == "serve" {
		os.Exit(runServe(os.Args[2:]))
	}

	os.Exit(cli.Run(os.Args[1:]))
}

func runServe(args []string) int {
	flagSet := flag.NewFlagSet("serve", flag.ExitOnError)
	port := flagSet.Int("port", 8420, "port to listen on")
	dbPath := flagSet.String("db", "", "path to sqlite database file")
	flagSet.Parse(args)

	path := *dbPath
	if path == "" {
		p, err := db.DefaultPath()
		if err != nil {
			log.Println("resolve db path:", err)
			return 1
		}
		path = p
	}

	conn, err := db.Open(path)
	if err != nil {
		log.Println("open db:", err)
		return 1
	}
	defer conn.Close()

	server := &api.Server{DB: conn, AI: ai.New()}
	handler := api.NewRouter(server, spaFS())

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("gtzy listening on %s (db: %s)", addr, path)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Println("server error:", err)
		return 1
	}
	return 0
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
