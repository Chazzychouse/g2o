package glclient

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/chazzy/g2o/internal/styles"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Issue struct{ *gitlab.Issue }

func (i Issue) String() string {
	return fmt.Sprintf("%s %s",
		styles.Value.Render(i.Title),
		styles.Label.Render("("+strconv.FormatInt(i.IID, 10)+")"))
}

func (g GitLab) Issues() ([]*gitlab.Issue, error) {
	issues, _, err := g.client.Issues.ListIssues(&gitlab.ListIssuesOptions{})
	if err != nil {
		return nil, ErrListIssuesFailed
	}
	return issues, nil
}

// AllIssues fetches all issues with pagination and optional UpdatedAfter filter.
func (g GitLab) AllIssues(ctx context.Context, updatedAfter *time.Time) ([]*gitlab.Issue, error) {
	var all []*gitlab.Issue
	page := int64(1)
	for {
		opts := &gitlab.ListIssuesOptions{
			ListOptions: gitlab.ListOptions{PerPage: perPage, Page: page},
			Scope:       gitlab.Ptr("assigned_to_me"),
		}
		if updatedAfter != nil {
			opts.UpdatedAfter = updatedAfter
		}
		issues, resp, err := g.client.Issues.ListIssues(opts)
		if err != nil {
			return nil, ErrListIssuesFailed
		}
		all = append(all, issues...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return all, nil
}

func listIssues(issues []*gitlab.Issue) {
	fmt.Println(styles.Title.Render(fmt.Sprintf("Issues: %d", len(issues))))
	for _, i := range issues {
		fmt.Println(Issue{i})
	}
}
