package glclient

import "fmt"

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
	fmt.Printf("name:     %s\nusername: %s\nemail:    %s\n", user.Name, user.Username, user.Email)
	return nil
}
