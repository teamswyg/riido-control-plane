package aiagentrisk

import (
	"testing"
)

const (
	manifestPath = "../../docs/30-architecture/ai-agent-risk-evidence.riido.json"
	humanDocPath = "../../docs/30-architecture/api-client-delivery.md"
)

var requiredRisks = []string{
	"private-worktree-url-redaction",
	"active-stream-handoff",
	"terminal-late-progress-fence",
	"assign-reconcile-scope",
	"stale-terminal-busy-count",
	"partial-progress-coalescing",
	"projection-read-repair",
	"additive-agent-active",
	"generated-fsm-control-plane-consumption",
	"generated-fsm-conformance",
}

var requiredBoundaries = []string{
	"private-repo-auth",
	"web-approval-round-trip",
	"client-active-stream-consumption",
}

func TestAIAgentRiskEvidenceManifest(t *testing.T) {
	manifest := loadManifest(t, manifestPath)
	doc := mustRead(t, humanDocPath)

	assertManifestHeader(t, manifest, doc)
	seen := map[string]bool{}
	for _, evidence := range manifest.LocalEvidence {
		assertLocalEvidence(t, evidence, doc)
		seen[evidence.Risk] = true
	}
	for _, evidence := range manifest.ExternalEvidence {
		assertExternalEvidence(t, evidence, doc)
		seen[evidence.Risk] = true
	}
	for _, risk := range requiredRisks {
		if !seen[risk] {
			t.Fatalf("manifest missing risk evidence %q", risk)
		}
	}

	boundaries := map[string]bool{}
	for _, boundary := range manifest.RemainingBoundary {
		assertRemainingBoundary(t, boundary)
		boundaries[boundary.ID] = true
	}
	for _, boundary := range requiredBoundaries {
		if !boundaries[boundary] {
			t.Fatalf("manifest missing remaining boundary %q", boundary)
		}
	}
}
