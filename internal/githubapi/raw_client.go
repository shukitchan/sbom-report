package githubapi

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type RawClient struct {
    hc *http.Client
}

func NewRawClient() *RawClient {
    return &RawClient{hc: &http.Client{Timeout: 30 * time.Second}}
}

type installationTokenResp struct {
    Token string `json:"token"`
}

// CreateInstallationToken exchanges an app JWT for an installation access token.
// POST /app/installations/{installation_id}/access_tokens
func (c *RawClient) CreateInstallationToken(ctx context.Context, appJWT string, installationID string) (string, error) {
    url := fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens", installationID)

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(nil))
    if err != nil {
        return "", err
    }
    req.Header.Set("Authorization", "Bearer "+appJWT)
    req.Header.Set("Accept", "application/vnd.github+json")
    req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

    resp, err := c.hc.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    b, _ := io.ReadAll(resp.Body)
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return "", fmt.Errorf("create installation token failed: status=%d body=%s", resp.StatusCode, string(b))
    }

    var tr installationTokenResp
    if err := json.Unmarshal(b, &tr); err != nil {
        return "", err
    }
    if tr.Token == "" {
        return "", fmt.Errorf("missing token in response")
    }
    return tr.Token, nil
}