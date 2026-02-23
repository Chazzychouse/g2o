package glclient

import gitlab "gitlab.com/gitlab-org/api/client-go"

type GitLab struct {
	client *gitlab.Client
}

func NewGitlab(token string) (GitLab, error) {
	if token == "" {
		return GitLab{}, ErrTokenRequired
	}

	client, err := gitlab.NewClient(token)
	if err != nil {
		return GitLab{}, ErrClientCreationFailed
	}
	return GitLab{client: client}, nil
}

func (g GitLab) CurrentUser() (*gitlab.User, error) {
	user, _, err := g.client.Users.CurrentUser()
	if err != nil {
		return nil, ErrCurrentUserFailed
	}
	return user, nil
}
