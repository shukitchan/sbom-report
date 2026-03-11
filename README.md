# sbom-report

Periodic SBOM license auditor for a GitHub Enterprise Cloud organization using a GitHub App installation token.

## What it does
- Lists repos in an org the App installation can access
- Fetches each repo SBOM via Dependency Graph (`/dependency-graph/sbom`)
- Flags dependency licenses not in allowlist (`MIT`, `Apache-2.0`)
- Lists `NOASSERTION` / missing license as `needs_manual_review`
- Writes JSON reports to `REPORT_DIR` (default: `/reports`)
- Does not create GitHub issues

## Configuration
Required env vars:
- `GITHUB_APP_ID`
- `GITHUB_APP_INSTALLATION_ID`
- `GITHUB_APP_PRIVATE_KEY_PEM`
- `GITHUB_ORG` (e.g. `yahoo-news`)

Optional env vars:
- `REPORT_DIR` (default `/reports`)
- `ALLOWED_LICENSES` (default `MIT,Apache-2.0`)
- `ECHO_JSON` (default `false`)

## Local run
```bash
mkdir -p ./reports

GITHUB_ORG=yahoo-news \
REPORT_DIR=./reports \
GITHUB_APP_ID=12345 \
GITHUB_APP_INSTALLATION_ID=67890 \
GITHUB_APP_PRIVATE_KEY_PEM="$(cat private-key.pem)" \
go run ./cmd/auditor
```

## Kubernetes (weekly CronJob)
See `deploy/kustomize`.

Default schedule is weekly on Sunday at 02:00 UTC (`0 2 * * 0`).

```bash
kubectl apply -k deploy/kustomize
```