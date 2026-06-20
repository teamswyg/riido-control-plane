package deploypolicy

import (
	"regexp"
	"testing"
)

func assertNoForbiddenRegex(t *testing.T, repoPath, body string, patterns []string) {
	t.Helper()
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			t.Fatalf("compile public CD forbidden regex %q: %v", pattern, err)
		}
		if match := re.FindString(body); match != "" {
			t.Fatalf("%s contains forbidden public CD pattern %q via %q", repoPath, pattern, match)
		}
	}
}

func liveWorkflowPaths() []string {
	return []string{
		".github/workflows/deploy-ai-agent-testnet.yml",
		".github/workflows/ai-agent-client-testnet-smoke.yml",
	}
}
