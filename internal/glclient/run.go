package glclient

import (
	"fmt"

	"github.com/chazzy/g2o/internal/styles"
)

func (g GitLab) RunGroups() error {
	groups, err := g.MyGroups()
	if err != nil {
		return err
	}
	listGroups(groups)
	return nil
}

func (g GitLab) RunProjects() error {
	projects, err := g.MyProjects()
	if err != nil {
		return err
	}
	listProjects(projects)
	return nil
}

func (g GitLab) RunCurrentUser() error {
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
