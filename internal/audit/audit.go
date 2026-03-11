package audit

import (
    "context"
    "sort"
    "strings"
    "time"

    "github.com/google/go-github/v66/github"
)

type GitHub interface {
    ListOrgRepos(ctx context.Context, org string) ([]*github.Repository, error)
    FetchSBOM(ctx context.Context, owner, repo string) (map[string]any, error)
}

type Auditor struct {
    gh      GitHub
    allowed map[string]struct{}
}

func NewAuditor(gh GitHub, allowed map[string]struct{}) *Auditor {
    return &Auditor{gh: gh, allowed: allowed}
}

type Result struct {
    Org          string       `json:"org"`
    GeneratedAt  time.Time    `json:"generated_at"`
    Repositories []RepoResult `json:"repositories"`
}

type RepoResult struct {
    Owner             string    `json:"owner"`
    Repo              string    `json:"repo"`
    NotAllowed        []Finding `json:"not_allowed"`
    NeedsManualReview []Finding `json:"needs_manual_review"`
    Error             string    `json:"error,omitempty"`
}

type Finding struct {
    Package string `json:"package"`
    Version string `json:"version,omitempty"`
    License string `json:"license"`
}

func ParseAllowed(csv string) map[string]struct{} {
    m := map[string]struct{}{}
    for _, p := range strings.Split(csv, ",") {
        p = strings.TrimSpace(p)
        if p != "" {
            m[p] = struct{}{}
        }
    }
    return m
}

func (a *Auditor) AuditOrg(ctx context.Context, org string) (*Result, error) {
    repos, err := a.gh.ListOrgRepos(ctx, org)
    if err != nil {
        return nil, err
    }

    out := &Result{Org: org, GeneratedAt: time.Now().UTC()}

    for _, r := range repos {
        owner := ""
        if r.Owner != nil {
            owner = r.Owner.GetLogin()
        }
        repoName := r.GetName()

        rr := RepoResult{Owner: owner, Repo: repoName}

        sbom, err := a.gh.FetchSBOM(ctx, owner, repoName)
        if err != nil {
            rr.Error = err.Error()
            out.Repositories = append(out.Repositories, rr)
            continue
        }

        pkgs := extractPackages(sbom)
        for _, p := range pkgs {
            lic := strings.TrimSpace(p.License)

            if lic == "" || strings.EqualFold(lic, "NOASSERTION") {
                rr.NeedsManualReview = append(rr.NeedsManualReview, Finding{Package: p.Name, Version: p.Version, License: emptyToManualReviewLabel(lic)})
                continue
            }

            if _, ok := a.allowed[lic]; ok {
                continue
            }

            rr.NotAllowed = append(rr.NotAllowed, Finding{Package: p.Name, Version: p.Version, License: lic})
        }

        sort.Slice(rr.NotAllowed, func(i, j int) bool { return rr.NotAllowed[i].Package < rr.NotAllowed[j].Package })
        sort.Slice(rr.NeedsManualReview, func(i, j int) bool { return rr.NeedsManualReview[i].Package < rr.NeedsManualReview[j].Package })

        out.Repositories = append(out.Repositories, rr)
    }

    sort.Slice(out.Repositories, func(i, j int) bool {
        if out.Repositories[i].Owner == out.Repositories[j].Owner {
            return out.Repositories[i].Repo < out.Repositories[j].Repo
        }
        return out.Repositories[i].Owner < out.Repositories[j].Owner
    })

    return out, nil
}

type pkg struct {
    Name    string
    Version string
    License string
}

func extractPackages(sbom map[string]any) []pkg {
    packagesAny, _ := sbom["packages"].([]any)
    out := make([]pkg, 0, len(packagesAny))
    for _, p := range packagesAny {
        pm, ok := p.(map[string]any)
        if !ok {
            continue
        }

        name := toString(pm["name"])
        version := toString(pm["versionInfo"])

        lic := toString(pm["licenseConcluded"])
        if lic == "" || strings.EqualFold(lic, "NOASSERTION") {
            decl := toString(pm["licenseDeclared"])
            if decl != "" {
                lic = decl
            }
        }

        out = append(out, pkg{Name: name, Version: version, License: lic})
    }
    return out
}

func toString(v any) string {
    s, _ := v.(string)
    return strings.TrimSpace(s)
}

func emptyToManualReviewLabel(lic string) string {
    if strings.TrimSpace(lic) == "" {
        return "MISSING"
    }
    return lic
}