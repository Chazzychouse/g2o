package glclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/chazzychouse/g2o/internal/styles"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Group struct{ *gitlab.Group }

const groupIssuesPerPage = 100
const groupIssuesState = "all"

func (g Group) String() string {
	return fmt.Sprintf("%s %s\n%s %s",
		styles.Label.Render("name: "),
		styles.Value.Render(g.Name),
		styles.Label.Render("gid:  "),
		styles.Value.Render(strconv.FormatInt(g.ID, 10)),
	)
}

func (g GitLab) MyGroups() ([]*gitlab.Group, error) {
	groups, _, err := g.client.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return nil, ErrListGroupsFailed
	}
	return groups, nil
}

func (g GitLab) GetGroup(id any) (*gitlab.Group, error) {
	group, _, err := g.client.Groups.GetGroup(id, nil)
	if err != nil {
		return nil, ErrGetGroupFailed
	}
	return group, nil
}

func (g GitLab) GetGroupsIssues(ctx context.Context, id any) (<-chan *gitlab.Issue, <-chan error) {
	ch := make(chan *gitlab.Issue)
	errc := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errc)
		page := int64(1)

		for {
			opts := &gitlab.ListGroupIssuesOptions{
				ListOptions: gitlab.ListOptions{
					PerPage: groupIssuesPerPage,
					Page:    page,
				},
				State: gitlab.Ptr(groupIssuesState),
			}

			issues, _, err := g.client.Issues.ListGroupIssues(id, opts)
			if err != nil {
				errc <- ErrListGroupIssuesFailed
				return
			}

			for _, issue := range issues {
				if b, err := json.MarshalIndent(issue, "", "  "); err == nil {
					g.log.Debug("issue fetched", "json", string(b))
				}
				if g.dump != nil {
					_ = g.dump.Encode(issue)
				}

				select {
				case ch <- issue:
				case <-ctx.Done():
					errc <- ctx.Err()
					return
				}
			}

			if len(issues) < groupIssuesPerPage {
				return
			}
			page++
		}
	}()

	return ch, errc
}

const perPage = 100

// AllGroups fetches all groups with pagination.
func (g GitLab) AllGroups(ctx context.Context) ([]*gitlab.Group, error) {
	var all []*gitlab.Group
	page := int64(1)
	for {
		groups, resp, err := g.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
			ListOptions: gitlab.ListOptions{PerPage: perPage, Page: page},
		})
		if err != nil {
			return nil, ErrListGroupsFailed
		}
		all = append(all, groups...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return all, nil
}

// AllGroupIssues fetches all issues for a group with optional UpdatedAfter filter.
func (g GitLab) AllGroupIssues(ctx context.Context, id any, updatedAfter *time.Time) ([]*gitlab.Issue, error) {
	var all []*gitlab.Issue
	page := int64(1)
	for {
		opts := &gitlab.ListGroupIssuesOptions{
			ListOptions: gitlab.ListOptions{PerPage: perPage, Page: page},
			State:       gitlab.Ptr(groupIssuesState),
		}
		if updatedAfter != nil {
			opts.UpdatedAfter = updatedAfter
		}
		issues, resp, err := g.client.Issues.ListGroupIssues(id, opts)
		if err != nil {
			return nil, ErrListGroupIssuesFailed
		}
		all = append(all, issues...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return all, nil
}

func listGroups(groups []*gitlab.Group) {
	fmt.Println(styles.Title.Render(fmt.Sprintf("Groups: %d", len(groups))))
	for _, g := range groups {
		fmt.Println(Group{g})
		fmt.Println()
	}
}
