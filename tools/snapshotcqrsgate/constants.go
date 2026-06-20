package main

const (
	defaultManifest      = "docs/30-architecture/ai-agent-snapshot-cqrs-gate.riido.json"
	manifestSchema       = "riido-control-plane-ai-agent-snapshot-cqrs-gate.v1"
	evidenceSchema       = "riido-ai-agent-snapshot-cqrs-gate-evidence.v1"
	requiredTask         = "RIID-4964"
	requiredID           = "control-plane-ai-agent-snapshot-cqrs-gate"
	requiredHumanDoc     = "docs/30-architecture/ai-agent-snapshot-cqrs-gate.md"
	workflow             = ".github/workflows/architecture-docs.yml"
	evidenceArtifact     = "snapshot-cqrs-gate-evidence"
	minDecisionThreshold = 50
)

var requiredOperations = []string{
	"ai_agent_client_snapshot_load",
	"ai_agent_client_snapshot_save",
	"store_poll_assignment",
}

var requiredSignals = []string{
	"ai_agent_client_snapshot_load_calls_total",
	"ai_agent_client_snapshot_save_calls_total",
	"ConsumedReadCapacityUnits",
	"ConsumedWriteCapacityUnits",
	"X-Ray",
}
