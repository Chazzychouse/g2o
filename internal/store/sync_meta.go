package store

import (
	"database/sql"
	"errors"
	"time"
)

// GetLastSynced returns the last sync time for a resource type.
// Returns zero time if never synced.
func (s *Store) GetLastSynced(resourceType string) (time.Time, error) {
	var ts string
	err := s.db.QueryRow(
		"SELECT last_synced_at FROM sync_meta WHERE resource_type = ?",
		resourceType,
	).Scan(&ts)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// SetLastSynced records the last sync time for a resource type.
func (s *Store) SetLastSynced(resourceType string, t time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO sync_meta (resource_type, last_synced_at, full_sync)
		VALUES (?, ?, 0)
		ON CONFLICT(resource_type) DO UPDATE SET last_synced_at = excluded.last_synced_at`,
		resourceType, t.UTC().Format(time.RFC3339),
	)
	return err
}

// SetFullSync records a full sync timestamp for a resource type.
func (s *Store) SetFullSync(resourceType string, t time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO sync_meta (resource_type, last_synced_at, full_sync)
		VALUES (?, ?, 1)
		ON CONFLICT(resource_type) DO UPDATE SET last_synced_at = excluded.last_synced_at, full_sync = 1`,
		resourceType, t.UTC().Format(time.RFC3339),
	)
	return err
}

// GetLastFullSync returns the last full sync time for a resource type.
func (s *Store) GetLastFullSync(resourceType string) (time.Time, error) {
	var ts string
	var full int
	err := s.db.QueryRow(
		"SELECT last_synced_at, full_sync FROM sync_meta WHERE resource_type = ?",
		resourceType,
	).Scan(&ts, &full)
	if errors.Is(err, sql.ErrNoRows) || full == 0 {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
