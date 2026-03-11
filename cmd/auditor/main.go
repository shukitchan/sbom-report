package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/shukitchan/sbom-report/internal/audit"
	"github.com/shukitchan/sbom-report/internal/githubapi"
	"github.com/shukitchan/sbom-report/internal/githubapp"
	"github.com/shukitchan/sbom-report/internal/report"
)

func main() {
	var (
		org       = flag.String("org", getenvDefault("GITHUB_ORG", ""), "GitHub org login to audit")
		reportDir = flag.String("report-dir", getenvDefault("REPORT_DIR", "/reports"), "Directory to write reports")
		allowCSV  = flag.String("allow", getenvDefault("ALLOWED_LICENSES", "MIT,Apache-2.0"), "Comma-separated SPDX allowlist")
		echoJSON  = flag.Bool("echo-json", getenvBoolDefault("ECHO_JSON", false), "Also print the JSON report to stdout")
	)
	flag.Parse()

	if *org == "" {
		log.Fatal("missing org (set --org or GITHUB_ORG)")
	}

	appID := os.Getenv("GITHUB_APP_ID")
	installationID := os.Getenv("GITHUB_APP_INSTALLATION_ID")
	privateKeyPEM := os.Getenv("GITHUB_APP_PRIVATE_KEY_PEM")
	if appID == "" || installationID == "" || privateKeyPEM == "" {
		log.Fatal("missing required env vars: GITHUB_APP_ID, GITHUB_APP_INSTALLATION_ID, GITHUB_APP_PRIVATE_KEY_PEM")
	}

	allowed := audit.ParseAllowed(*allowCSV)

	ctx := context.Background()

	appJWT, err := githubapp.NewAppJWT(appID, privateKeyPEM, time.Now)
	if err != nil {
		log.Fatalf("create app jwt: %v", err)
	}

	raw := githubapi.NewRawClient()
	token, err := raw.CreateInstallationToken(ctx, appJWT, installationID)
	if err != nil {
		log.Fatalf("create installation token: %v", err)
	}

	gh := githubapi.NewGitHub(token)

	a := audit.NewAuditor(gh, allowed)
	res, err := a.AuditOrg(ctx, *org)
	if err != nil {
		log.Fatalf("audit org: %v", err)
	}

	if err := os.MkdirAll(*reportDir, 0o755); err != nil {
		log.Fatalf("mkdir report dir: %v", err)
	}

	filename := fmt.Sprintf("%s-%s.json", *org, time.Now().UTC().Format("20060102-150405Z"))
	outPath := filepath.Join(*reportDir, filename)

	if err := report.WriteJSON(outPath, res); err != nil {
		log.Fatalf("write report: %v", err)
	}

	log.Printf("wrote report: %s (repos=%d)", outPath, len(res.Repositories))

	if *echoJSON {
		_ = json.NewEncoder(os.Stdout).Encode(res)
	}
}

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getenvBoolDefault(k string, def bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "TRUE", "yes", "YES", "y", "Y":
		return true
	case "0", "false", "FALSE", "no", "NO", "n", "N":
		return false
	default:
		return def
	}
}