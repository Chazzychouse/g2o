package glclient

import (
	"fmt"

	"github.com/chazzy/g2o/internal/styles"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Project struct{ *gitlab.Project }

func (p Project) String() string {
	return fmt.Sprintf("%s %s",
		styles.Value.Render(p.Name),
		styles.Label.Render("("+p.PathWithNamespace+")"))
}

func (g GitLab) MyProjects() ([]*gitlab.Project, error) {
	projects, _, err := g.client.Projects.ListProjects(&gitlab.ListProjectsOptions{
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
	fmt.Println(styles.Title.Render(fmt.Sprintf("Projects: %d", len(projects))))
	for _, p := range projects {
		fmt.Println(Project{p})
	}
}
