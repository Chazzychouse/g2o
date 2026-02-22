package glclient

import (
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Project struct{ *gitlab.Project }

func (p Project) String() string {
	return fmt.Sprintf("%s (%s)", p.Name, p.PathWithNamespace)
}

func (g GitLab) MyProjects() ([]*gitlab.Project, error) {
	projects, _, err := g.Client.Projects.ListProjects(&gitlab.ListProjectsOptions{
		Membership: gitlab.Ptr(true),
		Archived:   gitlab.Ptr(false),
	})
	if err != nil {
		return nil, ErrListProjectsFailed
	}

	var active []*gitlab.Project
	for _, p := range projects {
		if p.MarkedForDeletionOn == nil {
			active = append(active, p)
		}
	}
	return active, nil
}

func listProjects(projects []*gitlab.Project) {
	fmt.Printf("Projects: %d\n", len(projects))
	for _, p := range projects {
		fmt.Println(Project{p})
		fmt.Println()
	}
}
