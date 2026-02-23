package sync

import (
	"time"

	"github.com/chazzychouse/g2o/internal/store"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func convertGroups(groups []*gitlab.Group) []store.StoreGroup {
	out := make([]store.StoreGroup, len(groups))
	for i, g := range groups {
		out[i] = store.StoreGroup{
			ID:          g.ID,
			Name:        g.Name,
			Path:        g.Path,
			FullName:    g.FullName,
			FullPath:    g.FullPath,
			Description: g.Description,
			Visibility:  string(g.Visibility),
			WebURL:      g.WebURL,
			ParentID:    g.ParentID,
			CreatedAt:   ptrTime(g.CreatedAt),
		}
	}
	return out
}

func convertProjects(projects []*gitlab.Project) []store.StoreProject {
	out := make([]store.StoreProject, len(projects))
	for i, p := range projects {
		var nsID int64
		if p.Namespace != nil {
			nsID = p.Namespace.ID
		}
		out[i] = store.StoreProject{
			ID:                p.ID,
			Name:              p.Name,
			Path:              p.Path,
			PathWithNamespace: p.PathWithNamespace,
			NameWithNamespace: p.NameWithNamespace,
			Description:       p.Description,
			DefaultBranch:     p.DefaultBranch,
			Visibility:        string(p.Visibility),
			WebURL:            p.WebURL,
			NamespaceID:       nsID,
			CreatedAt:         ptrTime(p.CreatedAt),
			UpdatedAt:         ptrTime(p.UpdatedAt),
			LastActivityAt:    ptrTime(p.LastActivityAt),
			Archived:          p.Archived,
			OpenIssuesCount:   p.OpenIssuesCount,
		}
	}
	return out
}

func convertIssues(issues []*gitlab.Issue) []store.StoreIssue {
	out := make([]store.StoreIssue, len(issues))
	for i, issue := range issues {
		si := store.StoreIssue{
			ID:           issue.ID,
			IID:          issue.IID,
			ProjectID:    issue.ProjectID,
			Title:        issue.Title,
			State:        issue.State,
			Description:  issue.Description,
			WebURL:       issue.WebURL,
			Labels:       []string(issue.Labels),
			CreatedAt:    ptrTime(issue.CreatedAt),
			UpdatedAt:    ptrTime(issue.UpdatedAt),
			ClosedAt:     ptrTime(issue.ClosedAt),
			Weight:       issue.Weight,
			Confidential: issue.Confidential,
		}
		if issue.DueDate != nil {
			si.DueDate = time.Time(*issue.DueDate).Format("2006-01-02")
		}
		if issue.Author != nil {
			si.AuthorID = issue.Author.ID
			si.AuthorName = issue.Author.Name
			si.AuthorUsername = issue.Author.Username
		}
		for _, a := range issue.Assignees {
			si.Assignees = append(si.Assignees, store.StoreAssignee{
				ID:       a.ID,
				Name:     a.Name,
				Username: a.Username,
			})
		}
		if si.Labels == nil {
			si.Labels = []string{}
		}
		if si.Assignees == nil {
			si.Assignees = []store.StoreAssignee{}
		}
		out[i] = si
	}
	return out
}

func convertUser(u *gitlab.User) store.StoreUser {
	return store.StoreUser{
		ID:       u.ID,
		Name:     u.Name,
		Username: u.Username,
		Email:    u.Email,
		WebURL:   u.WebURL,
	}
}

func ptrTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
