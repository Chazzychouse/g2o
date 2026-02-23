package glclient

import (
	"encoding/json"
	"io"
	"log/slog"

	"github.com/chazzychouse/g2o/internal/store"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GitLab struct {
	client *gitlab.Client
	log    *slog.Logger
	dump   *json.Encoder
	store  *store.Store
}

type Option func(*GitLab)

func WithLogger(l *slog.Logger) Option {
	return func(g *GitLab) { g.log = l }
}

func WithDump(w io.Writer) Option {
	return func(g *GitLab) { g.dump = json.NewEncoder(w) }
}

func WithStore(s *store.Store) Option {
	return func(g *GitLab) { g.store = s }
}

func NewGitlab(token string, opts ...Option) (GitLab, error) {
	if token == "" {
		return GitLab{}, ErrTokenRequired
	}

	client, err := gitlab.NewClient(token)
	if err != nil {
		return GitLab{}, ErrClientCreationFailed
	}

	g := GitLab{
		client: client,
		log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	for _, opt := range opts {
		opt(&g)
	}
	return g, nil
}

func (g GitLab) CurrentUser() (*gitlab.User, error) {
	user, _, err := g.client.Users.CurrentUser()
	if err != nil {
		return nil, ErrCurrentUserFailed
	}
	return user, nil
}
