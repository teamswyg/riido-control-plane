package requirements

const (
	DefaultManifest = "docs/20-domain/ai-agent-client-api.riido.json"
	ManifestSchema  = "riido-ai-agent-client-api.v1"
	EvidenceSchema  = "riido-ai-agent-client-api-evidence.v1"
	ExpectedID      = "ai-agent-client-api"
	ExpectedTask    = "RIID-4721"
)

var RequiredGeneratedPaths = []string{
	"aiAgent.bootstrap",
	"aiAgent.profileThumbnails.uploads.create",
	"aiAgent.tasks.assign",
	"aiAgent.tasks.threadMessages.create",
	"v2.aiAgent.bootstrap",
	"v2.aiAgent.agents.create",
	"v2.aiAgent.profileThumbnails.uploads.create",
	"v2.aiAgent.tasks.assignedAgentProfiles",
	"v2.aiAgent.tasks.agentAssignments.create",
	"v2.aiAgent.tasks.threadStreamSubscription",
}

var RequiredRuntimeConfigKeys = []string{"RIIDO_AI_SERVER_AI_AGENT_CLIENT_DEVELOPMENT"}

var RequiredPublicFields = []string{
	"workspace_id", "profile_thumbnail_url", "provider_version",
	"assigned-agent-profiles", "agent-assignments", "thread-stream-subscription",
	"conversation_id", "parent_thread_id", "agent_snapshot_id", "agent_snapshots",
	"messages", "author_principal_id",
}

var RequiredDeploymentEvidence = []string{
	"not from a manual",
	"The workflow masks both values",
}
