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
		"actions/upload-pages-artifact@v4",
		"actions/deploy-pages@v4",
		"-evidence-out out/public-qa-status-pages-evidence.json",
		"name: public-qa-status-pages-evidence",
		"out/public-qa-status-pages-evidence.json",
		"name: public-qa-status-pages-site",
		"site/status.json",
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
