package requirements

const (
	DefaultManifest      = "docs/30-architecture/ai-agent-snapshot-cqrs-gate.riido.json"
	ManifestSchema       = "riido-control-plane-ai-agent-snapshot-cqrs-gate.v1"
	EvidenceSchema       = "riido-ai-agent-snapshot-cqrs-gate-evidence.v1"
	RequiredTask         = "RIID-4964"
	RequiredID           = "control-plane-ai-agent-snapshot-cqrs-gate"
	RequiredHumanDoc     = "docs/30-architecture/ai-agent-snapshot-cqrs-gate.md"
	Workflow             = ".github/workflows/architecture-docs.yml"
	EvidenceArtifact     = "snapshot-cqrs-gate-evidence"
	MinDecisionThreshold = 50
)

var Operations = []string{
	"ai_agent_client_snapshot_load",
	"ai_agent_client_snapshot_save",
	"store_poll_assignment",
}

var Signals = []string{
	"ai_agent_client_snapshot_load_calls_total",
	"ai_agent_client_snapshot_save_calls_total",
	"ConsumedReadCapacityUnits",
	"ConsumedWriteCapacityUnits",
	"X-Ray",
}
