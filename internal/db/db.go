package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// DefaultPath returns ~/.local/share/gtzy/gtzy.db, honoring GTZY_DB override.
func DefaultPath() (string, error) {
	if p := os.Getenv("GTZY_DB"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local", "share", "gtzy")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "gtzy.db"), nil
}

// Open opens the sqlite database at path, runs migrations, and seeds default data.
func Open(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("mkdir db dir: %w", err)
		}
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	conn.SetMaxOpenConns(1)

	if _, err := conn.Exec(schemaSQL); err != nil {
		conn.Close()
		return nil, fmt.Errorf("run schema: %w", err)
	}

	if err := seedCategories(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("seed categories: %w", err)
	}

	return conn, nil
}

func seedCategories(conn *sql.DB) error {
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339)
	defaults := []struct {
		name  string
		color string
	}{
		{"Work", "blue"},
		{"Learning", "green"},
		{"Personal", "mauve"},
		{"Health", "peach"},
	}
	for _, d := range defaults {
		if _, err := conn.Exec(
			`INSERT INTO categories (name, color, created_at) VALUES (?, ?, ?)`,
			d.name, d.color, now,
		); err != nil {
			return err
		}
	}
	return nil
}
