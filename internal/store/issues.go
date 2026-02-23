package store

import (
	"database/sql"
	"encoding/json"
	"time"
)

func (s *Store) UpsertIssues(issues []StoreIssue) error {
	if len(issues) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO issues (id, iid, project_id, title, state, description, web_url,
			author_id, author_name, author_username, labels, assignees,
			created_at, updated_at, closed_at, due_date, weight, confidential, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			iid=excluded.iid, project_id=excluded.project_id, title=excluded.title,
			state=excluded.state, description=excluded.description, web_url=excluded.web_url,
			author_id=excluded.author_id, author_name=excluded.author_name,
			author_username=excluded.author_username, labels=excluded.labels,
			assignees=excluded.assignees, created_at=excluded.created_at,
			updated_at=excluded.updated_at, closed_at=excluded.closed_at,
			due_date=excluded.due_date, weight=excluded.weight,
			confidential=excluded.confidential, synced_at=excluded.synced_at`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	for _, issue := range issues {
		labelsJSON, _ := json.Marshal(issue.Labels)
		assigneesJSON, _ := json.Marshal(issue.Assignees)
		_, err := stmt.Exec(
			issue.ID, issue.IID, issue.ProjectID, issue.Title, issue.State,
			issue.Description, issue.WebURL, issue.AuthorID, issue.AuthorName,
			issue.AuthorUsername, string(labelsJSON), string(assigneesJSON),
			fmtTime(issue.CreatedAt), fmtTime(issue.UpdatedAt), fmtTime(issue.ClosedAt),
			issue.DueDate, issue.Weight, boolToInt(issue.Confidential), now,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListIssues() ([]StoreIssue, error) {
	rows, err := s.db.Query(`SELECT id, iid, project_id, title, state, description, web_url,
		author_id, author_name, author_username, labels, assignees,
		created_at, updated_at, closed_at, due_date, weight, confidential
		FROM issues ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIssues(rows)
}

func (s *Store) ListIssuesByGroup(groupID int64) ([]StoreIssue, error) {
	rows, err := s.db.Query(`SELECT i.id, i.iid, i.project_id, i.title, i.state, i.description, i.web_url,
		i.author_id, i.author_name, i.author_username, i.labels, i.assignees,
		i.created_at, i.updated_at, i.closed_at, i.due_date, i.weight, i.confidential
		FROM issues i
		JOIN group_issues gi ON gi.issue_id = i.id
		WHERE gi.group_id = ?
		ORDER BY i.updated_at DESC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIssues(rows)
}

func (s *Store) LinkGroupIssues(groupID int64, issueIDs []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Remove old links for this group, then re-insert.
	if _, err := tx.Exec("DELETE FROM group_issues WHERE group_id = ?", groupID); err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO group_issues (group_id, issue_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range issueIDs {
		if _, err := stmt.Exec(groupID, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) DeleteStaleIssues(activeIDs []int64) error {
	if len(activeIDs) == 0 {
		_, err := s.db.Exec("DELETE FROM issues")
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("CREATE TEMP TABLE _active_issues (id INTEGER PRIMARY KEY)"); err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO _active_issues (id) VALUES (?)")
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

	if _, err := tx.Exec("DELETE FROM issues WHERE id NOT IN (SELECT id FROM _active_issues)"); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM group_issues WHERE issue_id NOT IN (SELECT id FROM _active_issues)"); err != nil {
		return err
	}
	if _, err := tx.Exec("DROP TABLE _active_issues"); err != nil {
		return err
	}
	return tx.Commit()
}

func scanIssues(rows *sql.Rows) ([]StoreIssue, error) {
	var issues []StoreIssue
	for rows.Next() {
		var issue StoreIssue
		var labelsJSON, assigneesJSON string
		var confidential int
		if err := rows.Scan(
			&issue.ID, &issue.IID, &issue.ProjectID, &issue.Title, &issue.State,
			&issue.Description, &issue.WebURL, &issue.AuthorID, &issue.AuthorName,
			&issue.AuthorUsername, &labelsJSON, &assigneesJSON,
			&issue.CreatedAt, &issue.UpdatedAt, &issue.ClosedAt,
			&issue.DueDate, &issue.Weight, &confidential,
		); err != nil {
			return nil, err
		}
		issue.Confidential = confidential != 0
		_ = json.Unmarshal([]byte(labelsJSON), &issue.Labels)
		_ = json.Unmarshal([]byte(assigneesJSON), &issue.Assignees)
		issues = append(issues, issue)
	}
	return issues, rows.Err()
}
