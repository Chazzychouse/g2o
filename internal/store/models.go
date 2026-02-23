package store

import "time"

type StoreGroup struct {
	ID          int64
	Name        string
	Path        string
	FullName    string
	FullPath    string
	Description string
	Visibility  string
	WebURL      string
	ParentID    int64
	CreatedAt   time.Time
	SyncedAt    time.Time
}

type StoreProject struct {
	ID                int64
	Name              string
	Path              string
	PathWithNamespace string
	NameWithNamespace string
	Description       string
	DefaultBranch     string
	Visibility        string
	WebURL            string
	NamespaceID       int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
	LastActivityAt    time.Time
	Archived          bool
	OpenIssuesCount   int64
	SyncedAt          time.Time
}

type StoreAssignee struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type StoreIssue struct {
	ID             int64
	IID            int64
	ProjectID      int64
	Title          string
	State          string
	Description    string
	WebURL         string
	AuthorID       int64
	AuthorName     string
	AuthorUsername string
	Labels         []string
	Assignees      []StoreAssignee
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ClosedAt       time.Time
	DueDate        string
	Weight         int64
	Confidential   bool
	SyncedAt       time.Time
}

type StoreUser struct {
	ID       int64
	Name     string
	Username string
	Email    string
	WebURL   string
	SyncedAt time.Time
}
