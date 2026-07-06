package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"gtzy/internal/api"
	"gtzy/internal/cli"
	"gtzy/internal/db"
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
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8420, "port to listen on")
	dbPath := fs.String("db", "", "path to sqlite database file")
	fs.Parse(args)

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

	server := &api.Server{DB: conn}
	handler := api.NewRouter(server, nil)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("gtzy listening on %s (db: %s)", addr, path)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Println("server error:", err)
		return 1
	}
	return 0
}
