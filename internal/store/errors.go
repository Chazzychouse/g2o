package store

import "errors"

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrDatabaseClosed = errors.New("database is closed")
)
