package store

import (
	"database/sql"
	"fmt"
	"time"
)

var migrations = []string{
	// v1: initial schema
	`CREATE TABLE IF NOT EXISTS sync_meta (
		resource_type TEXT PRIMARY KEY,
		last_synced_at TEXT NOT NULL,
		full_sync INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS groups (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL DEFAULT '',
		path TEXT NOT NULL DEFAULT '',
		full_name TEXT NOT NULL DEFAULT '',
		full_path TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		visibility TEXT NOT NULL DEFAULT '',
		web_url TEXT NOT NULL DEFAULT '',
		parent_id INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL DEFAULT '',
		synced_at TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL DEFAULT '',
		path TEXT NOT NULL DEFAULT '',
		path_with_namespace TEXT NOT NULL DEFAULT '',
		name_with_namespace TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		default_branch TEXT NOT NULL DEFAULT '',
		visibility TEXT NOT NULL DEFAULT '',
		web_url TEXT NOT NULL DEFAULT '',
		namespace_id INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL DEFAULT '',
		updated_at TEXT NOT NULL DEFAULT '',
		last_activity_at TEXT NOT NULL DEFAULT '',
		archived INTEGER NOT NULL DEFAULT 0,
		open_issues_count INTEGER NOT NULL DEFAULT 0,
		synced_at TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS issues (
		id INTEGER PRIMARY KEY,
		iid INTEGER NOT NULL DEFAULT 0,
		project_id INTEGER NOT NULL DEFAULT 0,
		title TEXT NOT NULL DEFAULT '',
		state TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		web_url TEXT NOT NULL DEFAULT '',
		author_id INTEGER NOT NULL DEFAULT 0,
		author_name TEXT NOT NULL DEFAULT '',
		author_username TEXT NOT NULL DEFAULT '',
		labels TEXT NOT NULL DEFAULT '[]',
		assignees TEXT NOT NULL DEFAULT '[]',
		created_at TEXT NOT NULL DEFAULT '',
		updated_at TEXT NOT NULL DEFAULT '',
		closed_at TEXT NOT NULL DEFAULT '',
		due_date TEXT NOT NULL DEFAULT '',
		weight INTEGER NOT NULL DEFAULT 0,
		confidential INTEGER NOT NULL DEFAULT 0,
		synced_at TEXT NOT NULL DEFAULT '',
		UNIQUE(project_id, iid)
	);

	CREATE INDEX IF NOT EXISTS idx_issues_project_id ON issues(project_id);
	CREATE INDEX IF NOT EXISTS idx_issues_state ON issues(state);
	CREATE INDEX IF NOT EXISTS idx_issues_updated_at ON issues(updated_at);

	CREATE TABLE IF NOT EXISTS group_issues (
		group_id INTEGER NOT NULL,
		issue_id INTEGER NOT NULL,
		PRIMARY KEY (group_id, issue_id)
	);

	CREATE TABLE IF NOT EXISTS current_user (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL DEFAULT '',
		username TEXT NOT NULL DEFAULT '',
		email TEXT NOT NULL DEFAULT '',
		web_url TEXT NOT NULL DEFAULT '',
		synced_at TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL
	);`,
}

func (s *Store) migrate() error {
	// Ensure schema_migrations exists (bootstrap).
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL
	)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	var current int
	row := s.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	if err := row.Scan(&current); err != nil {
		return fmt.Errorf("read migration version: %w", err)
	}

	for i := current; i < len(migrations); i++ {
		tx, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", i+1, err)
		}
		if _, err := tx.Exec(migrations[i]); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("run migration %d: %w", i+1, err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
			i+1, time.Now().UTC().Format(time.RFC3339)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %d: %w", i+1, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", i+1, err)
		}
	}
	return nil
}

func (s *Store) DB() *sql.DB { return s.db }
