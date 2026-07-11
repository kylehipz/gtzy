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

	if err := migrate(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	if err := seedCategories(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("seed categories: %w", err)
	}

	return conn, nil
}

// migrate applies idempotent schema fixups that CREATE TABLE IF NOT EXISTS can't
// make to pre-existing tables. Each step is guarded so it's a no-op on fresh or
// already-migrated databases.
func migrate(conn *sql.DB) error {
	// Estimates moved from whole minutes to seconds. On DBs that predate the
	// change the column is still `estimated_minutes`: rename it and convert the
	// stored values (minutes × 60). Guarded on the old column still existing.
	for _, table := range []string{"tasks", "recurrences"} {
		has, err := columnExists(conn, table, "estimated_minutes")
		if err != nil {
			return err
		}
		if !has {
			continue
		}
		if _, err := conn.Exec(
			fmt.Sprintf(`ALTER TABLE %s RENAME COLUMN estimated_minutes TO estimated_seconds`, table),
		); err != nil {
			return err
		}
		if _, err := conn.Exec(
			fmt.Sprintf(`UPDATE %s SET estimated_seconds = estimated_seconds * 60`, table),
		); err != nil {
			return err
		}
	}
	return nil
}

// columnExists reports whether table has a column named col.
func columnExists(conn *sql.DB, table, col string) (bool, error) {
	rows, err := conn.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			cid, notnull, pk int
			name, ctype      string
			dfltValue        sql.NullString
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if name == col {
			return true, nil
		}
	}
	return false, rows.Err()
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
