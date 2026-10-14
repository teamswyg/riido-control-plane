package main

const (
	defaultManifest = "docs/30-architecture/ai-agent-risk-evidence.riido.json"
	schemaVersion   = "riido-control-plane-ai-agent-risk-evidence.v1"
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
	"web-approval-contract-consumption",
	"web-approval-contract",
	"web-approval-round-trip",
}

var requiredBoundaries = []string{
	"private-repo-auth",
	"client-active-stream-consumption",
}
