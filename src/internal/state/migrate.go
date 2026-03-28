package state

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// ApplyMigrations creates schema_migrations and applies any pending numbered *.sql files in order.
func ApplyMigrations(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  applied_at INTEGER NOT NULL
);`); err != nil {
		return fmt.Errorf("schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		v, ok := migrationVersion(name)
		if !ok {
			return fmt.Errorf("migration %q: invalid name (expected NNNN_description.sql)", name)
		}
		applied, err := migrationApplied(db, v)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		body, err := migrationFS.ReadFile(path.Join("migrations", name))
		if err != nil {
			return fmt.Errorf("read migration %q: %w", name, err)
		}
		if err := execSQLScript(db, string(body)); err != nil {
			return fmt.Errorf("apply migration %q: %w", name, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, v, time.Now().UnixMilli()); err != nil {
			return fmt.Errorf("record migration %d: %w", v, err)
		}
	}
	return nil
}

func migrationVersion(filename string) (int, bool) {
	base := strings.TrimSuffix(filename, ".sql")
	part, _, ok := strings.Cut(base, "_")
	if !ok || part == "" {
		return 0, false
	}
	v, err := strconv.Atoi(part)
	if err != nil {
		return 0, false
	}
	return v, true
}

func migrationApplied(db *sql.DB, v int) (bool, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, v).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("migration %d applied check: %w", v, err)
	}
	return n > 0, nil
}

func execSQLScript(db *sql.DB, script string) error {
	for _, stmt := range strings.Split(script, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("%w\nstmt: %s", err, stmt)
		}
	}
	return nil
}
