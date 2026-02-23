package store

import (
	"database/sql"
	"errors"
	"time"
)

func (s *Store) UpsertProjects(projects []StoreProject) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO projects (id, name, path, path_with_namespace, name_with_namespace,
			description, default_branch, visibility, web_url, namespace_id,
			created_at, updated_at, last_activity_at, archived, open_issues_count, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, path=excluded.path,
			path_with_namespace=excluded.path_with_namespace,
			name_with_namespace=excluded.name_with_namespace,
			description=excluded.description, default_branch=excluded.default_branch,
			visibility=excluded.visibility, web_url=excluded.web_url,
			namespace_id=excluded.namespace_id, created_at=excluded.created_at,
			updated_at=excluded.updated_at, last_activity_at=excluded.last_activity_at,
			archived=excluded.archived, open_issues_count=excluded.open_issues_count,
			synced_at=excluded.synced_at`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	for _, p := range projects {
		_, err := stmt.Exec(p.ID, p.Name, p.Path, p.PathWithNamespace, p.NameWithNamespace,
			p.Description, p.DefaultBranch, p.Visibility, p.WebURL, p.NamespaceID,
			fmtTime(p.CreatedAt), fmtTime(p.UpdatedAt), fmtTime(p.LastActivityAt),
			boolToInt(p.Archived), p.OpenIssuesCount, now)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListProjects() ([]StoreProject, error) {
	rows, err := s.db.Query(`SELECT id, name, path, path_with_namespace, name_with_namespace,
		description, default_branch, visibility, web_url, namespace_id, archived, open_issues_count
		FROM projects WHERE archived = 0 ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []StoreProject
	for rows.Next() {
		var p StoreProject
		var archived int
		if err := rows.Scan(&p.ID, &p.Name, &p.Path, &p.PathWithNamespace, &p.NameWithNamespace,
			&p.Description, &p.DefaultBranch, &p.Visibility, &p.WebURL, &p.NamespaceID,
			&archived, &p.OpenIssuesCount); err != nil {
			return nil, err
		}
		p.Archived = archived != 0
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (s *Store) GetProject(id int64) (StoreProject, error) {
	var p StoreProject
	var archived int
	err := s.db.QueryRow(`SELECT id, name, path, path_with_namespace, name_with_namespace,
		description, default_branch, visibility, web_url, namespace_id, archived, open_issues_count
		FROM projects WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Path, &p.PathWithNamespace, &p.NameWithNamespace,
		&p.Description, &p.DefaultBranch, &p.Visibility, &p.WebURL, &p.NamespaceID,
		&archived, &p.OpenIssuesCount)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrRecordNotFound
	}
	p.Archived = archived != 0
	return p, err
}

func (s *Store) DeleteStaleProjects(activeIDs []int64) error {
	if len(activeIDs) == 0 {
		_, err := s.db.Exec("DELETE FROM projects")
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("CREATE TEMP TABLE _active_projects (id INTEGER PRIMARY KEY)"); err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO _active_projects (id) VALUES (?)")
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

	if _, err := tx.Exec("DELETE FROM projects WHERE id NOT IN (SELECT id FROM _active_projects)"); err != nil {
		return err
	}
	if _, err := tx.Exec("DROP TABLE _active_projects"); err != nil {
		return err
	}
	return tx.Commit()
}

func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
