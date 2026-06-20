package riidoaiserver

const (
	defaultAIAgentClientWorkspaceID = "workspace-dev-riid"
	aiAgentClientReplayEventLimit   = 200
	// Max progress lines persisted per task thread in the snapshot. The live SSE
	// stream is unaffected; this only bounds the replayable/persisted tail so the
	// single snapshot item stays under DynamoDB's 400 KB limit.
	aiAgentClientThreadProgressLineLimit = 200
	// Max threads persisted per task in the snapshot (most recent kept). Bounds
	// snapshot growth from accumulated runs across long-lived tasks.
	aiAgentClientThreadsPerTaskLimit = 50
)
