package main

import (
	"fmt"
	"os"
	"strings"
)

const publicStatusPagesWorkflow = ".github/workflows/public-qa-status-pages.yml"

func verifyPublicStatusPagesWorkflow(root string) error {
	data, err := os.ReadFile(repoPath(root, publicStatusPagesWorkflow))
	if err != nil {
		return fmt.Errorf("read workflow %s: %w", publicStatusPagesWorkflow, err)
	}
	text := string(data)
	required := []string{
		"pages: write",
		"id-token: write",
		"cancel-in-progress: true",
		"Cancel stale GitHub Pages deployments",
		`"$api/deployments?environment=github-pages&per_page=20"`,
		`"$api/pages/deployments/$sha"`,
		"deployment_queued|deployment_in_progress|deployment_progress",
		`"$api/pages/deployments/$sha/cancel"`,
		"actions/upload-pages-artifact@v4",
		"actions/deploy-pages@v4",
		"timeout: 600000",
		"-evidence-out out/public-qa-status-pages-evidence.json",
		"name: public-qa-status-pages-evidence",
		"out/public-qa-status-pages-evidence.json",
		"name: public-qa-status-pages-site",
		"site/status.json",
		"site/status-badge.json",
		"site/pages-status.json",
		"source_workflow",
		"source_commit",
		"source_run_id",
		"source_run_url",
		"go test ./tools/publicpageslive -count=1",
		"go run ./tools/publicpageslive",
		`-base-url "${{ steps.deployment.outputs.page_url }}"`,
		"-out out/public-qa-status-pages-live-evidence.json",
		"name: public-qa-status-pages-live-evidence",
		"out/public-qa-status-pages-live-evidence.json",
		"raw_response_included: false",
		"secrets_included: false",
		"grep -R -E",
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			return fmt.Errorf("public status pages workflow missing %q", needle)
		}
	}
	return nil
}
