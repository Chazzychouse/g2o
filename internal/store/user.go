package store

import (
	"database/sql"
	"errors"
	"time"
)

func (s *Store) UpsertUser(u StoreUser) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`
		INSERT INTO current_user (id, name, username, email, web_url, synced_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, username=excluded.username, email=excluded.email,
			web_url=excluded.web_url, synced_at=excluded.synced_at`,
		u.ID, u.Name, u.Username, u.Email, u.WebURL, now,
	)
	return err
}

func (s *Store) GetCurrentUser() (StoreUser, error) {
	var u StoreUser
	err := s.db.QueryRow("SELECT id, name, username, email, web_url FROM current_user LIMIT 1").
		Scan(&u.ID, &u.Name, &u.Username, &u.Email, &u.WebURL)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrRecordNotFound
	}
	return u, err
}
