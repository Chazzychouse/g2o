package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Store struct {
	db   *sql.DB
	path string
}

// DefaultPath returns ~/.g2o/g2o.db.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".g2o", "g2o.db")
}

// Open creates the DB directory if needed, opens the SQLite database, runs
// migrations, and enables WAL mode + foreign keys.
func Open(path string) (*Store, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// SQLite pragmas for performance and safety.
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	} {
		if _, err := db.Exec(pragma); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("set %s: %w", pragma, err)
		}
	}

	s := &Store{db: db, path: path}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// IsEmpty returns true if no groups and no projects have been synced.
func (s *Store) IsEmpty() (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM groups").Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}
	err = s.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}
