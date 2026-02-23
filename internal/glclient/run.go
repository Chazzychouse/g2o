package glclient

import (
	"context"
	"fmt"
	"strconv"

	"github.com/chazzychouse/g2o/internal/store"
	"github.com/chazzychouse/g2o/internal/styles"
)

func (g GitLab) RunGroups() error {
	if g.store != nil {
		groups, err := g.store.ListGroups()
		if err == nil && len(groups) > 0 {
			listStoreGroups(groups)
			return nil
		}
	}
	groups, err := g.MyGroups()
	if err != nil {
		return err
	}
	listGroups(groups)
	return nil
}

func (g GitLab) RunProjects() error {
	if g.store != nil {
		projects, err := g.store.ListProjects()
		if err == nil && len(projects) > 0 {
			listStoreProjects(projects)
			return nil
		}
	}
	projects, err := g.MyProjects()
	if err != nil {
		return err
	}
	listProjects(projects)
	return nil
}

func (g GitLab) RunCurrentUser() error {
	if g.store != nil {
		u, err := g.store.GetCurrentUser()
		if err == nil {
			fmt.Printf("%s %s\n%s %s\n%s %s\n",
				styles.Label.Render("name:    "), styles.Value.Render(u.Name),
				styles.Label.Render("username:"), styles.Value.Render(u.Username),
				styles.Label.Render("email:   "), styles.Value.Render(u.Email))
			return nil
		}
	}
	user, err := g.CurrentUser()
	if err != nil {
		return err
	}
	fmt.Printf("%s %s\n%s %s\n%s %s\n",
		styles.Label.Render("name:    "), styles.Value.Render(user.Name),
		styles.Label.Render("username:"), styles.Value.Render(user.Username),
		styles.Label.Render("email:   "), styles.Value.Render(user.Email))
	return nil
}

func (g GitLab) RunIssues() error {
	if g.store != nil {
		issues, err := g.store.ListIssues()
		if err == nil && len(issues) > 0 {
			listStoreIssues(issues)
			return nil
		}
	}
	issues, err := g.Issues()
	if err != nil {
		return err
	}
	listIssues(issues)
	return nil
}

func (g GitLab) RunGroupsIssues(ctx context.Context, id any) error {
	if g.store != nil {
		// Try parsing id as int64 for store lookup.
		if gid, err := toInt64(id); err == nil {
			issues, err := g.store.ListIssuesByGroup(gid)
			if err == nil && len(issues) > 0 {
				for _, issue := range issues {
					fmt.Printf("%s %s\n",
						styles.Value.Render(issue.Title),
						styles.Label.Render("("+strconv.FormatInt(issue.IID, 10)+")"))
				}
				fmt.Println(styles.Title.Render(fmt.Sprintf("Total: %d", len(issues))))
				return nil
			}
		}
	}

	ch, errc := g.GetGroupsIssues(ctx, id)
	count := 0
	for issue := range ch {
		fmt.Println(Issue{issue})
		count++
	}
	if err := <-errc; err != nil {
		return err
	}
	fmt.Println(styles.Title.Render(fmt.Sprintf("Total: %d", count)))
	return nil
}

func listStoreGroups(groups []store.StoreGroup) {
	fmt.Println(styles.Title.Render(fmt.Sprintf("Groups: %d", len(groups))))
	for _, g := range groups {
		fmt.Printf("%s %s\n%s %s\n\n",
			styles.Label.Render("name: "),
			styles.Value.Render(g.Name),
			styles.Label.Render("gid:  "),
			styles.Value.Render(strconv.FormatInt(g.ID, 10)))
	}
}

func listStoreProjects(projects []store.StoreProject) {
	fmt.Println(styles.Title.Render(fmt.Sprintf("Projects: %d", len(projects))))
	for _, p := range projects {
		fmt.Printf("%s %s\n",
			styles.Value.Render(p.Name),
			styles.Label.Render("("+p.PathWithNamespace+")"))
	}
}

func listStoreIssues(issues []store.StoreIssue) {
	fmt.Println(styles.Title.Render(fmt.Sprintf("Issues: %d", len(issues))))
	for _, i := range issues {
		fmt.Printf("%s %s\n",
			styles.Value.Render(i.Title),
			styles.Label.Render("("+strconv.FormatInt(i.IID, 10)+")"))
	}
}

func toInt64(v any) (int64, error) {
	switch val := v.(type) {
	case int:
		return int64(val), nil
	case int64:
		return val, nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}
