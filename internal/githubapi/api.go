package githubapi

import (
	"context"
	"net/http"

	"github.com/google/go-github/v66/github"
)

type GitHub struct {
	client *github.Client
}

func NewGitHub(installationToken string) *GitHub {
	hc := &http.Client{
		Transport: &authTransport{
			token: installationToken,
			base:  http.DefaultTransport,
		},
	}
	return &GitHub{client: github.NewClient(hc)}
}

type authTransport struct {
	token string
	base  http.RoundTripper
}

func (t *authTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	r2.Header.Set("Authorization", "token " + t.token)
	r2.Header.Set("Accept", "application/vnd.github+json")
	r2.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return t.base.RoundTrip(r2)
}

func (g *GitHub) ListOrgRepos(ctx context.Context, org string) ([]*github.Repository, error) {
	opt := &github.RepositoryListByOrgOptions{
		Type: "all",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var all []*github.Repository
	for {
		repos, resp, err := g.client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		all = append(all, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		}
	return all, nil
}