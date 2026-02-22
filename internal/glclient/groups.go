package glclient

import (
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Group struct{ *gitlab.Group }

func (g Group) String() string {
	return fmt.Sprintf("name:        %s\npath:        %s\ndescription: %s\nvisibility:  %v",
		g.Name, g.FullPath, g.Description, g.Visibility)
}

func (g GitLab) MyGroups() ([]*gitlab.Group, error) {
	groups, _, err := g.Client.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return nil, ErrListGroupsFailed
	}
	return groups, nil
}

func (g GitLab) GetGroup(id any) (*gitlab.Group, error) {
	group, _, err := g.Client.Groups.GetGroup(id, nil)
	if err != nil {
		return nil, ErrGetGroupFailed
	}
	return group, nil
}

func listGroups(groups []*gitlab.Group) {
	fmt.Printf("Groups: %d\n", len(groups))
	for _, g := range groups {
		fmt.Println(Group{g})
		fmt.Println()
	}
}
