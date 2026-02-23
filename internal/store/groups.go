package store

import (
	"database/sql"
	"errors"
	"time"
)

func (s *Store) UpsertGroups(groups []StoreGroup) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO groups (id, name, path, full_name, full_path, description, visibility, web_url, parent_id, created_at, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, path=excluded.path, full_name=excluded.full_name,
			full_path=excluded.full_path, description=excluded.description,
			visibility=excluded.visibility, web_url=excluded.web_url,
			parent_id=excluded.parent_id, created_at=excluded.created_at, synced_at=excluded.synced_at`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	for _, g := range groups {
		createdAt := ""
		if !g.CreatedAt.IsZero() {
			createdAt = g.CreatedAt.Format(time.RFC3339)
		}
		_, err := stmt.Exec(g.ID, g.Name, g.Path, g.FullName, g.FullPath,
			g.Description, g.Visibility, g.WebURL, g.ParentID, createdAt, now)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListGroups() ([]StoreGroup, error) {
	rows, err := s.db.Query("SELECT id, name, path, full_name, full_path, description, visibility, web_url, parent_id FROM groups ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []StoreGroup
	for rows.Next() {
		var g StoreGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Path, &g.FullName, &g.FullPath,
			&g.Description, &g.Visibility, &g.WebURL, &g.ParentID); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (s *Store) GetGroup(id int64) (StoreGroup, error) {
	var g StoreGroup
	err := s.db.QueryRow(
		"SELECT id, name, path, full_name, full_path, description, visibility, web_url, parent_id FROM groups WHERE id = ?", id,
	).Scan(&g.ID, &g.Name, &g.Path, &g.FullName, &g.FullPath,
		&g.Description, &g.Visibility, &g.WebURL, &g.ParentID)
	if errors.Is(err, sql.ErrNoRows) {
		return g, ErrRecordNotFound
	}
	return g, err
}

// DeleteStaleGroups removes groups whose IDs are not in the given set.
func (s *Store) DeleteStaleGroups(activeIDs []int64) error {
	if len(activeIDs) == 0 {
		_, err := s.db.Exec("DELETE FROM groups")
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Build temp table of active IDs.
	if _, err := tx.Exec("CREATE TEMP TABLE _active_groups (id INTEGER PRIMARY KEY)"); err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO _active_groups (id) VALUES (?)")
	if err != nil {
		return err
	}
	for _, id := range activeIDs {
		if _, err := stmt.Exec(id); err != nil {
			stmt.Close()
			return err
		}
	}
	stmt.Close()

	if _, err := tx.Exec("DELETE FROM groups WHERE id NOT IN (SELECT id FROM _active_groups)"); err != nil {
		return err
	}
	if _, err := tx.Exec("DROP TABLE _active_groups"); err != nil {
		return err
	}
	return tx.Commit()
}
