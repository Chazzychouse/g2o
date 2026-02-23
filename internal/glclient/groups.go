package glclient

import (
	"fmt"

	"github.com/chazzy/g2o/internal/styles"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Group struct{ *gitlab.Group }

func (g Group) String() string {
	return fmt.Sprintf("%s %s\n%s %s\n%s %s\n%s %v",
		styles.Label.Render("name:       "), styles.Value.Render(g.Name),
		styles.Label.Render("path:       "), styles.Value.Render(g.FullPath),
		styles.Label.Render("description:"), styles.Value.Render(g.Description),
		styles.Label.Render("visibility: "), styles.Value.Render(string(g.Visibility)))
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

func listGroups(groups []*gitlab.Group) {
	fmt.Println(styles.Title.Render(fmt.Sprintf("Groups: %d", len(groups))))
	for _, g := range groups {
		fmt.Println(Group{g})
		fmt.Println()
	}
}
