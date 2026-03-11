package githubapi

import (
    "context"
    "fmt"
)

// FetchSBOM fetches Dependency Graph SBOM (SPDX JSON) for a repo.
// Endpoint: GET /repos/{owner}/{repo}/dependency-graph/sbom
func (g *GitHub) FetchSBOM(ctx context.Context, owner, repo string) (map[string]any, error) {
    var out map[string]any

    req, err := g.client.NewRequest("GET", fmt.Sprintf("repos/%s/%s/dependency-graph/sbom", owner, repo), nil)
    if err != nil {
        return nil, err
    }
    _, err = g.client.Do(ctx, req, &out)
    if err != nil {
        return nil, err
    }
    return out, nil
}